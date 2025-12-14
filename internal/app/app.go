// Package app provides the CLI application layer that wires the SDK to CLI commands.
package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kiracore/sekai-cli/internal/cache"
	"github.com/kiracore/sekai-cli/internal/cli"
	"github.com/kiracore/sekai-cli/internal/config"
	"github.com/kiracore/sekai-cli/internal/output"
	"github.com/kiracore/sekai-cli/pkg/scenarios"
	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/client/docker"
	"github.com/kiracore/sekai-cli/pkg/sdk/client/rest"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/auth"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/bank"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/basket"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/bridge"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/collectives"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/custody"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/distributor"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/gov"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/keys"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/layer2"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/multistaking"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/recovery"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/slashing"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/spending"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/staking"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/status"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/tokens"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/ubi"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/upgrade"
	"github.com/kiracore/sekai-cli/pkg/sdk/types"
)

// App is the CLI application.
type App struct {
	config    *config.Config
	client    sdk.Client
	sdk       *sdk.SEKAI
	root      *cli.Command
	formatter output.Formatter
}

// New creates a new CLI application.
func New(cfg *config.Config) (*App, error) {
	app := &App{
		config:    cfg,
		formatter: output.NewFormatterFromString(cfg.Output),
	}

	// Build root command
	app.root = app.buildRootCommand()

	return app, nil
}

// Run executes the CLI with the given arguments.
func (a *App) Run(args []string) error {
	return a.root.Execute(args)
}

// buildRootCommand builds the root CLI command.
func (a *App) buildRootCommand() *cli.Command {
	root := cli.NewCommand("sekai-cli")
	root.Short = "SEKAI blockchain CLI"
	root.Long = `sekai-cli is a command-line interface for interacting with SEKAI blockchain.

It provides commands for managing keys, sending transactions, and querying
blockchain state. The CLI can communicate with a SEKAI node either through
Docker exec (for local development) or directly via REST API.`

	// Add global flags
	root.AddFlag(cli.Flag{Name: "help", Short: "h", Usage: "Show help"})
	root.AddFlag(cli.Flag{Name: "config", Short: "c", Usage: "Path to config file"})
	root.AddFlag(cli.Flag{Name: "output", Short: "o", Usage: "Output format (text, json, yaml)", Default: "text"})
	root.AddFlag(cli.Flag{Name: "container", Usage: "Docker container name", Default: "sekin-sekai-1"})
	root.AddFlag(cli.Flag{Name: "node", Usage: "Node RPC endpoint", Default: "tcp://localhost:26657"})
	root.AddFlag(cli.Flag{Name: "chain-id", Usage: "Chain ID"})
	root.AddFlag(cli.Flag{Name: "keyring-backend", Usage: "Keyring backend", Default: "test"})
	root.AddFlag(cli.Flag{Name: "home", Usage: "Sekaid home directory", Default: "/sekai"})
	root.AddFlag(cli.Flag{Name: "rest", Usage: "REST API endpoint (enables REST mode)"})

	// Add subcommands
	root.AddCommand(a.buildInitCommand())
	root.AddCommand(a.buildSyncCommand())
	root.AddCommand(a.buildCacheCommand())
	root.AddCommand(a.buildStatusCommand())
	root.AddCommand(a.buildKeysCommand())
	root.AddCommand(a.buildBankCommand())
	root.AddCommand(a.buildQueryCommand())
	root.AddCommand(a.buildTxCommand())
	root.AddCommand(a.buildVersionCommand())
	root.AddCommand(a.buildConfigCommand())
	root.AddCommand(a.buildScenarioCommand())
	root.AddCommand(a.buildCompletionCommand())

	return root
}

// getClient creates or returns the SDK client based on context flags.
// Priority order: flags > cache > config
func (a *App) getClient(ctx *cli.Context) (sdk.Client, error) {
	if a.client != nil {
		return a.client, nil
	}

	// Try to load cache for defaults (cache takes priority over config)
	cachedData := cache.TryLoad()

	// Helper to get value with priority: flag > cache > config
	getValueWithCache := func(flagVal, cacheVal, configVal string) string {
		if flagVal != "" {
			return flagVal
		}
		if cacheVal != "" {
			return cacheVal
		}
		return configVal
	}

	// Get chain ID: flag > cache > config
	var cacheChainID string
	if cachedData != nil {
		cacheChainID = cachedData.GetChainID()
	}
	chainID := getValueWithCache(ctx.GetFlag("chain-id"), cacheChainID, a.config.ChainID)

	// Check if REST mode is enabled
	restURL := ctx.GetFlag("rest")
	if restURL == "" {
		restURL = a.config.RESTURL
	}

	// If --rest flag is provided or UseREST is configured, use REST client
	if restURL != "" && (ctx.GetFlag("rest") != "" || a.config.UseREST) {
		client, err := rest.NewClient(restURL,
			rest.WithChainID(chainID),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create REST client: %w", err)
		}
		a.client = client
		return client, nil
	}

	// Get container: flag > cache > config > auto-detect
	var cacheContainer string
	if cachedData != nil {
		cacheContainer = cachedData.GetContainer()
	}
	container := getValueWithCache(ctx.GetFlag("container"), cacheContainer, a.config.Container)

	if container == "" {
		// Try to auto-detect
		detected, err := docker.FindSekaiContainer()
		if err != nil {
			return nil, fmt.Errorf("no container specified and auto-detection failed: %w\nRun 'sekai-cli init' to configure", err)
		}
		container = detected
	}

	// Get fees: flag > cache > config
	var cacheFees string
	if cachedData != nil {
		cacheFees = cachedData.GetMinFee()
		if cacheFees != "" {
			cacheFees = cacheFees + "ukex"
		}
	}
	fees := getValueWithCache(ctx.GetFlag("fees"), cacheFees, a.config.Fees)

	// Build options
	opts := []docker.Option{
		docker.WithChainID(chainID),
		docker.WithKeyringBackend(getStringOrDefault(ctx.GetFlag("keyring-backend"), a.config.KeyringBackend)),
		docker.WithHome(getStringOrDefault(ctx.GetFlag("home"), a.config.Home)),
		docker.WithNode(getStringOrDefault(ctx.GetFlag("node"), a.config.Node)),
		docker.WithFees(fees),
		docker.WithGas(a.config.Gas),
		docker.WithGasAdjustment(a.config.GasAdjustment),
	}

	client, err := docker.NewClient(container, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	a.client = client
	return client, nil
}

// getDefaultFrom returns the default signer key from cache.
func (a *App) getDefaultFrom() string {
	if cachedData := cache.TryLoad(); cachedData != nil {
		return cachedData.GetDefaultKey()
	}
	return ""
}

// getFromFlag returns the --from value, falling back to cache default.
func (a *App) getFromFlag(ctx *cli.Context) string {
	from := ctx.GetFlag("from")
	if from == "" {
		from = a.getDefaultFrom()
	}
	return from
}

// getSDK creates or returns the SDK instance.
func (a *App) getSDK(ctx *cli.Context) (*sdk.SEKAI, error) {
	if a.sdk != nil {
		return a.sdk, nil
	}

	client, err := a.getClient(ctx)
	if err != nil {
		return nil, err
	}

	a.sdk = sdk.New(client)
	return a.sdk, nil
}

// getFormatter returns the output formatter based on context.
func (a *App) getFormatter(ctx *cli.Context) output.Formatter {
	format := ctx.GetFlag("output")
	if format == "" {
		format = a.config.Output
	}
	return output.NewFormatterFromString(format)
}

// printOutput prints data using the configured formatter.
func (a *App) printOutput(ctx *cli.Context, data interface{}) error {
	formatter := a.getFormatter(ctx)
	return formatter.Format(ctx.Stdout, data)
}

// buildStatusCommand builds the status command.
func (a *App) buildStatusCommand() *cli.Command {
	cmd := cli.NewCommand("status")
	cmd.Short = "Get node status"
	cmd.Long = "Query the status of the connected SEKAI node."

	cmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}

		statusMod := status.New(client)
		resp, err := statusMod.Status(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		return a.printOutput(ctx, resp)
	}

	return cmd
}

// buildKeysCommand builds the keys command group.
func (a *App) buildKeysCommand() *cli.Command {
	keysCmd := cli.NewCommand("keys")
	keysCmd.Short = "Key management commands"
	keysCmd.Long = "Manage keys in the keyring."

	// keys list
	listCmd := cli.NewCommand("list")
	listCmd.Aliases = []string{"ls"}
	listCmd.Short = "List all keys"
	listCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		keysMod := keys.New(client)
		keysList, err := keysMod.List(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, keysList)
	}
	keysCmd.AddCommand(listCmd)

	// keys show
	showCmd := cli.NewCommand("show")
	showCmd.Short = "Show key details"
	showCmd.Args = []cli.Arg{{Name: "name", Required: true, Description: "Key name"}}
	showCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("key name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		keysMod := keys.New(client)
		info, err := keysMod.Show(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, info)
	}
	keysCmd.AddCommand(showCmd)

	// keys add
	addCmd := cli.NewCommand("add")
	addCmd.Short = "Create a new key"
	addCmd.Args = []cli.Arg{{Name: "name", Required: true, Description: "Key name"}}
	addCmd.AddFlag(cli.Flag{Name: "recover", Usage: "Recover from mnemonic"})
	addCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("key name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		keysMod := keys.New(client)
		opts := &sdk.KeyAddOptions{
			Recover: ctx.GetFlag("recover") == "true",
		}
		info, err := keysMod.Add(context.Background(), ctx.Args[0], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, info)
	}
	keysCmd.AddCommand(addCmd)

	// keys delete
	deleteCmd := cli.NewCommand("delete")
	deleteCmd.Aliases = []string{"rm"}
	deleteCmd.Short = "Delete a key"
	deleteCmd.Args = []cli.Arg{{Name: "name", Required: true, Description: "Key name"}}
	deleteCmd.AddFlag(cli.Flag{Name: "force", Short: "f", Usage: "Force deletion"})
	deleteCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("key name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		keysMod := keys.New(client)
		force := ctx.GetFlag("force") == "true"
		if err := keysMod.Delete(context.Background(), ctx.Args[0], force); err != nil {
			return err
		}
		ctx.Println("Key deleted successfully")
		return nil
	}
	keysCmd.AddCommand(deleteCmd)

	return keysCmd
}

// buildBankCommand builds the bank command group.
func (a *App) buildBankCommand() *cli.Command {
	bankCmd := cli.NewCommand("bank")
	bankCmd.Short = "Bank module commands"
	bankCmd.Long = "Query balances and send tokens."

	// bank balances
	balancesCmd := cli.NewCommand("balances")
	balancesCmd.Short = "Query account balances"
	balancesCmd.Args = []cli.Arg{{Name: "address", Required: true, Description: "Account address"}}
	balancesCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)
		balances, err := bankMod.Balances(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, balances)
	}
	bankCmd.AddCommand(balancesCmd)

	// bank send
	sendCmd := cli.NewCommand("send")
	sendCmd.Short = "Send tokens"
	sendCmd.Args = []cli.Arg{
		{Name: "from", Required: true, Description: "Sender key name"},
		{Name: "to", Required: true, Description: "Recipient address"},
		{Name: "amount", Required: true, Description: "Amount to send (e.g., 100ukex)"},
	}
	cli.AddTxFlags(sendCmd)
	sendCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 3 {
			return fmt.Errorf("from, to, and amount required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)

		coins, err := types.ParseCoins(ctx.Args[2])
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		opts := &bank.SendOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}

		resp, err := bankMod.Send(context.Background(), ctx.Args[0], ctx.Args[1], coins, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	bankCmd.AddCommand(sendCmd)

	return bankCmd
}

// buildQueryCommand builds the query command group.
func (a *App) buildQueryCommand() *cli.Command {
	queryCmd := cli.NewCommand("query")
	queryCmd.Aliases = []string{"q"}
	queryCmd.Short = "Query commands"
	queryCmd.Long = "Query blockchain state."

	queryCmd.AddCommand(a.buildQueryAuthCommand())
	queryCmd.AddCommand(a.buildQueryBankCommand())
	queryCmd.AddCommand(a.buildQueryCustomgovCommand())
	queryCmd.AddCommand(a.buildQueryCustomstakingCommand())
	queryCmd.AddCommand(a.buildQueryTokensCommand())
	queryCmd.AddCommand(a.buildQueryMultistakingCommand())
	queryCmd.AddCommand(a.buildQuerySpendingCommand())
	queryCmd.AddCommand(a.buildQueryUBICommand())
	queryCmd.AddCommand(a.buildQueryUpgradeCommand())
	queryCmd.AddCommand(a.buildQuerySlashingCommand())
	queryCmd.AddCommand(a.buildQueryDistributorCommand())
	queryCmd.AddCommand(a.buildQueryBasketCommand())
	queryCmd.AddCommand(a.buildQueryCollectivesCommand())
	queryCmd.AddCommand(a.buildQueryCustodyCommand())
	queryCmd.AddCommand(a.buildQueryBridgeCommand())
	queryCmd.AddCommand(a.buildQueryLayer2Command())
	queryCmd.AddCommand(a.buildQueryRecoveryCommand())

	return queryCmd
}

// buildQueryAuthCommand builds the query auth command group.
func (a *App) buildQueryAuthCommand() *cli.Command {
	authQuery := cli.NewCommand("auth")
	authQuery.Short = "Auth query commands"

	// query auth account
	accountCmd := cli.NewCommand("account")
	accountCmd.Short = "Query account by address"
	accountCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	accountCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		authMod := auth.New(client)
		account, err := authMod.Account(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, account)
	}
	authQuery.AddCommand(accountCmd)

	// query auth accounts
	accountsCmd := cli.NewCommand("accounts")
	accountsCmd.Short = "Query all accounts"
	accountsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		authMod := auth.New(client)
		accounts, err := authMod.Accounts(context.Background(), nil)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, accounts)
	}
	authQuery.AddCommand(accountsCmd)

	// query auth module-account
	moduleAccCmd := cli.NewCommand("module-account")
	moduleAccCmd.Short = "Query module account by name"
	moduleAccCmd.Args = []cli.Arg{{Name: "name", Required: true}}
	moduleAccCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("module name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		authMod := auth.New(client)
		account, err := authMod.ModuleAccount(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, account)
	}
	authQuery.AddCommand(moduleAccCmd)

	// query auth module-accounts
	moduleAccsCmd := cli.NewCommand("module-accounts")
	moduleAccsCmd.Short = "Query all module accounts"
	moduleAccsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		authMod := auth.New(client)
		accounts, err := authMod.ModuleAccounts(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, accounts)
	}
	authQuery.AddCommand(moduleAccsCmd)

	// query auth params
	paramsCmd := cli.NewCommand("params")
	paramsCmd.Short = "Query auth parameters"
	paramsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		authMod := auth.New(client)
		params, err := authMod.Params(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, params)
	}
	authQuery.AddCommand(paramsCmd)

	// query auth address-by-acc-num
	addrByAccNumCmd := cli.NewCommand("address-by-acc-num")
	addrByAccNumCmd.Short = "Query address by account number"
	addrByAccNumCmd.Aliases = []string{"address-by-id"}
	addrByAccNumCmd.Args = []cli.Arg{{Name: "acc-num", Required: true, Description: "Account number"}}
	addrByAccNumCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("account number required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		authMod := auth.New(client)
		result, err := authMod.AddressByAccNum(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	authQuery.AddCommand(addrByAccNumCmd)

	return authQuery
}

// buildQueryBankCommand builds the query bank command group.
func (a *App) buildQueryBankCommand() *cli.Command {
	bankQuery := cli.NewCommand("bank")
	bankQuery.Short = "Bank query commands"

	balancesCmd := cli.NewCommand("balances")
	balancesCmd.Short = "Query account balances"
	balancesCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	balancesCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)
		balances, err := bankMod.Balances(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, balances)
	}
	bankQuery.AddCommand(balancesCmd)

	totalCmd := cli.NewCommand("total")
	totalCmd.Short = "Query total supply"
	totalCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)
		supply, err := bankMod.TotalSupply(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, supply)
	}
	bankQuery.AddCommand(totalCmd)

	// spendable-balances
	spendableCmd := cli.NewCommand("spendable-balances")
	spendableCmd.Short = "Query spendable balances by address"
	spendableCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	spendableCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)
		balances, err := bankMod.SpendableBalances(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, balances)
	}
	bankQuery.AddCommand(spendableCmd)

	// denom-metadata
	denomMetaCmd := cli.NewCommand("denom-metadata")
	denomMetaCmd.Short = "Query denom metadata"
	denomMetaCmd.AddFlag(cli.Flag{Name: "denom", Usage: "Denom to query metadata for"})
	denomMetaCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)
		denom := ctx.GetFlag("denom")
		result, err := bankMod.DenomMetadata(context.Background(), denom)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	bankQuery.AddCommand(denomMetaCmd)

	// send-enabled
	sendEnabledCmd := cli.NewCommand("send-enabled")
	sendEnabledCmd.Short = "Query send enabled entries"
	sendEnabledCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)
		// Pass any denom args
		result, err := bankMod.SendEnabled(context.Background(), ctx.Args...)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	bankQuery.AddCommand(sendEnabledCmd)

	return bankQuery
}

// buildQueryCustomgovCommand builds the query customgov command group.
func (a *App) buildQueryCustomgovCommand() *cli.Command {
	govQuery := cli.NewCommand("customgov")
	govQuery.Short = "Governance query commands"

	// network-properties
	netPropsCmd := cli.NewCommand("network-properties")
	netPropsCmd.Short = "Query network properties"
	netPropsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		props, err := govMod.NetworkProperties(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, props)
	}
	govQuery.AddCommand(netPropsCmd)

	// proposals
	proposalsCmd := cli.NewCommand("proposals")
	proposalsCmd.Short = "Query proposals"
	proposalsCmd.AddFlag(cli.Flag{Name: "voter", Usage: "Filter by voter"})
	proposalsCmd.AddFlag(cli.Flag{Name: "status", Usage: "Filter by status"})
	proposalsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		opts := &gov.ProposalQueryOpts{
			Voter:  ctx.GetFlag("voter"),
			Status: ctx.GetFlag("status"),
		}
		proposals, err := govMod.Proposals(context.Background(), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, proposals)
	}
	govQuery.AddCommand(proposalsCmd)

	// proposal
	proposalCmd := cli.NewCommand("proposal")
	proposalCmd.Short = "Query proposal by ID"
	proposalCmd.Args = []cli.Arg{{Name: "proposal-id", Required: true}}
	proposalCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("proposal ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		proposal, err := govMod.Proposal(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, proposal)
	}
	govQuery.AddCommand(proposalCmd)

	// votes
	votesCmd := cli.NewCommand("votes")
	votesCmd.Short = "Query votes on a proposal"
	votesCmd.Args = []cli.Arg{{Name: "proposal-id", Required: true}}
	votesCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("proposal ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		votes, err := govMod.Votes(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, votes)
	}
	govQuery.AddCommand(votesCmd)

	// councilors
	councilorsCmd := cli.NewCommand("councilors")
	councilorsCmd.Short = "Query councilors"
	councilorsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		councilors, err := govMod.Councilors(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, councilors)
	}
	govQuery.AddCommand(councilorsCmd)

	// all-roles
	allRolesCmd := cli.NewCommand("all-roles")
	allRolesCmd.Short = "Query all roles"
	allRolesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		roles, err := govMod.AllRoles(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, roles)
	}
	govQuery.AddCommand(allRolesCmd)

	// role
	roleCmd := cli.NewCommand("role")
	roleCmd.Short = "Query role by ID or SID"
	roleCmd.Args = []cli.Arg{{Name: "identifier", Required: true}}
	roleCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("role ID or SID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		role, err := govMod.Role(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, role)
	}
	govQuery.AddCommand(roleCmd)

	// roles (by address)
	rolesCmd := cli.NewCommand("roles")
	rolesCmd.Short = "Query roles assigned to an address"
	rolesCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	rolesCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		roles, err := govMod.Roles(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, roles)
	}
	govQuery.AddCommand(rolesCmd)

	// permissions
	permissionsCmd := cli.NewCommand("permissions")
	permissionsCmd.Short = "Query permissions of an address"
	permissionsCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	permissionsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		perms, err := govMod.Permissions(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, perms)
	}
	govQuery.AddCommand(permissionsCmd)

	// all-execution-fees
	allFeesCmd := cli.NewCommand("all-execution-fees")
	allFeesCmd.Short = "Query all execution fees"
	allFeesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		fees, err := govMod.AllExecutionFees(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, fees)
	}
	govQuery.AddCommand(allFeesCmd)

	// execution-fee
	feeCmd := cli.NewCommand("execution-fee")
	feeCmd.Short = "Query execution fee by tx type"
	feeCmd.Args = []cli.Arg{{Name: "tx-type", Required: true}}
	feeCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("transaction type required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		fee, err := govMod.ExecutionFee(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, fee)
	}
	govQuery.AddCommand(feeCmd)

	// identity-records
	idRecordsCmd := cli.NewCommand("identity-records")
	idRecordsCmd.Short = "Query all identity records"
	idRecordsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		records, err := govMod.IdentityRecords(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, records)
	}
	govQuery.AddCommand(idRecordsCmd)

	// identity-record
	idRecordCmd := cli.NewCommand("identity-record")
	idRecordCmd.Short = "Query identity record by ID"
	idRecordCmd.Args = []cli.Arg{{Name: "id", Required: true}}
	idRecordCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("record ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		record, err := govMod.IdentityRecord(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, record)
	}
	govQuery.AddCommand(idRecordCmd)

	// identity-records-by-addr
	idRecordsByAddrCmd := cli.NewCommand("identity-records-by-addr")
	idRecordsByAddrCmd.Short = "Query identity records by address"
	idRecordsByAddrCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	idRecordsByAddrCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		records, err := govMod.IdentityRecordsByAddress(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, records)
	}
	govQuery.AddCommand(idRecordsByAddrCmd)

	// data-registry-keys
	dataKeysCmd := cli.NewCommand("data-registry-keys")
	dataKeysCmd.Short = "Query all data registry keys"
	dataKeysCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		keys, err := govMod.DataRegistryKeys(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, keys)
	}
	govQuery.AddCommand(dataKeysCmd)

	// data-registry
	dataRegCmd := cli.NewCommand("data-registry")
	dataRegCmd.Short = "Query data registry by key"
	dataRegCmd.Args = []cli.Arg{{Name: "key", Required: true}}
	dataRegCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("key required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		entry, err := govMod.DataRegistry(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, entry)
	}
	govQuery.AddCommand(dataRegCmd)

	// poor-network-messages
	poorNetCmd := cli.NewCommand("poor-network-messages")
	poorNetCmd.Short = "Query poor network messages"
	poorNetCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		msgs, err := govMod.PoorNetworkMessages(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, msgs)
	}
	govQuery.AddCommand(poorNetCmd)

	// custom-prefixes
	customPrefixesCmd := cli.NewCommand("custom-prefixes")
	customPrefixesCmd.Short = "Query custom prefixes"
	customPrefixesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		prefixes, err := govMod.CustomPrefixes(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, prefixes)
	}
	govQuery.AddCommand(customPrefixesCmd)

	// all-proposal-durations
	allProposalDurationsCmd := cli.NewCommand("all-proposal-durations")
	allProposalDurationsCmd.Short = "Query all proposal durations"
	allProposalDurationsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		durations, err := govMod.AllProposalDurations(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, durations)
	}
	govQuery.AddCommand(allProposalDurationsCmd)

	// proposal-duration
	proposalDurationCmd := cli.NewCommand("proposal-duration")
	proposalDurationCmd.Short = "Query proposal duration by type"
	proposalDurationCmd.Args = []cli.Arg{{Name: "proposal-type", Required: true}}
	proposalDurationCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("proposal type required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		duration, err := govMod.ProposalDuration(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, map[string]string{"duration": duration})
	}
	govQuery.AddCommand(proposalDurationCmd)

	// non-councilors
	nonCouncilorsCmd := cli.NewCommand("non-councilors")
	nonCouncilorsCmd.Short = "Query non-councilors"
	nonCouncilorsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		members, err := govMod.NonCouncilors(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, members)
	}
	govQuery.AddCommand(nonCouncilorsCmd)

	// council-registry
	councilRegistryCmd := cli.NewCommand("council-registry")
	councilRegistryCmd.Short = "Query council registry"
	councilRegistryCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		registry, err := govMod.CouncilRegistry(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, registry)
	}
	govQuery.AddCommand(councilRegistryCmd)

	// whitelisted-permission-addresses
	whitelistedPermAddrsCmd := cli.NewCommand("whitelisted-permission-addresses")
	whitelistedPermAddrsCmd.Short = "Query addresses with whitelisted permission"
	whitelistedPermAddrsCmd.Args = []cli.Arg{{Name: "permission", Required: true}}
	whitelistedPermAddrsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("permission ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		addrs, err := govMod.WhitelistedPermissionAddresses(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, addrs)
	}
	govQuery.AddCommand(whitelistedPermAddrsCmd)

	// blacklisted-permission-addresses
	blacklistedPermAddrsCmd := cli.NewCommand("blacklisted-permission-addresses")
	blacklistedPermAddrsCmd.Short = "Query addresses with blacklisted permission"
	blacklistedPermAddrsCmd.Args = []cli.Arg{{Name: "permission", Required: true}}
	blacklistedPermAddrsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("permission ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		addrs, err := govMod.BlacklistedPermissionAddresses(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, addrs)
	}
	govQuery.AddCommand(blacklistedPermAddrsCmd)

	// whitelisted-role-addresses
	whitelistedRoleAddrsCmd := cli.NewCommand("whitelisted-role-addresses")
	whitelistedRoleAddrsCmd.Short = "Query addresses with whitelisted role"
	whitelistedRoleAddrsCmd.Args = []cli.Arg{{Name: "role", Required: true}}
	whitelistedRoleAddrsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("role ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		addrs, err := govMod.WhitelistedRoleAddresses(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, addrs)
	}
	govQuery.AddCommand(whitelistedRoleAddrsCmd)

	// proposer-voters-count
	proposerVotersCountCmd := cli.NewCommand("proposer-voters-count")
	proposerVotersCountCmd.Short = "Query proposer and voters count"
	proposerVotersCountCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		count, err := govMod.ProposerVotersCount(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, count)
	}
	govQuery.AddCommand(proposerVotersCountCmd)

	// all-identity-record-verify-requests
	allIdVerifyReqsCmd := cli.NewCommand("all-identity-record-verify-requests")
	allIdVerifyReqsCmd.Short = "Query all identity record verify requests"
	allIdVerifyReqsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		reqs, err := govMod.AllIdentityRecordVerifyRequests(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, reqs)
	}
	govQuery.AddCommand(allIdVerifyReqsCmd)

	// identity-record-verify-request
	idVerifyReqCmd := cli.NewCommand("identity-record-verify-request")
	idVerifyReqCmd.Short = "Query identity record verify request by ID"
	idVerifyReqCmd.Args = []cli.Arg{{Name: "id", Required: true}}
	idVerifyReqCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("request ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		req, err := govMod.IdentityRecordVerifyRequest(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, req)
	}
	govQuery.AddCommand(idVerifyReqCmd)

	// identity-record-verify-requests-by-approver
	idVerifyReqsByApproverCmd := cli.NewCommand("identity-record-verify-requests-by-approver")
	idVerifyReqsByApproverCmd.Short = "Query identity record verify requests by approver"
	idVerifyReqsByApproverCmd.Args = []cli.Arg{{Name: "approver", Required: true}}
	idVerifyReqsByApproverCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("approver address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		reqs, err := govMod.IdentityRecordVerifyRequestsByApprover(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, reqs)
	}
	govQuery.AddCommand(idVerifyReqsByApproverCmd)

	// identity-record-verify-requests-by-requester
	idVerifyReqsByRequesterCmd := cli.NewCommand("identity-record-verify-requests-by-requester")
	idVerifyReqsByRequesterCmd.Short = "Query identity record verify requests by requester"
	idVerifyReqsByRequesterCmd.Args = []cli.Arg{{Name: "requester", Required: true}}
	idVerifyReqsByRequesterCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("requester address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		reqs, err := govMod.IdentityRecordVerifyRequestsByRequester(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, reqs)
	}
	govQuery.AddCommand(idVerifyReqsByRequesterCmd)

	// polls
	pollsCmd := cli.NewCommand("polls")
	pollsCmd.Short = "Query polls by address"
	pollsCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	pollsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		polls, err := govMod.Polls(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, polls)
	}
	govQuery.AddCommand(pollsCmd)

	// poll-votes
	pollVotesCmd := cli.NewCommand("poll-votes")
	pollVotesCmd.Short = "Query poll votes by ID"
	pollVotesCmd.Args = []cli.Arg{{Name: "poll-id", Required: true}}
	pollVotesCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("poll ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		votes, err := govMod.PollVotes(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, votes)
	}
	govQuery.AddCommand(pollVotesCmd)

	// vote
	voteCmd := cli.NewCommand("vote")
	voteCmd.Short = "Query vote on a proposal"
	voteCmd.Args = []cli.Arg{{Name: "proposal-id", Required: true}, {Name: "voter", Required: true}}
	voteCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("proposal ID and voter address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		vote, err := govMod.Vote(context.Background(), ctx.Args[0], ctx.Args[1])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, vote)
	}
	govQuery.AddCommand(voteCmd)

	// voters
	votersCmd := cli.NewCommand("voters")
	votersCmd.Short = "Query voters of a proposal"
	votersCmd.Args = []cli.Arg{{Name: "proposal-id", Required: true}}
	votersCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("proposal ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		voters, err := govMod.Voters(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, voters)
	}
	govQuery.AddCommand(votersCmd)

	return govQuery
}

// buildQueryCustomstakingCommand builds the query customstaking command group.
func (a *App) buildQueryCustomstakingCommand() *cli.Command {
	stakingQuery := cli.NewCommand("customstaking")
	stakingQuery.Short = "Staking query commands"

	// validators
	validatorsCmd := cli.NewCommand("validators")
	validatorsCmd.Short = "Query validators"
	validatorsCmd.AddFlag(cli.Flag{Name: "addr", Usage: "Filter by address"})
	validatorsCmd.AddFlag(cli.Flag{Name: "val-addr", Usage: "Filter by validator address"})
	validatorsCmd.AddFlag(cli.Flag{Name: "moniker", Usage: "Filter by moniker"})
	validatorsCmd.AddFlag(cli.Flag{Name: "status", Usage: "Filter by status"})
	validatorsCmd.AddFlag(cli.Flag{Name: "pubkey", Usage: "Filter by pubkey"})
	validatorsCmd.AddFlag(cli.Flag{Name: "proposer", Usage: "Filter by proposer"})
	validatorsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		stakingMod := staking.New(client)
		opts := &staking.ValidatorQueryOpts{
			Address:  ctx.GetFlag("addr"),
			ValAddr:  ctx.GetFlag("val-addr"),
			Moniker:  ctx.GetFlag("moniker"),
			Status:   ctx.GetFlag("status"),
			PubKey:   ctx.GetFlag("pubkey"),
			Proposer: ctx.GetFlag("proposer"),
		}
		validators, err := stakingMod.Validators(context.Background(), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, validators)
	}
	stakingQuery.AddCommand(validatorsCmd)

	// validator
	validatorCmd := cli.NewCommand("validator")
	validatorCmd.Short = "Query a validator by address, val-address, or moniker"
	validatorCmd.AddFlag(cli.Flag{Name: "addr", Usage: "Query by address"})
	validatorCmd.AddFlag(cli.Flag{Name: "val-addr", Usage: "Query by validator address"})
	validatorCmd.AddFlag(cli.Flag{Name: "moniker", Usage: "Query by moniker"})
	validatorCmd.Run = func(ctx *cli.Context) error {
		addr := ctx.GetFlag("addr")
		valAddr := ctx.GetFlag("val-addr")
		moniker := ctx.GetFlag("moniker")
		if addr == "" && valAddr == "" && moniker == "" {
			return fmt.Errorf("at least one of --addr, --val-addr, or --moniker required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		stakingMod := staking.New(client)
		validator, err := stakingMod.Validator(context.Background(), &staking.ValidatorQueryOpts{
			Address: addr,
			ValAddr: valAddr,
			Moniker: moniker,
		})
		if err != nil {
			return err
		}
		return a.printOutput(ctx, validator)
	}
	stakingQuery.AddCommand(validatorCmd)

	return stakingQuery
}

// buildQueryTokensCommand builds the query tokens command group.
func (a *App) buildQueryTokensCommand() *cli.Command {
	tokensQuery := cli.NewCommand("tokens")
	tokensQuery.Short = "Tokens query commands"

	// all-rates
	allRatesCmd := cli.NewCommand("all-rates")
	allRatesCmd.Short = "Query all token rates"
	allRatesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		rates, err := tokensMod.AllRates(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, rates)
	}
	tokensQuery.AddCommand(allRatesCmd)

	// rate
	rateCmd := cli.NewCommand("rate")
	rateCmd.Short = "Query token rate by denom"
	rateCmd.Args = []cli.Arg{{Name: "denom", Required: true}}
	rateCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("denom required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		rate, err := tokensMod.Rate(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, rate)
	}
	tokensQuery.AddCommand(rateCmd)

	// rates-by-denom
	ratesByDenomCmd := cli.NewCommand("rates-by-denom")
	ratesByDenomCmd.Short = "Query token rates by denom"
	ratesByDenomCmd.Args = []cli.Arg{{Name: "denom", Required: true}}
	ratesByDenomCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("denom required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		rates, err := tokensMod.RatesByDenom(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, rates)
	}
	tokensQuery.AddCommand(ratesByDenomCmd)

	// token-black-whites
	blackWhitesCmd := cli.NewCommand("token-black-whites")
	blackWhitesCmd.Short = "Query token black and white lists"
	blackWhitesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		lists, err := tokensMod.TokenBlackWhites(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, lists)
	}
	tokensQuery.AddCommand(blackWhitesCmd)

	return tokensQuery
}

// buildQueryMultistakingCommand builds the query multistaking command group.
func (a *App) buildQueryMultistakingCommand() *cli.Command {
	multistakingQuery := cli.NewCommand("multistaking")
	multistakingQuery.Short = "Multistaking query commands"

	// pools
	poolsCmd := cli.NewCommand("pools")
	poolsCmd.Short = "Query all staking pools"
	poolsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		pools, err := msMod.Pools(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, pools)
	}
	multistakingQuery.AddCommand(poolsCmd)

	// undelegations
	undelegationsCmd := cli.NewCommand("undelegations")
	undelegationsCmd.Short = "Query all undelegations for a delegator and validator"
	undelegationsCmd.Args = []cli.Arg{
		{Name: "delegator", Required: true},
		{Name: "val-addr", Required: true},
	}
	undelegationsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("delegator and validator address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		undelegations, err := msMod.Undelegations(context.Background(), ctx.Args[0], ctx.Args[1])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, undelegations)
	}
	multistakingQuery.AddCommand(undelegationsCmd)

	// outstanding-rewards
	rewardsCmd := cli.NewCommand("outstanding-rewards")
	rewardsCmd.Short = "Query outstanding rewards for a delegator"
	rewardsCmd.Args = []cli.Arg{{Name: "delegator", Required: true}}
	rewardsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("delegator address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		rewards, err := msMod.OutstandingRewards(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, rewards)
	}
	multistakingQuery.AddCommand(rewardsCmd)

	// compound-info
	compoundCmd := cli.NewCommand("compound-info")
	compoundCmd.Short = "Query compound information of a delegator"
	compoundCmd.Args = []cli.Arg{{Name: "delegator", Required: true}}
	compoundCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("delegator address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		info, err := msMod.CompoundInfo(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, info)
	}
	multistakingQuery.AddCommand(compoundCmd)

	// staking-pool-delegators
	delegatorsCmd := cli.NewCommand("staking-pool-delegators")
	delegatorsCmd.Short = "Query staking pool delegators for a validator"
	delegatorsCmd.Args = []cli.Arg{{Name: "validator", Required: true}}
	delegatorsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("validator address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		delegators, err := msMod.StakingPoolDelegators(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, delegators)
	}
	multistakingQuery.AddCommand(delegatorsCmd)

	return multistakingQuery
}

// buildQuerySpendingCommand builds the query spending command group.
func (a *App) buildQuerySpendingCommand() *cli.Command {
	spendingQuery := cli.NewCommand("spending")
	spendingQuery.Short = "Spending query commands"

	// pool-names
	poolNamesCmd := cli.NewCommand("pool-names")
	poolNamesCmd.Short = "Query all pool names"
	poolNamesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		result, err := spendingMod.PoolNames(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	spendingQuery.AddCommand(poolNamesCmd)

	// pool-by-name
	poolByNameCmd := cli.NewCommand("pool-by-name")
	poolByNameCmd.Short = "Query pool by name"
	poolByNameCmd.Args = []cli.Arg{{Name: "name", Required: true}}
	poolByNameCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("pool name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		result, err := spendingMod.PoolByName(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	spendingQuery.AddCommand(poolByNameCmd)

	// pool-proposals
	poolProposalsCmd := cli.NewCommand("pool-proposals")
	poolProposalsCmd.Short = "Query proposals for a pool"
	poolProposalsCmd.Args = []cli.Arg{{Name: "pool-name", Required: true}}
	poolProposalsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("pool name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		result, err := spendingMod.PoolProposals(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	spendingQuery.AddCommand(poolProposalsCmd)

	// pools-by-account
	poolsByAccountCmd := cli.NewCommand("pools-by-account")
	poolsByAccountCmd.Short = "Query pools by account"
	poolsByAccountCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	poolsByAccountCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		result, err := spendingMod.PoolsByAccount(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	spendingQuery.AddCommand(poolsByAccountCmd)

	return spendingQuery
}

// buildQueryUBICommand builds the query ubi command group.
func (a *App) buildQueryUBICommand() *cli.Command {
	ubiQuery := cli.NewCommand("ubi")
	ubiQuery.Short = "UBI query commands"

	// ubi-records
	recordsCmd := cli.NewCommand("ubi-records")
	recordsCmd.Short = "Query all UBI records"
	recordsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		ubiMod := ubi.New(client)
		result, err := ubiMod.Records(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	ubiQuery.AddCommand(recordsCmd)

	// ubi-record-by-name
	recordByNameCmd := cli.NewCommand("ubi-record-by-name")
	recordByNameCmd.Short = "Query UBI record by name"
	recordByNameCmd.Args = []cli.Arg{{Name: "name", Required: true}}
	recordByNameCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		ubiMod := ubi.New(client)
		result, err := ubiMod.RecordByName(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	ubiQuery.AddCommand(recordByNameCmd)

	return ubiQuery
}

// buildQueryUpgradeCommand builds the query upgrade command group.
func (a *App) buildQueryUpgradeCommand() *cli.Command {
	upgradeQuery := cli.NewCommand("upgrade")
	upgradeQuery.Short = "Upgrade query commands"

	// current-plan
	currentPlanCmd := cli.NewCommand("current-plan")
	currentPlanCmd.Short = "Query current upgrade plan"
	currentPlanCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		upgradeMod := upgrade.New(client)
		result, err := upgradeMod.CurrentPlan(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	upgradeQuery.AddCommand(currentPlanCmd)

	// next-plan
	nextPlanCmd := cli.NewCommand("next-plan")
	nextPlanCmd.Short = "Query next upgrade plan"
	nextPlanCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		upgradeMod := upgrade.New(client)
		result, err := upgradeMod.NextPlan(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	upgradeQuery.AddCommand(nextPlanCmd)

	return upgradeQuery
}

// buildQuerySlashingCommand builds the query customslashing command group.
func (a *App) buildQuerySlashingCommand() *cli.Command {
	slashingQuery := cli.NewCommand("customslashing")
	slashingQuery.Short = "Slashing query commands"

	// signing-info
	signingInfoCmd := cli.NewCommand("signing-info")
	signingInfoCmd.Short = "Query validator signing info"
	signingInfoCmd.Args = []cli.Arg{{Name: "cons-address", Required: true}}
	signingInfoCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("consensus address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		slashingMod := slashing.New(client)
		result, err := slashingMod.SigningInfo(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	slashingQuery.AddCommand(signingInfoCmd)

	// signing-infos
	signingInfosCmd := cli.NewCommand("signing-infos")
	signingInfosCmd.Short = "Query all validator signing infos"
	signingInfosCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		slashingMod := slashing.New(client)
		result, err := slashingMod.SigningInfos(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	slashingQuery.AddCommand(signingInfosCmd)

	// active-staking-pools
	activePoolsCmd := cli.NewCommand("active-staking-pools")
	activePoolsCmd.Short = "Query active staking pools"
	activePoolsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		slashingMod := slashing.New(client)
		result, err := slashingMod.ActiveStakingPools(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	slashingQuery.AddCommand(activePoolsCmd)

	// inactive-staking-pools
	inactivePoolsCmd := cli.NewCommand("inactive-staking-pools")
	inactivePoolsCmd.Short = "Query inactive staking pools"
	inactivePoolsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		slashingMod := slashing.New(client)
		result, err := slashingMod.InactiveStakingPools(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	slashingQuery.AddCommand(inactivePoolsCmd)

	// slashed-staking-pools
	slashedPoolsCmd := cli.NewCommand("slashed-staking-pools")
	slashedPoolsCmd.Short = "Query slashed staking pools"
	slashedPoolsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		slashingMod := slashing.New(client)
		result, err := slashingMod.SlashedStakingPools(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	slashingQuery.AddCommand(slashedPoolsCmd)

	// slash-proposals
	slashProposalsCmd := cli.NewCommand("slash-proposals")
	slashProposalsCmd.Short = "Query slash proposals"
	slashProposalsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		slashingMod := slashing.New(client)
		result, err := slashingMod.SlashProposals(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	slashingQuery.AddCommand(slashProposalsCmd)

	return slashingQuery
}

// buildQueryDistributorCommand builds the query distributor command group.
func (a *App) buildQueryDistributorCommand() *cli.Command {
	distributorQuery := cli.NewCommand("distributor")
	distributorQuery.Short = "Distributor query commands"

	// fees-treasury
	feesTreasuryCmd := cli.NewCommand("fees-treasury")
	feesTreasuryCmd.Short = "Query fees treasury"
	feesTreasuryCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		distributorMod := distributor.New(client)
		result, err := distributorMod.FeesTreasury(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	distributorQuery.AddCommand(feesTreasuryCmd)

	// periodic-snapshot
	periodicSnapshotCmd := cli.NewCommand("periodic-snapshot")
	periodicSnapshotCmd.Short = "Query periodic snapshot"
	periodicSnapshotCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		distributorMod := distributor.New(client)
		result, err := distributorMod.PeriodicSnapshot(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	distributorQuery.AddCommand(periodicSnapshotCmd)

	// snapshot-period
	snapshotPeriodCmd := cli.NewCommand("snapshot-period")
	snapshotPeriodCmd.Short = "Query snapshot period"
	snapshotPeriodCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		distributorMod := distributor.New(client)
		result, err := distributorMod.SnapshotPeriod(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	distributorQuery.AddCommand(snapshotPeriodCmd)

	// snapshot-period-performance
	snapshotPerfCmd := cli.NewCommand("snapshot-period-performance")
	snapshotPerfCmd.Short = "Query snapshot period performance for a validator"
	snapshotPerfCmd.Args = []cli.Arg{{Name: "validator", Required: true}}
	snapshotPerfCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("validator address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		distributorMod := distributor.New(client)
		result, err := distributorMod.SnapshotPeriodPerformance(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	distributorQuery.AddCommand(snapshotPerfCmd)

	// year-start-snapshot
	yearStartCmd := cli.NewCommand("year-start-snapshot")
	yearStartCmd.Short = "Query year start snapshot"
	yearStartCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		distributorMod := distributor.New(client)
		result, err := distributorMod.YearStartSnapshot(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	distributorQuery.AddCommand(yearStartCmd)

	return distributorQuery
}

// buildQueryBasketCommand builds the query basket command group.
func (a *App) buildQueryBasketCommand() *cli.Command {
	basketQuery := cli.NewCommand("basket")
	basketQuery.Short = "Basket query commands"

	// token-baskets
	tokenBasketsCmd := cli.NewCommand("token-baskets")
	tokenBasketsCmd.Short = "Query token baskets"
	tokenBasketsCmd.Args = []cli.Arg{
		{Name: "tokens", Required: true},
		{Name: "derivatives_only", Required: true},
	}
	tokenBasketsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("tokens and derivatives_only required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		derivOnly := ctx.Args[1] == "true"
		result, err := basketMod.TokenBaskets(context.Background(), ctx.Args[0], derivOnly)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	basketQuery.AddCommand(tokenBasketsCmd)

	// token-basket-by-id
	basketByIDCmd := cli.NewCommand("token-basket-by-id")
	basketByIDCmd.Short = "Query token basket by ID"
	basketByIDCmd.Args = []cli.Arg{{Name: "id", Required: true}}
	basketByIDCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("basket ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		result, err := basketMod.TokenBasketByID(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	basketQuery.AddCommand(basketByIDCmd)

	// token-basket-by-denom
	basketByDenomCmd := cli.NewCommand("token-basket-by-denom")
	basketByDenomCmd.Short = "Query token basket by denom"
	basketByDenomCmd.Args = []cli.Arg{{Name: "denom", Required: true}}
	basketByDenomCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("denom required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		result, err := basketMod.TokenBasketByDenom(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	basketQuery.AddCommand(basketByDenomCmd)

	// historical-mints
	histMintsCmd := cli.NewCommand("historical-mints")
	histMintsCmd.Short = "Query historical mints for a basket"
	histMintsCmd.Args = []cli.Arg{{Name: "basket-id", Required: true}}
	histMintsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("basket ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		result, err := basketMod.HistoricalMints(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	basketQuery.AddCommand(histMintsCmd)

	// historical-burns
	histBurnsCmd := cli.NewCommand("historical-burns")
	histBurnsCmd.Short = "Query historical burns for a basket"
	histBurnsCmd.Args = []cli.Arg{{Name: "basket-id", Required: true}}
	histBurnsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("basket ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		result, err := basketMod.HistoricalBurns(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	basketQuery.AddCommand(histBurnsCmd)

	// historical-swaps
	histSwapsCmd := cli.NewCommand("historical-swaps")
	histSwapsCmd.Short = "Query historical swaps for a basket"
	histSwapsCmd.Args = []cli.Arg{{Name: "basket-id", Required: true}}
	histSwapsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("basket ID required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		result, err := basketMod.HistoricalSwaps(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	basketQuery.AddCommand(histSwapsCmd)

	return basketQuery
}

// buildQueryCollectivesCommand builds the query collectives command group.
func (a *App) buildQueryCollectivesCommand() *cli.Command {
	collectivesQuery := cli.NewCommand("collectives")
	collectivesQuery.Short = "Collectives query commands"

	// collectives
	collectivesCmd := cli.NewCommand("collectives")
	collectivesCmd.Short = "Query all staking collectives"
	collectivesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		result, err := collectivesMod.Collectives(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	collectivesQuery.AddCommand(collectivesCmd)

	// collective
	collectiveCmd := cli.NewCommand("collective")
	collectiveCmd.Short = "Query a collective by name"
	collectiveCmd.Args = []cli.Arg{{Name: "name", Required: true}}
	collectiveCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("collective name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		result, err := collectivesMod.Collective(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	collectivesQuery.AddCommand(collectiveCmd)

	// collectives-by-account
	collectivesByAccountCmd := cli.NewCommand("collectives-by-account")
	collectivesByAccountCmd.Short = "Query collectives by account"
	collectivesByAccountCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	collectivesByAccountCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		result, err := collectivesMod.CollectivesByAccount(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	collectivesQuery.AddCommand(collectivesByAccountCmd)

	// collectives-proposals
	collectivesProposalsCmd := cli.NewCommand("collectives-proposals")
	collectivesProposalsCmd.Short = "Query collectives proposals"
	collectivesProposalsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		result, err := collectivesMod.CollectivesProposals(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	collectivesQuery.AddCommand(collectivesProposalsCmd)

	return collectivesQuery
}

// buildQueryCustodyCommand builds the query custody command group.
func (a *App) buildQueryCustodyCommand() *cli.Command {
	custodyQuery := cli.NewCommand("custody")
	custodyQuery.Short = "Custody query commands"

	// get
	getCmd := cli.NewCommand("get")
	getCmd.Short = "Query custody assigned to an address"
	getCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	getCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		custodyMod := custody.New(client)
		result, err := custodyMod.Get(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	custodyQuery.AddCommand(getCmd)

	// custodians
	custodiansCmd := cli.NewCommand("custodians")
	custodiansCmd.Short = "Query custody custodians"
	custodiansCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	custodiansCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		custodyMod := custody.New(client)
		result, err := custodyMod.Custodians(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	custodyQuery.AddCommand(custodiansCmd)

	// whitelist
	whitelistCmd := cli.NewCommand("whitelist")
	whitelistCmd.Short = "Query custody whitelist"
	whitelistCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	whitelistCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		custodyMod := custody.New(client)
		result, err := custodyMod.Whitelist(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	custodyQuery.AddCommand(whitelistCmd)

	// limits
	limitsCmd := cli.NewCommand("limits")
	limitsCmd.Short = "Query custody limits"
	limitsCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	limitsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		custodyMod := custody.New(client)
		result, err := custodyMod.Limits(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	custodyQuery.AddCommand(limitsCmd)

	return custodyQuery
}

// buildQueryBridgeCommand builds the query bridge command group.
func (a *App) buildQueryBridgeCommand() *cli.Command {
	bridgeQuery := cli.NewCommand("bridge")
	bridgeQuery.Short = "Bridge query commands"

	// get_cosmos_ethereum
	cosmosEthCmd := cli.NewCommand("get_cosmos_ethereum")
	cosmosEthCmd.Short = "Query Cosmos to Ethereum changes for an address"
	cosmosEthCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	cosmosEthCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bridgeMod := bridge.New(client)
		result, err := bridgeMod.GetCosmosEthereum(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	bridgeQuery.AddCommand(cosmosEthCmd)

	// get_ethereum_cosmos
	ethCosmosCmd := cli.NewCommand("get_ethereum_cosmos")
	ethCosmosCmd.Short = "Query Ethereum to Cosmos changes for an address"
	ethCosmosCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	ethCosmosCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bridgeMod := bridge.New(client)
		result, err := bridgeMod.GetEthereumCosmos(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	bridgeQuery.AddCommand(ethCosmosCmd)

	return bridgeQuery
}

// buildQueryLayer2Command builds the query layer2 command group.
func (a *App) buildQueryLayer2Command() *cli.Command {
	layer2Query := cli.NewCommand("layer2")
	layer2Query.Short = "Layer2 query commands"

	// all-dapps
	allDappsCmd := cli.NewCommand("all-dapps")
	allDappsCmd.Short = "Query all dapps"
	allDappsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		layer2Mod := layer2.New(client)
		result, err := layer2Mod.AllDapps(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	layer2Query.AddCommand(allDappsCmd)

	// execution-registrar
	execRegCmd := cli.NewCommand("execution-registrar")
	execRegCmd.Short = "Query execution registrar for a dapp"
	execRegCmd.Args = []cli.Arg{{Name: "dapp-name", Required: true}}
	execRegCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("dapp name required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		layer2Mod := layer2.New(client)
		result, err := layer2Mod.ExecutionRegistrar(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	layer2Query.AddCommand(execRegCmd)

	// transfer-dapps
	transferDappsCmd := cli.NewCommand("transfer-dapps")
	transferDappsCmd.Short = "Query transfer dapps"
	transferDappsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		layer2Mod := layer2.New(client)
		result, err := layer2Mod.TransferDapps(context.Background())
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	layer2Query.AddCommand(transferDappsCmd)

	return layer2Query
}

// buildQueryRecoveryCommand builds the query recovery command group.
func (a *App) buildQueryRecoveryCommand() *cli.Command {
	recoveryQuery := cli.NewCommand("recovery")
	recoveryQuery.Short = "Recovery query commands"

	// recovery-record
	recoveryRecordCmd := cli.NewCommand("recovery-record")
	recoveryRecordCmd.Short = "Query recovery record for an account"
	recoveryRecordCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	recoveryRecordCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		recoveryMod := recovery.New(client)
		result, err := recoveryMod.RecoveryRecord(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	recoveryQuery.AddCommand(recoveryRecordCmd)

	// recovery-token
	recoveryTokenCmd := cli.NewCommand("recovery-token")
	recoveryTokenCmd.Short = "Query recovery token for an account"
	recoveryTokenCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	recoveryTokenCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		recoveryMod := recovery.New(client)
		result, err := recoveryMod.RecoveryToken(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	recoveryQuery.AddCommand(recoveryTokenCmd)

	// rr-holder-rewards
	rrHolderRewardsCmd := cli.NewCommand("rr-holder-rewards")
	rrHolderRewardsCmd.Short = "Query RR holder rewards for an account"
	rrHolderRewardsCmd.Args = []cli.Arg{{Name: "address", Required: true}}
	rrHolderRewardsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("address required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		recoveryMod := recovery.New(client)
		result, err := recoveryMod.RRHolderRewards(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	recoveryQuery.AddCommand(rrHolderRewardsCmd)

	// rr-holders
	rrHoldersCmd := cli.NewCommand("rr-holders")
	rrHoldersCmd.Short = "Query registered RR holders for a token"
	rrHoldersCmd.Args = []cli.Arg{{Name: "rr_token", Required: true}}
	rrHoldersCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("rr_token required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		recoveryMod := recovery.New(client)
		result, err := recoveryMod.RRHolders(context.Background(), ctx.Args[0])
		if err != nil {
			return err
		}
		return a.printOutput(ctx, result)
	}
	recoveryQuery.AddCommand(rrHoldersCmd)

	return recoveryQuery
}

// buildTxCommand builds the tx command group.
func (a *App) buildTxCommand() *cli.Command {
	txCmd := cli.NewCommand("tx")
	txCmd.Short = "Transaction commands"
	txCmd.Long = "Submit transactions to the blockchain."

	// Add bank subcommand to tx
	bankTx := cli.NewCommand("bank")
	bankTx.Short = "Bank transaction commands"

	sendCmd := cli.NewCommand("send")
	sendCmd.Short = "Send tokens"
	sendCmd.Args = []cli.Arg{
		{Name: "from", Required: true},
		{Name: "to", Required: true},
		{Name: "amount", Required: true},
	}
	cli.AddTxFlags(sendCmd)
	sendCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 3 {
			return fmt.Errorf("from, to, and amount required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)

		coins, err := types.ParseCoins(ctx.Args[2])
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		opts := &bank.SendOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}

		resp, err := bankMod.Send(context.Background(), ctx.Args[0], ctx.Args[1], coins, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	bankTx.AddCommand(sendCmd)

	// multi-send
	multiSendCmd := cli.NewCommand("multi-send")
	multiSendCmd.Short = "Send tokens to multiple recipients"
	multiSendCmd.Long = "Send funds from one account to two or more accounts. By default, sends the amount to each address. Using --split, the amount is split equally between addresses."
	multiSendCmd.Args = []cli.Arg{
		{Name: "from", Required: true, Description: "Sender key name"},
		{Name: "to...", Required: true, Description: "Recipient addresses (space-separated)"},
		{Name: "amount", Required: true, Description: "Amount to send (e.g., 100ukex)"},
	}
	multiSendCmd.Flags = append(multiSendCmd.Flags, cli.Flag{Name: "split", Usage: "Split amount equally between recipients"})
	cli.AddTxFlags(multiSendCmd)
	multiSendCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 3 {
			return fmt.Errorf("from, at least one recipient, and amount required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bankMod := bank.New(client)

		// Last arg is amount, first is from, rest are recipients
		from := ctx.Args[0]
		amount := ctx.Args[len(ctx.Args)-1]
		toAddresses := ctx.Args[1 : len(ctx.Args)-1]

		if len(toAddresses) < 1 {
			return fmt.Errorf("at least one recipient address required")
		}

		coins, err := types.ParseCoins(amount)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		opts := &bank.MultiSendOptions{
			SendOptions: bank.SendOptions{
				Fees:          ctx.GetFlag("fees"),
				Gas:           ctx.GetFlag("gas"),
				Memo:          ctx.GetFlag("memo"),
				BroadcastMode: ctx.GetFlag("broadcast-mode"),
			},
			Split: ctx.GetFlag("split") == "true",
		}

		resp, err := bankMod.MultiSend(context.Background(), from, toAddresses, coins, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	bankTx.AddCommand(multiSendCmd)

	txCmd.AddCommand(bankTx)

	// Add multistaking subcommand to tx
	multistakingTx := cli.NewCommand("multistaking")
	multistakingTx.Short = "Multistaking transaction commands"

	// delegate
	delegateCmd := cli.NewCommand("delegate")
	delegateCmd.Short = "Delegate tokens to a validator pool"
	delegateCmd.Args = []cli.Arg{
		{Name: "validator", Required: true},
		{Name: "coins", Required: true},
	}
	cli.AddTxFlags(delegateCmd)
	delegateCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("validator and coins required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.Delegate(context.Background(), from, ctx.Args[0], ctx.Args[1], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(delegateCmd)

	// undelegate
	undelegateCmd := cli.NewCommand("undelegate")
	undelegateCmd.Short = "Start undelegation from a validator pool"
	undelegateCmd.Args = []cli.Arg{
		{Name: "validator", Required: true},
		{Name: "coins", Required: true},
	}
	cli.AddTxFlags(undelegateCmd)
	undelegateCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("validator and coins required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.Undelegate(context.Background(), from, ctx.Args[0], ctx.Args[1], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(undelegateCmd)

	// claim-rewards
	claimRewardsCmd := cli.NewCommand("claim-rewards")
	claimRewardsCmd.Short = "Claim rewards from a validator pool"
	claimRewardsCmd.Args = []cli.Arg{
		{Name: "validator", Required: true},
	}
	cli.AddTxFlags(claimRewardsCmd)
	claimRewardsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("validator required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.ClaimRewards(context.Background(), from, ctx.Args[0], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(claimRewardsCmd)

	// claim-undelegation
	claimUndelegationCmd := cli.NewCommand("claim-undelegation")
	claimUndelegationCmd.Short = "Claim a matured undelegation"
	claimUndelegationCmd.Args = []cli.Arg{
		{Name: "undelegation-id", Required: true},
	}
	cli.AddTxFlags(claimUndelegationCmd)
	claimUndelegationCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("undelegation-id required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.ClaimUndelegation(context.Background(), from, ctx.Args[0], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(claimUndelegationCmd)

	// claim-matured-undelegations
	claimMaturedCmd := cli.NewCommand("claim-matured-undelegations")
	claimMaturedCmd.Short = "Claim all matured undelegations"
	cli.AddTxFlags(claimMaturedCmd)
	claimMaturedCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.ClaimMaturedUndelegations(context.Background(), from, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(claimMaturedCmd)

	// register-delegator
	registerDelCmd := cli.NewCommand("register-delegator")
	registerDelCmd.Short = "Register as a delegator"
	cli.AddTxFlags(registerDelCmd)
	registerDelCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.RegisterDelegator(context.Background(), from, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(registerDelCmd)

	// set-compound-info
	setCompoundCmd := cli.NewCommand("set-compound-info")
	setCompoundCmd.Short = "Set compound info for staking"
	setCompoundCmd.Args = []cli.Arg{
		{Name: "all_denom", Required: true, Description: "Apply to all denoms (true/false)"},
		{Name: "specific_denoms", Required: true, Description: "Comma-separated list of specific denoms"},
	}
	cli.AddTxFlags(setCompoundCmd)
	setCompoundCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("all_denom and specific_denoms required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		allDenom := ctx.Args[0] == "true"
		opts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.SetCompoundInfo(context.Background(), from, allDenom, ctx.Args[1], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(setCompoundCmd)

	// upsert-staking-pool
	upsertPoolCmd := cli.NewCommand("upsert-staking-pool")
	upsertPoolCmd.Short = "Create or update a staking pool"
	upsertPoolCmd.Args = []cli.Arg{
		{Name: "validator_key", Required: true, Description: "Validator key for the pool"},
	}
	upsertPoolCmd.Flags = []cli.Flag{
		{Name: "enabled", Usage: "Enable the pool (default true)"},
		{Name: "commission", Usage: "Commission rate (0.01 to 0.5)"},
	}
	cli.AddTxFlags(upsertPoolCmd)
	upsertPoolCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("validator_key required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		msMod := multistaking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		poolOpts := &multistaking.UpsertStakingPoolOpts{
			Enabled:    ctx.GetFlag("enabled") == "true" || ctx.GetFlag("enabled") == "",
			Commission: ctx.GetFlag("commission"),
		}
		txOpts := &multistaking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := msMod.UpsertStakingPool(context.Background(), from, ctx.Args[0], poolOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	multistakingTx.AddCommand(upsertPoolCmd)

	txCmd.AddCommand(multistakingTx)

	// Add customgov subcommand to tx
	govTx := cli.NewCommand("customgov")
	govTx.Short = "Governance transaction commands"

	// proposal subcommand
	proposalTx := cli.NewCommand("proposal")
	proposalTx.Short = "Proposal transaction commands"

	// vote
	voteCmd := cli.NewCommand("vote")
	voteCmd.Short = "Vote on a proposal"
	voteCmd.Args = []cli.Arg{
		{Name: "proposal-id", Required: true},
		{Name: "vote-option", Required: true},
	}
	cli.AddTxFlags(voteCmd)
	voteCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("proposal-id and vote-option required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		voteOption := 0
		if _, err := fmt.Sscanf(ctx.Args[1], "%d", &voteOption); err != nil {
			return fmt.Errorf("invalid vote-option: %w", err)
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.VoteProposal(context.Background(), from, ctx.Args[0], voteOption, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(voteCmd)

	// proposal assign-role
	propAssignRoleCmd := cli.NewCommand("assign-role")
	propAssignRoleCmd.Short = "Create proposal to assign a role"
	propAssignRoleCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to assign role to", Required: true},
		{Name: "role", Usage: "Role to assign", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propAssignRoleCmd)
	propAssignRoleCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalAssignRole(context.Background(), from, ctx.GetFlag("addr"), ctx.GetFlag("role"), ctx.GetFlag("title"), ctx.GetFlag("description"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propAssignRoleCmd)

	// proposal unassign-role
	propUnassignRoleCmd := cli.NewCommand("unassign-role")
	propUnassignRoleCmd.Short = "Create proposal to unassign a role"
	propUnassignRoleCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to unassign role from", Required: true},
		{Name: "role", Usage: "Role to unassign", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propUnassignRoleCmd)
	propUnassignRoleCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalUnassignRole(context.Background(), from, ctx.GetFlag("addr"), ctx.GetFlag("role"), ctx.GetFlag("title"), ctx.GetFlag("description"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propUnassignRoleCmd)

	// proposal set-network-property
	propSetNetworkPropertyCmd := cli.NewCommand("set-network-property")
	propSetNetworkPropertyCmd.Short = "Create proposal to set a network property"
	propSetNetworkPropertyCmd.Flags = []cli.Flag{
		{Name: "property", Usage: "Property name", Required: true},
		{Name: "value", Usage: "Property value", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propSetNetworkPropertyCmd)
	propSetNetworkPropertyCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalSetNetworkProperty(context.Background(), from, ctx.GetFlag("property"), ctx.GetFlag("value"), ctx.GetFlag("title"), ctx.GetFlag("description"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propSetNetworkPropertyCmd)

	// proposal account subcommand
	proposalAccountTx := cli.NewCommand("account")
	proposalAccountTx.Short = "Account permission proposal commands"

	// proposal account whitelist-permission
	propWhitelistAccPermCmd := cli.NewCommand("whitelist-permission")
	propWhitelistAccPermCmd.Short = "Create proposal to whitelist a permission for an account"
	propWhitelistAccPermCmd.Args = []cli.Arg{
		{Name: "permission-id", Required: true},
	}
	propWhitelistAccPermCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to whitelist permission for", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propWhitelistAccPermCmd)
	propWhitelistAccPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(0))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
			Addr:        ctx.GetFlag("addr"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalWhitelistAccountPermission(context.Background(), from, perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalAccountTx.AddCommand(propWhitelistAccPermCmd)

	// proposal account blacklist-permission
	propBlacklistAccPermCmd := cli.NewCommand("blacklist-permission")
	propBlacklistAccPermCmd.Short = "Create proposal to blacklist a permission for an account"
	propBlacklistAccPermCmd.Args = []cli.Arg{
		{Name: "permission-id", Required: true},
	}
	propBlacklistAccPermCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to blacklist permission for", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propBlacklistAccPermCmd)
	propBlacklistAccPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(0))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
			Addr:        ctx.GetFlag("addr"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalBlacklistAccountPermission(context.Background(), from, perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalAccountTx.AddCommand(propBlacklistAccPermCmd)

	// proposal account remove-whitelisted-permission
	propRemoveWhitelistAccPermCmd := cli.NewCommand("remove-whitelisted-permission")
	propRemoveWhitelistAccPermCmd.Short = "Create proposal to remove whitelisted permission from an account"
	propRemoveWhitelistAccPermCmd.Args = []cli.Arg{
		{Name: "permission-id", Required: true},
	}
	propRemoveWhitelistAccPermCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to remove whitelisted permission from", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propRemoveWhitelistAccPermCmd)
	propRemoveWhitelistAccPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(0))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
			Addr:        ctx.GetFlag("addr"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalRemoveWhitelistedAccountPermission(context.Background(), from, perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalAccountTx.AddCommand(propRemoveWhitelistAccPermCmd)

	// proposal account remove-blacklisted-permission
	propRemoveBlacklistAccPermCmd := cli.NewCommand("remove-blacklisted-permission")
	propRemoveBlacklistAccPermCmd.Short = "Create proposal to remove blacklisted permission from an account"
	propRemoveBlacklistAccPermCmd.Args = []cli.Arg{
		{Name: "permission-id", Required: true},
	}
	propRemoveBlacklistAccPermCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to remove blacklisted permission from", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propRemoveBlacklistAccPermCmd)
	propRemoveBlacklistAccPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(0))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
			Addr:        ctx.GetFlag("addr"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalRemoveBlacklistedAccountPermission(context.Background(), from, perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalAccountTx.AddCommand(propRemoveBlacklistAccPermCmd)

	proposalTx.AddCommand(proposalAccountTx)

	// proposal role subcommand
	proposalRoleTx := cli.NewCommand("role")
	proposalRoleTx.Short = "Role proposal commands"

	// proposal role create
	propCreateRoleCmd := cli.NewCommand("create")
	propCreateRoleCmd.Short = "Create proposal to create a new role"
	propCreateRoleCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "role-description", Required: true},
	}
	propCreateRoleCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "whitelist", Usage: "Whitelist permissions in format 1,2,3"},
		{Name: "blacklist", Usage: "Blacklist permissions in format 1,2,3"},
	}
	cli.AddTxFlags(propCreateRoleCmd)
	propCreateRoleCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalCreateRoleOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
			Whitelist:   ctx.GetFlag("whitelist"),
			Blacklist:   ctx.GetFlag("blacklist"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalCreateRole(context.Background(), from, ctx.GetArg(0), ctx.GetArg(1), propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalRoleTx.AddCommand(propCreateRoleCmd)

	// proposal role remove
	propRemoveRoleCmd := cli.NewCommand("remove")
	propRemoveRoleCmd.Short = "Create proposal to remove a role"
	propRemoveRoleCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
	}
	propRemoveRoleCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propRemoveRoleCmd)
	propRemoveRoleCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalRemoveRole(context.Background(), from, ctx.GetArg(0), propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalRoleTx.AddCommand(propRemoveRoleCmd)

	// proposal role whitelist-permission
	propWhitelistRolePermCmd := cli.NewCommand("whitelist-permission")
	propWhitelistRolePermCmd.Short = "Create proposal to whitelist a permission for a role"
	propWhitelistRolePermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission-id", Required: true},
	}
	propWhitelistRolePermCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propWhitelistRolePermCmd)
	propWhitelistRolePermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalWhitelistRolePermission(context.Background(), from, ctx.GetArg(0), perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalRoleTx.AddCommand(propWhitelistRolePermCmd)

	// proposal role blacklist-permission
	propBlacklistRolePermCmd := cli.NewCommand("blacklist-permission")
	propBlacklistRolePermCmd.Short = "Create proposal to blacklist a permission for a role"
	propBlacklistRolePermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission-id", Required: true},
	}
	propBlacklistRolePermCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propBlacklistRolePermCmd)
	propBlacklistRolePermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalBlacklistRolePermission(context.Background(), from, ctx.GetArg(0), perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalRoleTx.AddCommand(propBlacklistRolePermCmd)

	// proposal role remove-whitelisted-permission
	propRemoveWhitelistRolePermCmd := cli.NewCommand("remove-whitelisted-permission")
	propRemoveWhitelistRolePermCmd.Short = "Create proposal to remove whitelisted permission from a role"
	propRemoveWhitelistRolePermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission-id", Required: true},
	}
	propRemoveWhitelistRolePermCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propRemoveWhitelistRolePermCmd)
	propRemoveWhitelistRolePermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalRemoveWhitelistedRolePermission(context.Background(), from, ctx.GetArg(0), perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalRoleTx.AddCommand(propRemoveWhitelistRolePermCmd)

	// proposal role remove-blacklisted-permission
	propRemoveBlacklistRolePermCmd := cli.NewCommand("remove-blacklisted-permission")
	propRemoveBlacklistRolePermCmd.Short = "Create proposal to remove blacklisted permission from a role"
	propRemoveBlacklistRolePermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission-id", Required: true},
	}
	propRemoveBlacklistRolePermCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propRemoveBlacklistRolePermCmd)
	propRemoveBlacklistRolePermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission-id: %w", err)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalRemoveBlacklistedRolePermission(context.Background(), from, ctx.GetArg(0), perm, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalRoleTx.AddCommand(propRemoveBlacklistRolePermCmd)

	proposalTx.AddCommand(proposalRoleTx)

	// proposal set-poor-network-msgs
	propSetPoorNetworkMsgsCmd := cli.NewCommand("set-poor-network-msgs")
	propSetPoorNetworkMsgsCmd.Short = "Create proposal to set poor network messages"
	propSetPoorNetworkMsgsCmd.Args = []cli.Arg{
		{Name: "messages", Required: true},
	}
	propSetPoorNetworkMsgsCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propSetPoorNetworkMsgsCmd)
	propSetPoorNetworkMsgsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalSetPoorNetworkMsgs(context.Background(), from, ctx.GetArg(0), propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propSetPoorNetworkMsgsCmd)

	// proposal set-proposal-durations
	propSetProposalDurationsCmd := cli.NewCommand("set-proposal-durations")
	propSetProposalDurationsCmd.Short = "Create proposal to set proposal durations"
	propSetProposalDurationsCmd.Args = []cli.Arg{
		{Name: "proposal-types", Required: true},
		{Name: "durations", Required: true},
	}
	propSetProposalDurationsCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propSetProposalDurationsCmd)
	propSetProposalDurationsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalSetProposalDurations(context.Background(), from, ctx.GetArg(0), ctx.GetArg(1), propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propSetProposalDurationsCmd)

	// proposal upsert-data-registry
	propUpsertDataRegistryCmd := cli.NewCommand("upsert-data-registry")
	propUpsertDataRegistryCmd.Short = "Create proposal to upsert a data registry entry"
	propUpsertDataRegistryCmd.Args = []cli.Arg{
		{Name: "key", Required: true},
		{Name: "hash", Required: true},
		{Name: "reference", Required: true},
		{Name: "encoding", Required: true},
		{Name: "size", Required: true},
	}
	propUpsertDataRegistryCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propUpsertDataRegistryCmd)
	propUpsertDataRegistryCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalUpsertDataRegistry(context.Background(), from, ctx.GetArg(0), ctx.GetArg(1), ctx.GetArg(2), ctx.GetArg(3), ctx.GetArg(4), propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propUpsertDataRegistryCmd)

	// proposal set-execution-fees
	propSetExecutionFeesCmd := cli.NewCommand("set-execution-fees")
	propSetExecutionFeesCmd.Short = "Create proposal to set execution fees"
	propSetExecutionFeesCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "tx-types", Usage: "Transaction types (comma-separated)", Required: true},
		{Name: "execution-fees", Usage: "Execution fees (comma-separated)", Required: true},
		{Name: "failure-fees", Usage: "Failure fees (comma-separated)", Required: true},
		{Name: "timeouts", Usage: "Timeouts (comma-separated)", Required: true},
		{Name: "default-params", Usage: "Default parameters (comma-separated)", Required: true},
	}
	cli.AddTxFlags(propSetExecutionFeesCmd)
	propSetExecutionFeesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalSetExecutionFeesOpts{
			Title:         ctx.GetFlag("title"),
			Description:   ctx.GetFlag("description"),
			TxTypes:       ctx.GetFlag("tx-types"),
			ExecutionFees: ctx.GetFlag("execution-fees"),
			FailureFees:   ctx.GetFlag("failure-fees"),
			Timeouts:      ctx.GetFlag("timeouts"),
			DefaultParams: ctx.GetFlag("default-params"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalSetExecutionFees(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propSetExecutionFeesCmd)

	// proposal jail-councilor
	propJailCouncilorCmd := cli.NewCommand("jail-councilor")
	propJailCouncilorCmd.Short = "Create proposal to jail councilors"
	propJailCouncilorCmd.Args = []cli.Arg{
		{Name: "councilors", Required: true},
	}
	propJailCouncilorCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propJailCouncilorCmd)
	propJailCouncilorCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalJailCouncilor(context.Background(), from, ctx.GetArg(0), propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propJailCouncilorCmd)

	// proposal reset-whole-councilor-rank
	propResetCouncilorRankCmd := cli.NewCommand("reset-whole-councilor-rank")
	propResetCouncilorRankCmd.Short = "Create proposal to reset whole councilor rank"
	propResetCouncilorRankCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propResetCouncilorRankCmd)
	propResetCouncilorRankCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.ProposalResetWholeCouncilorRank(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	proposalTx.AddCommand(propResetCouncilorRankCmd)

	govTx.AddCommand(proposalTx)

	// councilor subcommand
	councilorTx := cli.NewCommand("councilor")
	councilorTx.Short = "Councilor transaction commands"

	// councilor claim-seat
	claimSeatCouncilCmd := cli.NewCommand("claim-seat")
	claimSeatCouncilCmd.Short = "Claim a councilor seat"
	claimSeatCouncilCmd.Flags = []cli.Flag{
		{Name: "moniker", Usage: "The councilor moniker"},
	}
	cli.AddTxFlags(claimSeatCouncilCmd)
	claimSeatCouncilCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.CouncilorClaimSeat(context.Background(), from, ctx.GetFlag("moniker"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	councilorTx.AddCommand(claimSeatCouncilCmd)

	// councilor activate
	activateCmd := cli.NewCommand("activate")
	activateCmd.Short = "Activate a councilor"
	cli.AddTxFlags(activateCmd)
	activateCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.CouncilorActivate(context.Background(), from, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	councilorTx.AddCommand(activateCmd)

	// councilor pause
	pauseCmd := cli.NewCommand("pause")
	pauseCmd.Short = "Pause a councilor"
	cli.AddTxFlags(pauseCmd)
	pauseCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.CouncilorPause(context.Background(), from, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	councilorTx.AddCommand(pauseCmd)

	// councilor unpause
	unpauseCmd := cli.NewCommand("unpause")
	unpauseCmd.Short = "Unpause a councilor"
	cli.AddTxFlags(unpauseCmd)
	unpauseCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.CouncilorUnpause(context.Background(), from, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	councilorTx.AddCommand(unpauseCmd)
	govTx.AddCommand(councilorTx)

	// permission subcommand
	permissionTx := cli.NewCommand("permission")
	permissionTx.Short = "Permission transaction commands"

	// permission whitelist
	permWhitelistCmd := cli.NewCommand("whitelist")
	permWhitelistCmd.Short = "Whitelist a permission for an address"
	permWhitelistCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to whitelist permission for", Required: true},
		{Name: "permission", Usage: "Permission ID", Required: true},
	}
	cli.AddTxFlags(permWhitelistCmd)
	permWhitelistCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		permission := 0
		if _, err := fmt.Sscanf(ctx.GetFlag("permission"), "%d", &permission); err != nil {
			return fmt.Errorf("invalid permission: %w", err)
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.PermissionWhitelist(context.Background(), from, ctx.GetFlag("addr"), permission, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	permissionTx.AddCommand(permWhitelistCmd)

	// permission blacklist
	permBlacklistCmd := cli.NewCommand("blacklist")
	permBlacklistCmd.Short = "Blacklist a permission for an address"
	permBlacklistCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to blacklist permission for", Required: true},
		{Name: "permission", Usage: "Permission ID", Required: true},
	}
	cli.AddTxFlags(permBlacklistCmd)
	permBlacklistCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		permission := 0
		if _, err := fmt.Sscanf(ctx.GetFlag("permission"), "%d", &permission); err != nil {
			return fmt.Errorf("invalid permission: %w", err)
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.PermissionBlacklist(context.Background(), from, ctx.GetFlag("addr"), permission, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	permissionTx.AddCommand(permBlacklistCmd)

	// remove-whitelisted
	permRemoveWhitelistCmd := cli.NewCommand("remove-whitelisted")
	permRemoveWhitelistCmd.Short = "Remove whitelisted permission from address"
	permRemoveWhitelistCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to modify", Required: true},
		{Name: "permission", Usage: "Permission number", Required: true},
	}
	cli.AddTxFlags(permRemoveWhitelistCmd)
	permRemoveWhitelistCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetFlag("permission"))
		if err != nil {
			return fmt.Errorf("invalid permission number: %w", err)
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.PermissionRemoveWhitelisted(context.Background(), from, ctx.GetFlag("addr"), perm, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	permissionTx.AddCommand(permRemoveWhitelistCmd)

	// remove-blacklisted
	permRemoveBlacklistCmd := cli.NewCommand("remove-blacklisted")
	permRemoveBlacklistCmd.Short = "Remove blacklisted permission from address"
	permRemoveBlacklistCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to modify", Required: true},
		{Name: "permission", Usage: "Permission number", Required: true},
	}
	cli.AddTxFlags(permRemoveBlacklistCmd)
	permRemoveBlacklistCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		perm, err := strconv.Atoi(ctx.GetFlag("permission"))
		if err != nil {
			return fmt.Errorf("invalid permission number: %w", err)
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.PermissionRemoveBlacklisted(context.Background(), from, ctx.GetFlag("addr"), perm, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	permissionTx.AddCommand(permRemoveBlacklistCmd)

	govTx.AddCommand(permissionTx)

	// role subcommand
	roleTx := cli.NewCommand("role")
	roleTx.Short = "Role transaction commands"

	// role create
	roleCreateCmd := cli.NewCommand("create")
	roleCreateCmd.Short = "Create a new role"
	roleCreateCmd.Flags = []cli.Flag{
		{Name: "sid", Usage: "Role string ID", Required: true},
		{Name: "description", Usage: "Role description"},
	}
	cli.AddTxFlags(roleCreateCmd)
	roleCreateCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleCreate(context.Background(), from, ctx.GetFlag("sid"), ctx.GetFlag("description"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleCreateCmd)

	// role assign
	roleAssignCmd := cli.NewCommand("assign")
	roleAssignCmd.Short = "Assign a role to an account"
	roleAssignCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to assign role to", Required: true},
		{Name: "role", Usage: "Role ID", Required: true},
	}
	cli.AddTxFlags(roleAssignCmd)
	roleAssignCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		roleID := 0
		if _, err := fmt.Sscanf(ctx.GetFlag("role"), "%d", &roleID); err != nil {
			return fmt.Errorf("invalid role: %w", err)
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleAssign(context.Background(), from, ctx.GetFlag("addr"), roleID, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleAssignCmd)

	// role unassign
	roleUnassignCmd := cli.NewCommand("unassign")
	roleUnassignCmd.Short = "Unassign a role from an account"
	roleUnassignCmd.Flags = []cli.Flag{
		{Name: "addr", Usage: "Address to unassign role from", Required: true},
		{Name: "role", Usage: "Role ID", Required: true},
	}
	cli.AddTxFlags(roleUnassignCmd)
	roleUnassignCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		roleID := 0
		if _, err := fmt.Sscanf(ctx.GetFlag("role"), "%d", &roleID); err != nil {
			return fmt.Errorf("invalid role: %w", err)
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleUnassign(context.Background(), from, ctx.GetFlag("addr"), roleID, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleUnassignCmd)

	// whitelist-permission
	roleWhitelistPermCmd := cli.NewCommand("whitelist-permission")
	roleWhitelistPermCmd.Short = "Whitelist a permission for a role"
	roleWhitelistPermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission", Required: true},
	}
	cli.AddTxFlags(roleWhitelistPermCmd)
	roleWhitelistPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		roleSID := ctx.GetArg(0)
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission number: %w", err)
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleWhitelistPermission(context.Background(), from, roleSID, perm, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleWhitelistPermCmd)

	// blacklist-permission
	roleBlacklistPermCmd := cli.NewCommand("blacklist-permission")
	roleBlacklistPermCmd.Short = "Blacklist a permission for a role"
	roleBlacklistPermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission", Required: true},
	}
	cli.AddTxFlags(roleBlacklistPermCmd)
	roleBlacklistPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		roleSID := ctx.GetArg(0)
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission number: %w", err)
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleBlacklistPermission(context.Background(), from, roleSID, perm, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleBlacklistPermCmd)

	// remove-whitelisted-permission
	roleRemoveWhitelistPermCmd := cli.NewCommand("remove-whitelisted-permission")
	roleRemoveWhitelistPermCmd.Short = "Remove whitelisted permission from role"
	roleRemoveWhitelistPermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission", Required: true},
	}
	cli.AddTxFlags(roleRemoveWhitelistPermCmd)
	roleRemoveWhitelistPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		roleSID := ctx.GetArg(0)
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission number: %w", err)
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleRemoveWhitelistedPermission(context.Background(), from, roleSID, perm, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleRemoveWhitelistPermCmd)

	// remove-blacklisted-permission
	roleRemoveBlacklistPermCmd := cli.NewCommand("remove-blacklisted-permission")
	roleRemoveBlacklistPermCmd.Short = "Remove blacklisted permission from role"
	roleRemoveBlacklistPermCmd.Args = []cli.Arg{
		{Name: "role-sid", Required: true},
		{Name: "permission", Required: true},
	}
	cli.AddTxFlags(roleRemoveBlacklistPermCmd)
	roleRemoveBlacklistPermCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		roleSID := ctx.GetArg(0)
		perm, err := strconv.Atoi(ctx.GetArg(1))
		if err != nil {
			return fmt.Errorf("invalid permission number: %w", err)
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RoleRemoveBlacklistedPermission(context.Background(), from, roleSID, perm, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	roleTx.AddCommand(roleRemoveBlacklistPermCmd)

	govTx.AddCommand(roleTx)

	// poll subcommand
	pollTx := cli.NewCommand("poll")
	pollTx.Short = "Poll transaction commands"

	// poll create
	pollCreateCmd := cli.NewCommand("create")
	pollCreateCmd.Short = "Create a new poll"
	pollCreateCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Poll title", Required: true},
		{Name: "description", Usage: "Poll description"},
		{Name: "reference", Usage: "Poll reference"},
		{Name: "checksum", Usage: "Poll checksum"},
		{Name: "roles", Usage: "Roles that can vote"},
		{Name: "poll-type", Usage: "Poll type"},
		{Name: "options", Usage: "Poll options"},
		{Name: "selection-count", Usage: "Number of selections allowed"},
		{Name: "duration", Usage: "Poll duration in seconds"},
	}
	cli.AddTxFlags(pollCreateCmd)
	pollCreateCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		pollOpts := &gov.PollCreateOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
			Reference:   ctx.GetFlag("reference"),
			Checksum:    ctx.GetFlag("checksum"),
			Roles:       ctx.GetFlag("roles"),
			PollType:    ctx.GetFlag("poll-type"),
			Options:     ctx.GetFlag("options"),
		}
		if sc := ctx.GetFlag("selection-count"); sc != "" {
			fmt.Sscanf(sc, "%d", &pollOpts.SelectionCount)
		}
		if dur := ctx.GetFlag("duration"); dur != "" {
			fmt.Sscanf(dur, "%d", &pollOpts.Duration)
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.PollCreate(context.Background(), from, pollOpts, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	pollTx.AddCommand(pollCreateCmd)

	// poll vote
	pollVoteCmd := cli.NewCommand("vote")
	pollVoteCmd.Short = "Vote on a poll"
	pollVoteCmd.Flags = []cli.Flag{
		{Name: "poll-id", Usage: "Poll ID", Required: true},
		{Name: "options", Usage: "Vote options", Required: true},
	}
	cli.AddTxFlags(pollVoteCmd)
	pollVoteCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.PollVote(context.Background(), from, ctx.GetFlag("poll-id"), ctx.GetFlag("options"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	pollTx.AddCommand(pollVoteCmd)
	govTx.AddCommand(pollTx)

	// Identity commands
	// register-identity-records
	registerIdRecordsCmd := cli.NewCommand("register-identity-records")
	registerIdRecordsCmd.Short = "Register identity records"
	registerIdRecordsCmd.Flags = []cli.Flag{
		{Name: "infos-json", Usage: "JSON string with identity records", Required: true},
	}
	cli.AddTxFlags(registerIdRecordsCmd)
	registerIdRecordsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		infosJSON := ctx.GetFlag("infos-json")
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RegisterIdentityRecords(context.Background(), from, infosJSON, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(registerIdRecordsCmd)

	// delete-identity-records
	deleteIdRecordsCmd := cli.NewCommand("delete-identity-records")
	deleteIdRecordsCmd.Short = "Delete identity records"
	deleteIdRecordsCmd.Flags = []cli.Flag{
		{Name: "keys", Usage: "Comma-separated keys to delete", Required: true},
	}
	cli.AddTxFlags(deleteIdRecordsCmd)
	deleteIdRecordsCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		keys := ctx.GetFlag("keys")
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.DeleteIdentityRecords(context.Background(), from, keys, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(deleteIdRecordsCmd)

	// request-identity-record-verify
	requestIdVerifyCmd := cli.NewCommand("request-identity-record-verify")
	requestIdVerifyCmd.Short = "Request identity record verification"
	requestIdVerifyCmd.Flags = []cli.Flag{
		{Name: "verifier", Usage: "Verifier address", Required: true},
		{Name: "record-ids", Usage: "Comma-separated record IDs", Required: true},
		{Name: "verifier-tip", Usage: "Tip for verifier (required due to sekaid bug)", Required: true},
	}
	cli.AddTxFlags(requestIdVerifyCmd)
	requestIdVerifyCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.RequestIdentityRecordVerify(context.Background(), from, ctx.GetFlag("verifier"), ctx.GetFlag("record-ids"), ctx.GetFlag("verifier-tip"), txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(requestIdVerifyCmd)

	// handle-identity-records-verify-request
	handleIdVerifyCmd := cli.NewCommand("handle-identity-records-verify-request")
	handleIdVerifyCmd.Short = "Handle identity records verify request"
	handleIdVerifyCmd.Args = []cli.Arg{
		{Name: "request-id", Required: true},
	}
	handleIdVerifyCmd.Flags = []cli.Flag{
		{Name: "approve", Usage: "Approve the request (default: true)"},
	}
	cli.AddTxFlags(handleIdVerifyCmd)
	handleIdVerifyCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		requestID := ctx.GetArg(0)
		approve := ctx.GetFlag("approve") != "false"
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.HandleIdentityRecordsVerifyRequest(context.Background(), from, requestID, approve, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(handleIdVerifyCmd)

	// cancel-identity-records-verify-request
	cancelIdVerifyCmd := cli.NewCommand("cancel-identity-records-verify-request")
	cancelIdVerifyCmd.Short = "Cancel identity records verify request"
	cancelIdVerifyCmd.Args = []cli.Arg{
		{Name: "request-id", Required: true},
	}
	cli.AddTxFlags(cancelIdVerifyCmd)
	cancelIdVerifyCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		requestID := ctx.GetArg(0)
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.CancelIdentityRecordsVerifyRequest(context.Background(), from, requestID, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(cancelIdVerifyCmd)

	// set-network-properties (sudo/direct command)
	setNetworkPropertiesCmd := cli.NewCommand("set-network-properties")
	setNetworkPropertiesCmd.Short = "Set network properties (sudo only)"
	setNetworkPropertiesCmd.Flags = []cli.Flag{
		{Name: "min_tx_fee", Usage: "Minimum transaction fee"},
		{Name: "max_tx_fee", Usage: "Maximum transaction fee"},
		{Name: "min_validators", Usage: "Minimum validators"},
		{Name: "min_custody_reward", Usage: "Minimum custody reward"},
	}
	cli.AddTxFlags(setNetworkPropertiesCmd)
	setNetworkPropertiesCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		properties := make(map[string]string)
		if v := ctx.GetFlag("min_tx_fee"); v != "" {
			properties["min_tx_fee"] = v
		}
		if v := ctx.GetFlag("max_tx_fee"); v != "" {
			properties["max_tx_fee"] = v
		}
		if v := ctx.GetFlag("min_validators"); v != "" {
			properties["min_validators"] = v
		}
		if v := ctx.GetFlag("min_custody_reward"); v != "" {
			properties["min_custody_reward"] = v
		}
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.SetNetworkProperties(context.Background(), from, properties, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(setNetworkPropertiesCmd)

	// set-execution-fee (sudo/direct command)
	setExecutionFeeCmd := cli.NewCommand("set-execution-fee")
	setExecutionFeeCmd.Short = "Set execution fee (sudo only)"
	setExecutionFeeCmd.Flags = []cli.Flag{
		{Name: "transaction_type", Usage: "Transaction type", Required: true},
		{Name: "execution_fee", Usage: "Execution fee", Required: true},
		{Name: "failure_fee", Usage: "Failure fee", Required: true},
		{Name: "timeout", Usage: "Timeout", Required: true},
		{Name: "default_parameters", Usage: "Default parameters", Required: true},
	}
	cli.AddTxFlags(setExecutionFeeCmd)
	setExecutionFeeCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		govMod := gov.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		var executionFee, failureFee, timeout, defaultParams uint64
		fmt.Sscanf(ctx.GetFlag("execution_fee"), "%d", &executionFee)
		fmt.Sscanf(ctx.GetFlag("failure_fee"), "%d", &failureFee)
		fmt.Sscanf(ctx.GetFlag("timeout"), "%d", &timeout)
		fmt.Sscanf(ctx.GetFlag("default_parameters"), "%d", &defaultParams)
		txOpts := &gov.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := govMod.SetExecutionFee(context.Background(), from, ctx.GetFlag("transaction_type"), executionFee, failureFee, timeout, defaultParams, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	govTx.AddCommand(setExecutionFeeCmd)

	txCmd.AddCommand(govTx)

	// Add customstaking subcommand to tx
	stakingTx := cli.NewCommand("customstaking")
	stakingTx.Short = "Staking transaction commands"

	// claim-validator-seat
	claimSeatCmd := cli.NewCommand("claim-validator-seat")
	claimSeatCmd.Short = "Claim validator seat to become a Validator"
	claimSeatCmd.Flags = []cli.Flag{
		{Name: "moniker", Usage: "The validator moniker"},
		{Name: "pubkey", Usage: "The validator public key"},
	}
	cli.AddTxFlags(claimSeatCmd)
	claimSeatCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		stakingMod := staking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		seatOpts := &staking.ClaimValidatorSeatOpts{
			Moniker: ctx.GetFlag("moniker"),
			PubKey:  ctx.GetFlag("pubkey"),
		}
		txOpts := &staking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := stakingMod.ClaimValidatorSeat(context.Background(), from, seatOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	stakingTx.AddCommand(claimSeatCmd)

	// proposal unjail-validator
	unjailValCmd := cli.NewCommand("proposal-unjail-validator")
	unjailValCmd.Short = "Create proposal to unjail a validator"
	unjailValCmd.Args = []cli.Arg{
		{Name: "val_addr", Required: true, Description: "Validator address"},
		{Name: "reference", Required: true, Description: "Reference for proposal"},
	}
	unjailValCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title"},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(unjailValCmd)
	unjailValCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("val_addr and reference required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		stakingMod := staking.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &staking.ProposalUnjailValidatorOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &staking.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := stakingMod.ProposalUnjailValidator(context.Background(), from, ctx.Args[0], ctx.Args[1], propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	stakingTx.AddCommand(unjailValCmd)

	txCmd.AddCommand(stakingTx)

	// Add spending subcommand to tx
	spendingTx := cli.NewCommand("spending")
	spendingTx.Short = "Spending pool transaction commands"

	// claim-spending-pool
	claimPoolCmd := cli.NewCommand("claim-spending-pool")
	claimPoolCmd.Short = "Claim from spending pool"
	claimPoolCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "The name of the spending pool", Required: true},
	}
	cli.AddTxFlags(claimPoolCmd)
	claimPoolCmd.Run = func(ctx *cli.Context) error {
		poolName := ctx.GetFlag("name")
		if poolName == "" {
			return fmt.Errorf("--name flag required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.ClaimSpendingPool(context.Background(), from, poolName, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(claimPoolCmd)

	// deposit-spending-pool
	depositPoolCmd := cli.NewCommand("deposit-spending-pool")
	depositPoolCmd.Short = "Deposit into spending pool"
	depositPoolCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "The name of the spending pool", Required: true},
		{Name: "amount", Usage: "The amount of coins to deposit", Required: true},
	}
	cli.AddTxFlags(depositPoolCmd)
	depositPoolCmd.Run = func(ctx *cli.Context) error {
		poolName := ctx.GetFlag("name")
		if poolName == "" {
			return fmt.Errorf("--name flag required")
		}
		amount := ctx.GetFlag("amount")
		if amount == "" {
			return fmt.Errorf("--amount flag required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.DepositSpendingPool(context.Background(), from, poolName, amount, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(depositPoolCmd)

	// create-spending-pool
	createSpendingPoolCmd := cli.NewCommand("create-spending-pool")
	createSpendingPoolCmd.Short = "Create a new spending pool"
	createSpendingPoolCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "Pool name", Required: true},
		{Name: "claim-start", Usage: "Claim start timestamp"},
		{Name: "claim-end", Usage: "Claim end timestamp"},
		{Name: "claim-expiry", Usage: "Claim expiry time"},
		{Name: "rates", Usage: "Reward rates"},
		{Name: "vote-quorum", Usage: "Vote quorum"},
		{Name: "vote-period", Usage: "Vote period"},
		{Name: "vote-enactment", Usage: "Vote enactment period"},
		{Name: "owner-accounts", Usage: "Owner accounts"},
		{Name: "owner-roles", Usage: "Owner roles"},
		{Name: "beneficiary-accounts", Usage: "Beneficiary accounts"},
		{Name: "beneficiary-roles", Usage: "Beneficiary roles"},
		{Name: "dynamic-rate", Usage: "Enable dynamic rate"},
		{Name: "dynamic-rate-period", Usage: "Dynamic rate period"},
	}
	cli.AddTxFlags(createSpendingPoolCmd)
	createSpendingPoolCmd.Run = func(ctx *cli.Context) error {
		name := ctx.GetFlag("name")
		if name == "" {
			return fmt.Errorf("--name flag required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		var claimStart, claimEnd, claimExpiry, votePeriod, voteEnactment, dynamicRatePeriod uint64
		if v := ctx.GetFlag("claim-start"); v != "" {
			fmt.Sscanf(v, "%d", &claimStart)
		}
		if v := ctx.GetFlag("claim-end"); v != "" {
			fmt.Sscanf(v, "%d", &claimEnd)
		}
		if v := ctx.GetFlag("claim-expiry"); v != "" {
			fmt.Sscanf(v, "%d", &claimExpiry)
		}
		if v := ctx.GetFlag("vote-period"); v != "" {
			fmt.Sscanf(v, "%d", &votePeriod)
		}
		if v := ctx.GetFlag("vote-enactment"); v != "" {
			fmt.Sscanf(v, "%d", &voteEnactment)
		}
		if v := ctx.GetFlag("dynamic-rate-period"); v != "" {
			fmt.Sscanf(v, "%d", &dynamicRatePeriod)
		}
		poolOpts := &spending.CreateSpendingPoolOpts{
			Name:              name,
			ClaimStart:        claimStart,
			ClaimEnd:          claimEnd,
			ClaimExpiry:       claimExpiry,
			Rates:             ctx.GetFlag("rates"),
			VoteQuorum:        ctx.GetFlag("vote-quorum"),
			VotePeriod:        votePeriod,
			VoteEnactment:     voteEnactment,
			Owners:            ctx.GetFlag("owner-accounts"),
			Beneficiaries:     ctx.GetFlag("beneficiary-accounts"),
			DynamicRate:       ctx.GetFlag("dynamic-rate") == "true",
			DynamicRatePeriod: dynamicRatePeriod,
		}
		txOpts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.CreateSpendingPool(context.Background(), from, poolOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(createSpendingPoolCmd)

	// register-spending-pool-beneficiary
	registerBeneficiaryCmd := cli.NewCommand("register-spending-pool-beneficiary")
	registerBeneficiaryCmd.Short = "Register as spending pool beneficiary"
	registerBeneficiaryCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "Pool name", Required: true},
	}
	cli.AddTxFlags(registerBeneficiaryCmd)
	registerBeneficiaryCmd.Run = func(ctx *cli.Context) error {
		name := ctx.GetFlag("name")
		if name == "" {
			return fmt.Errorf("--name flag required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		txOpts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.RegisterSpendingPoolBeneficiary(context.Background(), from, name, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(registerBeneficiaryCmd)

	// proposal-spending-pool-distribution
	proposalDistributionCmd := cli.NewCommand("proposal-spending-pool-distribution")
	proposalDistributionCmd.Short = "Create a proposal to distribute from spending pool"
	proposalDistributionCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "Pool name", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(proposalDistributionCmd)
	proposalDistributionCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &spending.ProposalSpendingPoolDistributionOpts{
			Name:        ctx.GetFlag("name"),
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.ProposalSpendingPoolDistribution(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(proposalDistributionCmd)

	// proposal-spending-pool-withdraw
	proposalWithdrawCmd := cli.NewCommand("proposal-spending-pool-withdraw")
	proposalWithdrawCmd.Short = "Create a proposal to withdraw from spending pool"
	proposalWithdrawCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "Pool name", Required: true},
		{Name: "beneficiary-accounts", Usage: "Beneficiary accounts"},
		{Name: "amount", Usage: "Amount to withdraw"},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(proposalWithdrawCmd)
	proposalWithdrawCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &spending.ProposalSpendingPoolWithdrawOpts{
			Name:                ctx.GetFlag("name"),
			BeneficiaryAccounts: ctx.GetFlag("beneficiary-accounts"),
			Amount:              ctx.GetFlag("amount"),
			Title:               ctx.GetFlag("title"),
			Description:         ctx.GetFlag("description"),
		}
		txOpts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.ProposalSpendingPoolWithdraw(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(proposalWithdrawCmd)

	// proposal-update-spending-pool
	proposalUpdatePoolCmd := cli.NewCommand("proposal-update-spending-pool")
	proposalUpdatePoolCmd.Short = "Create a proposal to update spending pool"
	proposalUpdatePoolCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "Pool name", Required: true},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "claim-start", Usage: "Claim start timestamp"},
		{Name: "claim-end", Usage: "Claim end timestamp"},
		{Name: "rates", Usage: "Reward rates"},
		{Name: "vote-quorum", Usage: "Vote quorum"},
		{Name: "vote-period", Usage: "Vote period"},
		{Name: "vote-enactment", Usage: "Vote enactment period"},
		{Name: "owner-accounts", Usage: "Owner accounts"},
		{Name: "owner-roles", Usage: "Owner roles"},
		{Name: "beneficiary-accounts", Usage: "Beneficiary accounts"},
		{Name: "beneficiary-roles", Usage: "Beneficiary roles"},
		{Name: "beneficiary-account-weights", Usage: "Beneficiary account weights"},
		{Name: "beneficiary-role-weights", Usage: "Beneficiary role weights"},
		{Name: "dynamic-rate", Usage: "Enable dynamic rate"},
		{Name: "dynamic-rate-period", Usage: "Dynamic rate period"},
	}
	cli.AddTxFlags(proposalUpdatePoolCmd)
	proposalUpdatePoolCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		spendingMod := spending.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		claimStart, _ := strconv.ParseInt(ctx.GetFlag("claim-start"), 10, 32)
		claimEnd, _ := strconv.ParseInt(ctx.GetFlag("claim-end"), 10, 32)
		voteQuorum, _ := strconv.ParseInt(ctx.GetFlag("vote-quorum"), 10, 32)
		votePeriod, _ := strconv.ParseInt(ctx.GetFlag("vote-period"), 10, 32)
		voteEnactment, _ := strconv.ParseInt(ctx.GetFlag("vote-enactment"), 10, 32)
		propOpts := &spending.ProposalUpdateSpendingPoolOpts{
			Name:                      ctx.GetFlag("name"),
			Title:                     ctx.GetFlag("title"),
			Description:               ctx.GetFlag("description"),
			ClaimStart:                int32(claimStart),
			ClaimEnd:                  int32(claimEnd),
			Rates:                     ctx.GetFlag("rates"),
			VoteQuorum:                int32(voteQuorum),
			VotePeriod:                int32(votePeriod),
			VoteEnactment:             int32(voteEnactment),
			OwnerAccounts:             ctx.GetFlag("owner-accounts"),
			OwnerRoles:                ctx.GetFlag("owner-roles"),
			BeneficiaryAccounts:       ctx.GetFlag("beneficiary-accounts"),
			BeneficiaryRoles:          ctx.GetFlag("beneficiary-roles"),
			BeneficiaryAccountWeights: ctx.GetFlag("beneficiary-account-weights"),
			BeneficiaryRoleWeights:    ctx.GetFlag("beneficiary-role-weights"),
			DynamicRate:               ctx.GetFlag("dynamic-rate") == "true",
			DynamicRatePeriod:         ctx.GetFlag("dynamic-rate-period"),
		}
		txOpts := &spending.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := spendingMod.ProposalUpdateSpendingPool(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	spendingTx.AddCommand(proposalUpdatePoolCmd)

	txCmd.AddCommand(spendingTx)

	// Add basket subcommand to tx
	basketTx := cli.NewCommand("basket")
	basketTx.Short = "Basket transaction commands"

	// mint-basket-tokens
	mintBasketCmd := cli.NewCommand("mint-basket-tokens")
	mintBasketCmd.Short = "Mint basket tokens"
	mintBasketCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
		{Name: "deposit-coins", Required: true},
	}
	cli.AddTxFlags(mintBasketCmd)
	mintBasketCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("basket-id and deposit-coins required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.MintBasketTokens(context.Background(), from, ctx.Args[0], ctx.Args[1], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(mintBasketCmd)

	// burn-basket-tokens
	burnBasketCmd := cli.NewCommand("burn-basket-tokens")
	burnBasketCmd.Short = "Burn basket tokens"
	burnBasketCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
		{Name: "burn-amount", Required: true},
	}
	cli.AddTxFlags(burnBasketCmd)
	burnBasketCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("basket-id and burn-amount required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.BurnBasketTokens(context.Background(), from, ctx.Args[0], ctx.Args[1], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(burnBasketCmd)

	// swap-basket-tokens
	swapBasketCmd := cli.NewCommand("swap-basket-tokens")
	swapBasketCmd.Short = "Swap basket tokens"
	swapBasketCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
		{Name: "swap-in", Required: true},
		{Name: "swap-out", Required: true},
	}
	cli.AddTxFlags(swapBasketCmd)
	swapBasketCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 3 {
			return fmt.Errorf("basket-id, swap-in, and swap-out required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.SwapBasketTokens(context.Background(), from, ctx.Args[0], ctx.Args[1], ctx.Args[2], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(swapBasketCmd)

	// basket-claim-rewards
	claimBasketRewardsCmd := cli.NewCommand("basket-claim-rewards")
	claimBasketRewardsCmd.Short = "Claim rewards from a staking derivative basket"
	claimBasketRewardsCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
	}
	cli.AddTxFlags(claimBasketRewardsCmd)
	claimBasketRewardsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("basket-id required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.BasketClaimRewards(context.Background(), from, ctx.Args[0], opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(claimBasketRewardsCmd)

	// disable-basket-deposits
	disableDepositsCmd := cli.NewCommand("disable-basket-deposits")
	disableDepositsCmd.Short = "Disable deposits for a basket"
	disableDepositsCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
		{Name: "disabled", Required: true, Description: "true or false"},
	}
	cli.AddTxFlags(disableDepositsCmd)
	disableDepositsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("basket-id and disabled required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		disabled := ctx.Args[1] == "true"
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.DisableBasketDeposits(context.Background(), from, ctx.Args[0], disabled, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(disableDepositsCmd)

	// disable-basket-withdraws
	disableWithdrawsCmd := cli.NewCommand("disable-basket-withdraws")
	disableWithdrawsCmd.Short = "Disable withdraws for a basket"
	disableWithdrawsCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
		{Name: "disabled", Required: true, Description: "true or false"},
	}
	cli.AddTxFlags(disableWithdrawsCmd)
	disableWithdrawsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("basket-id and disabled required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		disabled := ctx.Args[1] == "true"
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.DisableBasketWithdraws(context.Background(), from, ctx.Args[0], disabled, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(disableWithdrawsCmd)

	// disable-basket-swaps
	disableSwapsCmd := cli.NewCommand("disable-basket-swaps")
	disableSwapsCmd.Short = "Disable swaps for a basket"
	disableSwapsCmd.Args = []cli.Arg{
		{Name: "basket-id", Required: true},
		{Name: "disabled", Required: true, Description: "true or false"},
	}
	cli.AddTxFlags(disableSwapsCmd)
	disableSwapsCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("basket-id and disabled required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		disabled := ctx.Args[1] == "true"
		opts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.DisableBasketSwaps(context.Background(), from, ctx.Args[0], disabled, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(disableSwapsCmd)

	// proposal-create-basket
	proposalCreateBasketCmd := cli.NewCommand("proposal-create-basket")
	proposalCreateBasketCmd.Short = "Create proposal to create a new basket"
	proposalCreateBasketCmd.Flags = []cli.Flag{
		{Name: "basket-suffix", Usage: "Basket suffix"},
		{Name: "basket-description", Usage: "Basket description"},
		{Name: "basket-tokens", Usage: "Comma-separated list of tokens in basket"},
		{Name: "tokens-cap", Usage: "Tokens cap"},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "mints-min", Usage: "Minimum mints"},
		{Name: "mints-max", Usage: "Maximum mints"},
		{Name: "mints-disabled", Usage: "Disable mints"},
		{Name: "burns-min", Usage: "Minimum burns"},
		{Name: "burns-max", Usage: "Maximum burns"},
		{Name: "burns-disabled", Usage: "Disable burns"},
		{Name: "swaps-min", Usage: "Minimum swaps"},
		{Name: "swaps-max", Usage: "Maximum swaps"},
		{Name: "swaps-disabled", Usage: "Disable swaps"},
		{Name: "swap-fee", Usage: "Swap fee"},
		{Name: "slippage-fee-min", Usage: "Minimum slippage fee"},
		{Name: "limits-period", Usage: "Limits period"},
	}
	cli.AddTxFlags(proposalCreateBasketCmd)
	proposalCreateBasketCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &basket.ProposalCreateBasketOpts{
			BasketSuffix:      ctx.GetFlag("basket-suffix"),
			BasketDescription: ctx.GetFlag("basket-description"),
			BasketTokens:      ctx.GetFlag("basket-tokens"),
			TokensCap:         ctx.GetFlag("tokens-cap"),
			Title:             ctx.GetFlag("title"),
			Description:       ctx.GetFlag("description"),
			MintsMin:          ctx.GetFlag("mints-min"),
			MintsMax:          ctx.GetFlag("mints-max"),
			MintsDisabled:     ctx.GetFlag("mints-disabled") == "true",
			BurnsMin:          ctx.GetFlag("burns-min"),
			BurnsMax:          ctx.GetFlag("burns-max"),
			BurnsDisabled:     ctx.GetFlag("burns-disabled") == "true",
			SwapsMin:          ctx.GetFlag("swaps-min"),
			SwapsMax:          ctx.GetFlag("swaps-max"),
			SwapsDisabled:     ctx.GetFlag("swaps-disabled") == "true",
			SwapFee:           ctx.GetFlag("swap-fee"),
			SlippageFeeMin:    ctx.GetFlag("slippage-fee-min"),
		}
		if ctx.GetFlag("limits-period") != "" {
			var limitsPeriod uint64
			fmt.Sscanf(ctx.GetFlag("limits-period"), "%d", &limitsPeriod)
			propOpts.LimitsPeriod = limitsPeriod
		}
		txOpts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.ProposalCreateBasket(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(proposalCreateBasketCmd)

	// proposal-edit-basket
	proposalEditBasketCmd := cli.NewCommand("proposal-edit-basket")
	proposalEditBasketCmd.Short = "Create proposal to edit an existing basket"
	proposalEditBasketCmd.Flags = []cli.Flag{
		{Name: "basket-id", Usage: "Basket ID to edit", Required: true},
		{Name: "basket-suffix", Usage: "Basket suffix"},
		{Name: "basket-description", Usage: "Basket description"},
		{Name: "basket-tokens", Usage: "Comma-separated list of tokens in basket"},
		{Name: "tokens-cap", Usage: "Tokens cap"},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "mints-min", Usage: "Minimum mints"},
		{Name: "mints-max", Usage: "Maximum mints"},
		{Name: "mints-disabled", Usage: "Disable mints"},
		{Name: "burns-min", Usage: "Minimum burns"},
		{Name: "burns-max", Usage: "Maximum burns"},
		{Name: "burns-disabled", Usage: "Disable burns"},
		{Name: "swaps-min", Usage: "Minimum swaps"},
		{Name: "swaps-max", Usage: "Maximum swaps"},
		{Name: "swaps-disabled", Usage: "Disable swaps"},
		{Name: "swap-fee", Usage: "Swap fee"},
		{Name: "slippage-fee-min", Usage: "Minimum slippage fee"},
		{Name: "limits-period", Usage: "Limits period"},
	}
	cli.AddTxFlags(proposalEditBasketCmd)
	proposalEditBasketCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		var basketID uint64
		fmt.Sscanf(ctx.GetFlag("basket-id"), "%d", &basketID)
		propOpts := &basket.ProposalEditBasketOpts{
			BasketID:          basketID,
			BasketSuffix:      ctx.GetFlag("basket-suffix"),
			BasketDescription: ctx.GetFlag("basket-description"),
			BasketTokens:      ctx.GetFlag("basket-tokens"),
			TokensCap:         ctx.GetFlag("tokens-cap"),
			Title:             ctx.GetFlag("title"),
			Description:       ctx.GetFlag("description"),
			MintsMin:          ctx.GetFlag("mints-min"),
			MintsMax:          ctx.GetFlag("mints-max"),
			MintsDisabled:     ctx.GetFlag("mints-disabled") == "true",
			BurnsMin:          ctx.GetFlag("burns-min"),
			BurnsMax:          ctx.GetFlag("burns-max"),
			BurnsDisabled:     ctx.GetFlag("burns-disabled") == "true",
			SwapsMin:          ctx.GetFlag("swaps-min"),
			SwapsMax:          ctx.GetFlag("swaps-max"),
			SwapsDisabled:     ctx.GetFlag("swaps-disabled") == "true",
			SwapFee:           ctx.GetFlag("swap-fee"),
			SlippageFeeMin:    ctx.GetFlag("slippage-fee-min"),
		}
		if ctx.GetFlag("limits-period") != "" {
			var limitsPeriod uint64
			fmt.Sscanf(ctx.GetFlag("limits-period"), "%d", &limitsPeriod)
			propOpts.LimitsPeriod = limitsPeriod
		}
		txOpts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.ProposalEditBasket(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(proposalEditBasketCmd)

	// proposal-basket-withdraw-surplus
	proposalWithdrawSurplusCmd := cli.NewCommand("proposal-basket-withdraw-surplus")
	proposalWithdrawSurplusCmd.Short = "Create proposal to withdraw surplus from baskets"
	proposalWithdrawSurplusCmd.Args = []cli.Arg{
		{Name: "basket-ids", Required: true, Description: "Comma-separated basket IDs"},
		{Name: "withdraw-target", Required: true, Description: "Target address for withdrawal"},
	}
	proposalWithdrawSurplusCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(proposalWithdrawSurplusCmd)
	proposalWithdrawSurplusCmd.Run = func(ctx *cli.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("basket-ids and withdraw-target required")
		}
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		basketMod := basket.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &basket.ProposalWithdrawSurplusOpts{
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		txOpts := &basket.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := basketMod.ProposalWithdrawSurplus(context.Background(), from, ctx.Args[0], ctx.Args[1], propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	basketTx.AddCommand(proposalWithdrawSurplusCmd)

	txCmd.AddCommand(basketTx)

	// Add collectives subcommand to tx
	collectivesTx := cli.NewCommand("collectives")
	collectivesTx.Short = "Collectives transaction commands"

	// create-collective
	createCollectiveCmd := cli.NewCommand("create-collective")
	createCollectiveCmd.Short = "Create a new collective"
	createCollectiveCmd.Flags = []cli.Flag{
		{Name: "collective-name", Usage: "Collective name", Required: true},
		{Name: "collective-description", Usage: "Collective description"},
	}
	cli.AddTxFlags(createCollectiveCmd)
	createCollectiveCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.CreateCollective(context.Background(), from, ctx.GetFlag("collective-name"), ctx.GetFlag("collective-description"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(createCollectiveCmd)

	// contribute-collective
	contributeCollectiveCmd := cli.NewCommand("contribute-collective")
	contributeCollectiveCmd.Short = "Contribute to a collective"
	contributeCollectiveCmd.Flags = []cli.Flag{
		{Name: "collective-name", Usage: "Collective name", Required: true},
		{Name: "bonds", Usage: "Bonds to contribute", Required: true},
	}
	cli.AddTxFlags(contributeCollectiveCmd)
	contributeCollectiveCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.ContributeCollective(context.Background(), from, ctx.GetFlag("collective-name"), ctx.GetFlag("bonds"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(contributeCollectiveCmd)

	// withdraw-collective
	withdrawCollectiveCmd := cli.NewCommand("withdraw-collective")
	withdrawCollectiveCmd.Short = "Withdraw from a collective"
	withdrawCollectiveCmd.Flags = []cli.Flag{
		{Name: "collective-name", Usage: "Collective name", Required: true},
		{Name: "bonds", Usage: "Bonds to withdraw", Required: true},
	}
	cli.AddTxFlags(withdrawCollectiveCmd)
	withdrawCollectiveCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.WithdrawCollective(context.Background(), from, ctx.GetFlag("collective-name"), ctx.GetFlag("bonds"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(withdrawCollectiveCmd)

	// donate-collective
	donateCollectiveCmd := cli.NewCommand("donate-collective")
	donateCollectiveCmd.Short = "Set lock and donation for bonds on a collective"
	donateCollectiveCmd.Flags = []cli.Flag{
		{Name: "collective-name", Usage: "Collective name", Required: true},
		{Name: "locking", Usage: "Lock period in seconds"},
		{Name: "donation", Usage: "Donation percentage"},
		{Name: "donation-lock", Usage: "Lock contribution on the collective"},
	}
	cli.AddTxFlags(donateCollectiveCmd)
	donateCollectiveCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		var locking uint64
		if ctx.GetFlag("locking") != "" {
			fmt.Sscanf(ctx.GetFlag("locking"), "%d", &locking)
		}
		donationLock := ctx.GetFlag("donation-lock") == "true"
		opts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.DonateCollective(context.Background(), from, ctx.GetFlag("collective-name"), locking, ctx.GetFlag("donation"), donationLock, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(donateCollectiveCmd)

	// proposal-collective-update
	proposalCollectiveUpdateCmd := cli.NewCommand("proposal-collective-update")
	proposalCollectiveUpdateCmd.Short = "Create proposal to update a collective"
	proposalCollectiveUpdateCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "collective-name", Usage: "Collective name", Required: true},
		{Name: "collective-description", Usage: "New collective description"},
		{Name: "collective-status", Usage: "Collective status (ACTIVE, PAUSED, etc.)"},
	}
	cli.AddTxFlags(proposalCollectiveUpdateCmd)
	proposalCollectiveUpdateCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &collectives.ProposalCollectiveUpdateOpts{
			Title:                 ctx.GetFlag("title"),
			Description:           ctx.GetFlag("description"),
			CollectiveName:        ctx.GetFlag("collective-name"),
			CollectiveDescription: ctx.GetFlag("collective-description"),
			CollectiveStatus:      ctx.GetFlag("collective-status"),
		}
		txOpts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.ProposalCollectiveUpdate(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(proposalCollectiveUpdateCmd)

	// proposal-remove-collective
	proposalRemoveCollectiveCmd := cli.NewCommand("proposal-remove-collective")
	proposalRemoveCollectiveCmd.Short = "Create proposal to remove a collective"
	proposalRemoveCollectiveCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "collective-name", Usage: "Collective name to remove", Required: true},
	}
	cli.AddTxFlags(proposalRemoveCollectiveCmd)
	proposalRemoveCollectiveCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &collectives.ProposalRemoveCollectiveOpts{
			Title:          ctx.GetFlag("title"),
			Description:    ctx.GetFlag("description"),
			CollectiveName: ctx.GetFlag("collective-name"),
		}
		txOpts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.ProposalRemoveCollective(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(proposalRemoveCollectiveCmd)

	// proposal-send-donation
	proposalSendDonationCmd := cli.NewCommand("proposal-send-donation")
	proposalSendDonationCmd.Short = "Create proposal to send donation from collective"
	proposalSendDonationCmd.Flags = []cli.Flag{
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
		{Name: "collective-name", Usage: "Collective name", Required: true},
		{Name: "addr", Usage: "Recipient address", Required: true},
		{Name: "amounts", Usage: "Amounts to send (e.g. 100ukex)", Required: true},
	}
	cli.AddTxFlags(proposalSendDonationCmd)
	proposalSendDonationCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		collectivesMod := collectives.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		propOpts := &collectives.ProposalSendDonationOpts{
			Title:          ctx.GetFlag("title"),
			Description:    ctx.GetFlag("description"),
			CollectiveName: ctx.GetFlag("collective-name"),
			Address:        ctx.GetFlag("addr"),
			Amounts:        ctx.GetFlag("amounts"),
		}
		txOpts := &collectives.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := collectivesMod.ProposalSendDonation(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	collectivesTx.AddCommand(proposalSendDonationCmd)

	txCmd.AddCommand(collectivesTx)

	// Add tokens subcommand to tx
	tokensTx := cli.NewCommand("tokens")
	tokensTx.Short = "Tokens transaction commands"

	// upsert-rate
	upsertRateCmd := cli.NewCommand("upsert-rate")
	upsertRateCmd.Short = "Upsert token rate"
	upsertRateCmd.Flags = []cli.Flag{
		{Name: "denom", Usage: "Token denomination", Required: true},
		{Name: "fee-rate", Usage: "Fee rate"},
		{Name: "fee-payments", Usage: "Allow fee payments"},
		{Name: "decimals", Usage: "Token decimals"},
		{Name: "description", Usage: "Token description"},
	}
	cli.AddTxFlags(upsertRateCmd)
	upsertRateCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &tokens.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		rateOpts := &tokens.UpsertRateOpts{
			Denom:       ctx.GetFlag("denom"),
			FeeRate:     ctx.GetFlag("fee-rate"),
			FeePayments: ctx.GetFlag("fee-payments") == "true",
			Description: ctx.GetFlag("description"),
		}
		if decimals := ctx.GetFlag("decimals"); decimals != "" {
			var d uint32
			fmt.Sscanf(decimals, "%d", &d)
			rateOpts.Decimals = d
		}
		resp, err := tokensMod.UpsertRate(context.Background(), from, rateOpts, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	tokensTx.AddCommand(upsertRateCmd)

	// proposal-upsert-rate
	propUpsertRateCmd := cli.NewCommand("proposal-upsert-rate")
	propUpsertRateCmd.Short = "Create proposal to upsert token rate"
	propUpsertRateCmd.Flags = []cli.Flag{
		{Name: "denom", Usage: "Token denomination", Required: true},
		{Name: "decimals", Usage: "Max decimal places"},
		{Name: "fee-rate", Usage: "Fee rate (max decimal 9, max value 10^10)"},
		{Name: "fee-payments", Usage: "Use registry as fee payment (bool)"},
		{Name: "name", Usage: "Token name (e.g. Kira, Bitcoin)"},
		{Name: "symbol", Usage: "Token symbol (e.g. KEX, BTC)"},
		{Name: "icon", Usage: "URL link to token icon"},
		{Name: "website", Usage: "Token website"},
		{Name: "social", Usage: "Social links"},
		{Name: "stake-token", Usage: "Flag if staking token (bool)"},
		{Name: "stake-cap", Usage: "Rewards allocation for token"},
		{Name: "stake-min", Usage: "Min amount to stake"},
		{Name: "supply", Usage: "Token supply"},
		{Name: "supply-cap", Usage: "Token supply cap"},
		{Name: "nft-hash", Usage: "NFT hash"},
		{Name: "nft-metadata", Usage: "NFT metadata"},
		{Name: "owner", Usage: "Token owner"},
		{Name: "owner-edit-disabled", Usage: "Disable owner editing (bool)"},
		{Name: "invalidated", Usage: "Mark token as invalidated (bool)"},
		{Name: "token-rate", Usage: "Token rate"},
		{Name: "token-type", Usage: "Token type"},
		{Name: "minting-fee", Usage: "Minting fee"},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propUpsertRateCmd)
	propUpsertRateCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		txOpts := &tokens.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		propOpts := &tokens.ProposalUpsertRateOpts{
			Denom:             ctx.GetFlag("denom"),
			FeeRate:           ctx.GetFlag("fee-rate"),
			FeePayments:       ctx.GetFlag("fee-payments") == "true",
			Name:              ctx.GetFlag("name"),
			Symbol:            ctx.GetFlag("symbol"),
			Icon:              ctx.GetFlag("icon"),
			Website:           ctx.GetFlag("website"),
			Social:            ctx.GetFlag("social"),
			StakeToken:        ctx.GetFlag("stake-token") == "true",
			StakeCap:          ctx.GetFlag("stake-cap"),
			StakeMin:          ctx.GetFlag("stake-min"),
			Supply:            ctx.GetFlag("supply"),
			SupplyCap:         ctx.GetFlag("supply-cap"),
			NftHash:           ctx.GetFlag("nft-hash"),
			NftMetadata:       ctx.GetFlag("nft-metadata"),
			Owner:             ctx.GetFlag("owner"),
			OwnerEditDisabled: ctx.GetFlag("owner-edit-disabled") == "true",
			Invalidated:       ctx.GetFlag("invalidated") == "true",
			TokenRate:         ctx.GetFlag("token-rate"),
			TokenType:         ctx.GetFlag("token-type"),
			MintingFee:        ctx.GetFlag("minting-fee"),
			Title:             ctx.GetFlag("title"),
			Description:       ctx.GetFlag("description"),
		}
		if decimals := ctx.GetFlag("decimals"); decimals != "" {
			var d uint32
			fmt.Sscanf(decimals, "%d", &d)
			propOpts.Decimals = d
		}
		resp, err := tokensMod.ProposalUpsertRate(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	tokensTx.AddCommand(propUpsertRateCmd)

	// proposal-update-tokens-blackwhite
	propUpdateTokensBWCmd := cli.NewCommand("proposal-update-tokens-blackwhite")
	propUpdateTokensBWCmd.Short = "Create proposal to update token blacklist/whitelist"
	propUpdateTokensBWCmd.Flags = []cli.Flag{
		{Name: "tokens", Usage: "Comma-separated list of tokens", Required: true},
		{Name: "is-blacklist", Usage: "True to modify blacklist (bool)"},
		{Name: "is-add", Usage: "True to add tokens, false to remove (bool)"},
		{Name: "title", Usage: "Proposal title", Required: true},
		{Name: "description", Usage: "Proposal description", Required: true},
	}
	cli.AddTxFlags(propUpdateTokensBWCmd)
	propUpdateTokensBWCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		tokensMod := tokens.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		txOpts := &tokens.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		// Parse tokens as comma-separated list
		tokensStr := ctx.GetFlag("tokens")
		var tokensList []string
		if tokensStr != "" {
			for _, t := range strings.Split(tokensStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tokensList = append(tokensList, t)
				}
			}
		}
		propOpts := &tokens.ProposalUpdateTokensBlackWhiteOpts{
			Tokens:      tokensList,
			IsBlacklist: ctx.GetFlag("is-blacklist") == "true",
			IsAdd:       ctx.GetFlag("is-add") == "true",
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		resp, err := tokensMod.ProposalUpdateTokensBlackWhite(context.Background(), from, propOpts, txOpts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	tokensTx.AddCommand(propUpdateTokensBWCmd)

	txCmd.AddCommand(tokensTx)

	// Add ubi subcommand to tx
	ubiTx := cli.NewCommand("ubi")
	ubiTx.Short = "UBI transaction commands"

	// proposal-upsert-ubi
	propUpsertUBICmd := cli.NewCommand("proposal-upsert-ubi")
	propUpsertUBICmd.Short = "Create proposal to upsert UBI record"
	propUpsertUBICmd.Flags = []cli.Flag{
		{Name: "name", Usage: "UBI name", Required: true},
		{Name: "distr-start", Usage: "Distribution start time (unix timestamp)"},
		{Name: "distr-end", Usage: "Distribution end time (unix timestamp)"},
		{Name: "amount", Usage: "Amount"},
		{Name: "period", Usage: "Period in seconds"},
		{Name: "pool", Usage: "Pool name"},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propUpsertUBICmd)
	propUpsertUBICmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		ubiMod := ubi.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &ubi.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		ubiOpts := &ubi.ProposalUpsertUBIOpts{
			Name:        ctx.GetFlag("name"),
			PoolName:    ctx.GetFlag("pool"),
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		if ds := ctx.GetFlag("distr-start"); ds != "" {
			fmt.Sscanf(ds, "%d", &ubiOpts.DistrStart)
		}
		if de := ctx.GetFlag("distr-end"); de != "" {
			fmt.Sscanf(de, "%d", &ubiOpts.DistrEnd)
		}
		if amt := ctx.GetFlag("amount"); amt != "" {
			fmt.Sscanf(amt, "%d", &ubiOpts.Amount)
		}
		if per := ctx.GetFlag("period"); per != "" {
			fmt.Sscanf(per, "%d", &ubiOpts.Period)
		}
		resp, err := ubiMod.ProposalUpsertUBI(context.Background(), from, ubiOpts, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	ubiTx.AddCommand(propUpsertUBICmd)

	// proposal-remove-ubi
	propRemoveUBICmd := cli.NewCommand("proposal-remove-ubi")
	propRemoveUBICmd.Short = "Create proposal to remove UBI record"
	propRemoveUBICmd.Flags = []cli.Flag{
		{Name: "name", Usage: "UBI name", Required: true},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propRemoveUBICmd)
	propRemoveUBICmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		ubiMod := ubi.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &ubi.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		propOpts := &ubi.ProposalRemoveUBIOpts{
			Name:        ctx.GetFlag("name"),
			Title:       ctx.GetFlag("title"),
			Description: ctx.GetFlag("description"),
		}
		resp, err := ubiMod.ProposalRemoveUBI(context.Background(), from, propOpts, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	ubiTx.AddCommand(propRemoveUBICmd)
	txCmd.AddCommand(ubiTx)

	// Add upgrade subcommand to tx
	upgradeTx := cli.NewCommand("upgrade")
	upgradeTx.Short = "Upgrade transaction commands"

	// proposal-set-plan
	propSetPlanCmd := cli.NewCommand("proposal-set-plan")
	propSetPlanCmd.Short = "Create proposal to set upgrade plan"
	propSetPlanCmd.Flags = []cli.Flag{
		{Name: "name", Usage: "Plan name", Required: true},
		{Name: "min-upgrade-time", Usage: "Minimum upgrade time (unix timestamp)"},
		{Name: "old-chain-id", Usage: "Old chain ID"},
		{Name: "new-chain-id", Usage: "New chain ID"},
		{Name: "max-enrollment-duration", Usage: "Max enrollment duration"},
		{Name: "instate-upgrade", Usage: "In-state upgrade flag"},
		{Name: "reboot-required", Usage: "Reboot required flag"},
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propSetPlanCmd)
	propSetPlanCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		upgradeMod := upgrade.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &upgrade.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		planOpts := &upgrade.ProposalSetPlanOpts{
			Name:           ctx.GetFlag("name"),
			OldChainID:     ctx.GetFlag("old-chain-id"),
			NewChainID:     ctx.GetFlag("new-chain-id"),
			InstateUpgrade: ctx.GetFlag("instate-upgrade") == "true",
			RebootRequired: ctx.GetFlag("reboot-required") == "true",
			Description:    ctx.GetFlag("description"),
		}
		if mut := ctx.GetFlag("min-upgrade-time"); mut != "" {
			fmt.Sscanf(mut, "%d", &planOpts.MinUpgradeTime)
		}
		if med := ctx.GetFlag("max-enrollment-duration"); med != "" {
			fmt.Sscanf(med, "%d", &planOpts.MaxEnrollmentDuration)
		}
		resp, err := upgradeMod.ProposalSetPlan(context.Background(), from, planOpts, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	upgradeTx.AddCommand(propSetPlanCmd)

	// proposal-cancel-plan
	propCancelPlanCmd := cli.NewCommand("proposal-cancel-plan")
	propCancelPlanCmd.Short = "Create proposal to cancel upgrade plan"
	propCancelPlanCmd.Flags = []cli.Flag{
		{Name: "description", Usage: "Proposal description"},
	}
	cli.AddTxFlags(propCancelPlanCmd)
	propCancelPlanCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		upgradeMod := upgrade.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &upgrade.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		propOpts := &upgrade.ProposalCancelPlanOpts{
			Description: ctx.GetFlag("description"),
		}
		resp, err := upgradeMod.ProposalCancelPlan(context.Background(), from, propOpts, opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	upgradeTx.AddCommand(propCancelPlanCmd)
	txCmd.AddCommand(upgradeTx)

	// Add bridge subcommand to tx
	bridgeTx := cli.NewCommand("bridge")
	bridgeTx.Short = "Bridge transaction commands"

	// change-cosmos-ethereum
	changeCosmosEthCmd := cli.NewCommand("change-cosmos-ethereum")
	changeCosmosEthCmd.Short = "Create change request from Cosmos to Ethereum"
	changeCosmosEthCmd.Flags = []cli.Flag{
		{Name: "cosmos-address", Usage: "Cosmos address", Required: true},
		{Name: "eth-address", Usage: "Ethereum address", Required: true},
		{Name: "amount", Usage: "Amount to bridge", Required: true},
	}
	cli.AddTxFlags(changeCosmosEthCmd)
	changeCosmosEthCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bridgeMod := bridge.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &bridge.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := bridgeMod.ChangeCosmosEthereum(context.Background(), from, ctx.GetFlag("cosmos-address"), ctx.GetFlag("eth-address"), ctx.GetFlag("amount"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	bridgeTx.AddCommand(changeCosmosEthCmd)

	// change-ethereum-cosmos
	changeEthCosmosCmd := cli.NewCommand("change-ethereum-cosmos")
	changeEthCosmosCmd.Short = "Create change request from Ethereum to Cosmos"
	changeEthCosmosCmd.Flags = []cli.Flag{
		{Name: "cosmos-address", Usage: "Cosmos address", Required: true},
		{Name: "eth-tx-hash", Usage: "Ethereum transaction hash", Required: true},
		{Name: "amount", Usage: "Amount to bridge", Required: true},
	}
	cli.AddTxFlags(changeEthCosmosCmd)
	changeEthCosmosCmd.Run = func(ctx *cli.Context) error {
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}
		bridgeMod := bridge.New(client)
		from := a.getFromFlag(ctx)
		if from == "" {
			return fmt.Errorf("--from flag required (run 'sekai-cli init' to set default)")
		}
		opts := &bridge.TxOptions{
			Fees:          ctx.GetFlag("fees"),
			Gas:           ctx.GetFlag("gas"),
			Memo:          ctx.GetFlag("memo"),
			BroadcastMode: ctx.GetFlag("broadcast-mode"),
		}
		resp, err := bridgeMod.ChangeEthereumCosmos(context.Background(), from, ctx.GetFlag("cosmos-address"), ctx.GetFlag("eth-tx-hash"), ctx.GetFlag("amount"), opts)
		if err != nil {
			return err
		}
		return a.printOutput(ctx, resp)
	}
	bridgeTx.AddCommand(changeEthCosmosCmd)
	txCmd.AddCommand(bridgeTx)

	return txCmd
}

// buildVersionCommand builds the version command.
func (a *App) buildVersionCommand() *cli.Command {
	cmd := cli.NewCommand("version")
	cmd.Short = "Print version information"

	cmd.Run = func(ctx *cli.Context) error {
		ctx.Printf("sekai-cli version %s\n", sdk.Version())
		return nil
	}

	return cmd
}

// buildCompletionCommand builds the completion command for shell completions.
func (a *App) buildCompletionCommand() *cli.Command {
	cmd := cli.NewCommand("completion")
	cmd.Short = "Generate shell completion scripts"
	cmd.Long = `Generate shell completion scripts for sekai-cli.

To load completions:

Bash:
  $ sekai-cli completion bash | sudo tee /etc/bash_completion.d/sekai-cli
  # or for current session:
  $ source <(sekai-cli completion bash)

Zsh:
  $ sekai-cli completion zsh > "${fpath[1]}/_sekai-cli"
  # or
  $ echo 'source <(sekai-cli completion zsh)' >> ~/.zshrc

Fish:
  $ sekai-cli completion fish > ~/.config/fish/completions/sekai-cli.fish`

	// Bash subcommand
	bashCmd := cli.NewCommand("bash")
	bashCmd.Short = "Generate bash completion script"
	bashCmd.Run = func(ctx *cli.Context) error {
		ctx.Printf("%s", cli.GenerateBashCompletion(a.root))
		return nil
	}

	// Zsh subcommand
	zshCmd := cli.NewCommand("zsh")
	zshCmd.Short = "Generate zsh completion script"
	zshCmd.Run = func(ctx *cli.Context) error {
		ctx.Printf("%s", cli.GenerateZshCompletion(a.root))
		return nil
	}

	// Fish subcommand
	fishCmd := cli.NewCommand("fish")
	fishCmd.Short = "Generate fish completion script"
	fishCmd.Run = func(ctx *cli.Context) error {
		ctx.Printf("%s", cli.GenerateFishCompletion(a.root))
		return nil
	}

	cmd.AddCommands(bashCmd, zshCmd, fishCmd)
	return cmd
}

// buildInitCommand builds the init command for first-time setup.
func (a *App) buildInitCommand() *cli.Command {
	cmd := cli.NewCommand("init")
	cmd.Short = "Initialize CLI by detecting container and caching network config"
	cmd.Long = `Initialize sekai-cli by auto-detecting the running sekaid container,
querying network properties, and caching keys. This eliminates the need to
specify --chain-id, --fees, and --from flags for every command.

Run 'sekai-cli sync' to refresh the cache after network changes.`

	cmd.AddFlag(cli.Flag{Name: "container", Usage: "Container name (auto-detects if not provided)"})
	cmd.AddFlag(cli.Flag{Name: "default-key", Usage: "Set default signing key"})
	cmd.AddFlag(cli.Flag{Name: "force", Short: "f", Usage: "Overwrite existing cache"})

	cmd.Run = func(ctx *cli.Context) error {
		// Check if cache already exists
		if cache.Exists() && ctx.GetFlag("force") == "" {
			ctx.Printf("Cache already exists at %s\n", cache.DefaultCachePath())
			ctx.Printf("Use --force to overwrite or 'sekai-cli sync' to refresh.\n")
			return nil
		}

		// Get container
		container := ctx.GetFlag("container")
		if container == "" {
			container = a.config.Container
		}

		// Try to auto-detect if not specified
		if container == "" || container == "sekai-node" {
			ctx.Printf("Detecting container...\n")
			detected, err := docker.FindSekaiContainer()
			if err != nil {
				return fmt.Errorf("no container specified and auto-detection failed: %w", err)
			}
			container = detected
			ctx.Printf("Found container: %s\n", container)
		}

		// Verify container is running
		if !docker.IsContainerRunning(container) {
			return fmt.Errorf("container '%s' is not running", container)
		}

		// Create docker client
		home := a.config.Home
		if home == "" || home == "/.sekaid" {
			home = "/sekai" // Default for sekai containers
		}
		keyringBackend := a.config.KeyringBackend
		if keyringBackend == "" {
			keyringBackend = "test"
		}
		client, err := docker.NewClient(container,
			docker.WithKeyringBackend(keyringBackend),
			docker.WithHome(home),
		)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer client.Close()

		// Query status for chain ID
		ctx.Printf("Querying node status...\n")
		statusResp, err := client.Status(context.Background())
		if err != nil {
			return fmt.Errorf("failed to query status: %w", err)
		}

		// Query network properties
		ctx.Printf("Querying network properties...\n")
		govMod := gov.New(client)
		props, err := govMod.NetworkProperties(context.Background())
		if err != nil {
			return fmt.Errorf("failed to query network properties: %w", err)
		}

		// Query keys
		ctx.Printf("Querying keys...\n")
		keysList, err := client.Keys().List(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list keys: %w", err)
		}

		// Build cache
		c := cache.New()
		c.Container = container
		c.Network = cache.NetworkCache{
			ChainID:                  statusResp.NodeInfo.Network,
			Moniker:                  statusResp.NodeInfo.Moniker,
			MinTxFee:                 props.MinTxFee,
			MaxTxFee:                 props.MaxTxFee,
			VoteQuorum:               props.VoteQuorum,
			MinimumProposalEndTime:   props.MinimumProposalEndTime,
			ProposalEnactmentTime:    props.ProposalEnactmentTime,
			EnableForeignFeePayments: props.EnableForeignFeePayments,
			MinValidators:            props.MinValidators,
			PoorNetworkMaxBankSend:   props.PoorNetworkMaxBankSend,
			UnjailMaxTime:            props.UnjailMaxTime,
			UnstakingPeriod:          props.UnstakingPeriod,
			MaxDelegators:            props.MaxDelegators,
		}

		// Cache keys
		for _, k := range keysList {
			c.Keys = append(c.Keys, cache.KeyCache{
				Name:    k.Name,
				Address: k.Address,
				Type:    k.Type,
			})
		}

		// Set default key
		defaultKey := ctx.GetFlag("default-key")
		if defaultKey != "" {
			if err := c.SetDefaultKey(defaultKey); err != nil {
				return err
			}
		} else if len(c.Keys) > 0 {
			c.DefaultKey = c.Keys[0].Name
		}

		// Save cache
		if err := c.Save(); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		// Also update config with detected values
		a.config.Container = container
		a.config.ChainID = c.Network.ChainID
		a.config.Home = "/sekai" // Always use /sekai to avoid /.sekaid ghost directory
		if err := a.config.Save(config.DefaultConfigPath()); err != nil {
			ctx.Printf("Warning: failed to save config: %v\n", err)
		}

		// Print summary
		ctx.Printf("\nInitialization complete!\n")
		ctx.Printf("Container:   %s\n", c.Container)
		ctx.Printf("Chain ID:    %s\n", c.Network.ChainID)
		ctx.Printf("Min TX Fee:  %s\n", c.Network.MinTxFee)
		ctx.Printf("Keys found:  %d\n", len(c.Keys))
		if c.DefaultKey != "" {
			ctx.Printf("Default key: %s\n", c.DefaultKey)
		}
		ctx.Printf("\nCache saved to %s\n", cache.DefaultCachePath())
		return nil
	}

	return cmd
}

// buildSyncCommand builds the sync command to refresh cache.
func (a *App) buildSyncCommand() *cli.Command {
	cmd := cli.NewCommand("sync")
	cmd.Short = "Refresh cached network config and keys"
	cmd.Long = `Refresh the cached network configuration and keys from the running container.
Use this after network properties have been changed via governance proposals.`

	cmd.AddFlag(cli.Flag{Name: "keys-only", Usage: "Only refresh keys"})
	cmd.AddFlag(cli.Flag{Name: "network-only", Usage: "Only refresh network properties"})

	cmd.Run = func(ctx *cli.Context) error {
		// Load existing cache
		c, err := cache.Load()
		if err != nil {
			return fmt.Errorf("no cache found: %w\nRun 'sekai-cli init' first", err)
		}

		keysOnly := ctx.GetFlag("keys-only") != ""
		networkOnly := ctx.GetFlag("network-only") != ""

		// Verify container is still running
		if !docker.IsContainerRunning(c.Container) {
			return fmt.Errorf("container '%s' is not running", c.Container)
		}

		// Create docker client
		home := a.config.Home
		if home == "" || home == "/.sekaid" {
			home = "/sekai" // Default for sekai containers
		}
		keyringBackend := a.config.KeyringBackend
		if keyringBackend == "" {
			keyringBackend = "test"
		}
		client, err := docker.NewClient(c.Container,
			docker.WithKeyringBackend(keyringBackend),
			docker.WithHome(home),
		)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer client.Close()

		ctx.Printf("Syncing from container: %s\n", c.Container)

		// Sync network properties
		if !keysOnly {
			ctx.Printf("Refreshing network properties...\n")
			statusResp, err := client.Status(context.Background())
			if err != nil {
				return fmt.Errorf("failed to query status: %w", err)
			}

			govMod := gov.New(client)
			props, err := govMod.NetworkProperties(context.Background())
			if err != nil {
				return fmt.Errorf("failed to query network properties: %w", err)
			}

			// Track changes
			oldMinFee := c.Network.MinTxFee
			oldMaxFee := c.Network.MaxTxFee

			c.Network = cache.NetworkCache{
				ChainID:                  statusResp.NodeInfo.Network,
				Moniker:                  statusResp.NodeInfo.Moniker,
				MinTxFee:                 props.MinTxFee,
				MaxTxFee:                 props.MaxTxFee,
				VoteQuorum:               props.VoteQuorum,
				MinimumProposalEndTime:   props.MinimumProposalEndTime,
				ProposalEnactmentTime:    props.ProposalEnactmentTime,
				EnableForeignFeePayments: props.EnableForeignFeePayments,
				MinValidators:            props.MinValidators,
				PoorNetworkMaxBankSend:   props.PoorNetworkMaxBankSend,
				UnjailMaxTime:            props.UnjailMaxTime,
				UnstakingPeriod:          props.UnstakingPeriod,
				MaxDelegators:            props.MaxDelegators,
			}

			// Report changes
			if oldMinFee != c.Network.MinTxFee {
				ctx.Printf("  min_tx_fee: %s -> %s\n", oldMinFee, c.Network.MinTxFee)
			}
			if oldMaxFee != c.Network.MaxTxFee {
				ctx.Printf("  max_tx_fee: %s -> %s\n", oldMaxFee, c.Network.MaxTxFee)
			}
		}

		// Sync keys
		if !networkOnly {
			ctx.Printf("Refreshing keys...\n")
			keysList, err := client.Keys().List(context.Background())
			if err != nil {
				return fmt.Errorf("failed to list keys: %w", err)
			}

			oldKeyCount := len(c.Keys)
			c.Keys = nil
			for _, k := range keysList {
				c.Keys = append(c.Keys, cache.KeyCache{
					Name:    k.Name,
					Address: k.Address,
					Type:    k.Type,
				})
			}

			if len(c.Keys) != oldKeyCount {
				ctx.Printf("  keys: %d -> %d\n", oldKeyCount, len(c.Keys))
			}

			// Verify default key still exists
			if c.DefaultKey != "" && c.GetKeyByName(c.DefaultKey) == nil {
				ctx.Printf("  Warning: default key '%s' no longer exists\n", c.DefaultKey)
				if len(c.Keys) > 0 {
					c.DefaultKey = c.Keys[0].Name
					ctx.Printf("  New default key: %s\n", c.DefaultKey)
				} else {
					c.DefaultKey = ""
				}
			}
		}

		// Save updated cache
		if err := c.Save(); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		// Also update config with synced values
		a.config.Container = c.Container
		if c.Network.ChainID != "" {
			a.config.ChainID = c.Network.ChainID
		}
		a.config.Home = "/sekai" // Always use /sekai to avoid /.sekaid ghost directory
		if err := a.config.Save(config.DefaultConfigPath()); err != nil {
			ctx.Printf("Warning: failed to save config: %v\n", err)
		}

		ctx.Printf("Cache updated.\n")
		return nil
	}

	return cmd
}

// buildCacheCommand builds the cache command group.
func (a *App) buildCacheCommand() *cli.Command {
	cacheCmd := cli.NewCommand("cache")
	cacheCmd.Short = "Cache management commands"

	// cache show
	showCmd := cli.NewCommand("show")
	showCmd.Short = "Show cached configuration"
	showCmd.Run = func(ctx *cli.Context) error {
		c, err := cache.Load()
		if err != nil {
			return err
		}
		ctx.Printf("%s", c.Summary())
		return nil
	}
	cacheCmd.AddCommand(showCmd)

	// cache clear
	clearCmd := cli.NewCommand("clear")
	clearCmd.Short = "Clear cached configuration"
	clearCmd.Run = func(ctx *cli.Context) error {
		if !cache.Exists() {
			ctx.Printf("No cache file exists.\n")
			return nil
		}
		if err := cache.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		ctx.Printf("Cache cleared.\n")
		return nil
	}
	cacheCmd.AddCommand(clearCmd)

	// cache set-default-key
	setKeyCmd := cli.NewCommand("set-default-key")
	setKeyCmd.Short = "Set the default signing key"
	setKeyCmd.Args = []cli.Arg{{Name: "key-name", Required: true, Description: "Name of the key to use as default"}}
	setKeyCmd.Run = func(ctx *cli.Context) error {
		keyName := ctx.GetArg(0)
		if keyName == "" {
			return fmt.Errorf("key name required")
		}

		c, err := cache.Load()
		if err != nil {
			return err
		}

		if err := c.SetDefaultKey(keyName); err != nil {
			return err
		}

		if err := c.Save(); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		ctx.Printf("Default key set to: %s\n", keyName)
		return nil
	}
	cacheCmd.AddCommand(setKeyCmd)

	return cacheCmd
}

func (a *App) buildConfigCommand() *cli.Command {
	configCmd := cli.NewCommand("config")
	configCmd.Short = "Configuration commands"

	// config show
	showCmd := cli.NewCommand("show")
	showCmd.Short = "Show current configuration"
	showCmd.Run = func(ctx *cli.Context) error {
		return a.printOutput(ctx, a.config)
	}
	configCmd.AddCommand(showCmd)

	// config init
	initCmd := cli.NewCommand("init")
	initCmd.Short = "Initialize configuration file"
	initCmd.Run = func(ctx *cli.Context) error {
		path := config.DefaultConfigPath()
		if err := a.config.Save(path); err != nil {
			return err
		}
		ctx.Printf("Configuration saved to %s\n", path)
		return nil
	}
	configCmd.AddCommand(initCmd)

	return configCmd
}

// buildScenarioCommand creates the scenario command for running playbooks.
func (a *App) buildScenarioCommand() *cli.Command {
	scenarioCmd := cli.NewCommand("scenario")
	scenarioCmd.Short = "Execute scenario playbooks"
	scenarioCmd.Long = `Execute scenario playbooks (YAML files) that define sequences of blockchain operations.

Scenarios allow you to automate complex workflows like:
  - Setting up validators
  - Creating and voting on proposals
  - Managing roles and permissions
  - Token transfers and staking

Example scenario file:
  name: transfer-tokens
  variables:
    amount: 1000ukex
  steps:
    - name: Send tokens
      module: bank
      action: send
      params:
        from: alice
        to: bob
        amount: "{{ amount }}"`

	// run subcommand
	runCmd := cli.NewCommand("run")
	runCmd.Short = "Execute a scenario file"
	runCmd.Args = []cli.Arg{
		{Name: "file", Required: true, Description: "Path to the scenario YAML file"},
	}
	runCmd.Flags = []cli.Flag{
		{Name: "var", Usage: "Override variable (can be repeated): --var key=value"},
		{Name: "dry-run", Usage: "Show what would be executed without running"},
		{Name: "verbose", Usage: "Show detailed output"},
		{Name: "continue-on-error", Usage: "Continue executing even if a step fails"},
		{Name: "tx-timeout", Usage: "Timeout for transaction confirmation (default: 60s)"},
	}
	cli.AddGlobalFlags(runCmd)
	runCmd.Run = func(ctx *cli.Context) error {
		filePath := ctx.Args[0]

		// Load scenario
		scenario, err := scenarios.LoadFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load scenario: %w", err)
		}

		// Get client
		client, err := a.getClient(ctx)
		if err != nil {
			return err
		}

		// Parse CLI variable overrides
		varOverrides := make(map[string]string)
		if varFlag := ctx.GetFlag("var"); varFlag != "" {
			// Handle multiple --var flags
			vars := strings.Split(varFlag, ",")
			for _, v := range vars {
				parts := strings.SplitN(v, "=", 2)
				if len(parts) == 2 {
					varOverrides[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		// Build executor options
		opts := scenarios.DefaultExecutorOptions()
		opts.DryRun = ctx.GetFlag("dry-run") == "true"
		opts.Verbose = ctx.GetFlag("verbose") == "true"
		opts.ContinueOnError = ctx.GetFlag("continue-on-error") == "true"
		opts.Variables = varOverrides

		if timeout := ctx.GetFlag("tx-timeout"); timeout != "" {
			// Parse duration
			if d, err := parseDuration(timeout); err == nil {
				opts.TxWaitTimeout = d
			}
		}

		// Create executor and run
		executor := scenarios.NewExecutor(client, opts)
		result, err := executor.Execute(context.Background(), scenario)
		if err != nil {
			return err
		}

		// Output result
		if !result.Success {
			return fmt.Errorf("scenario failed: %s", result.Error)
		}

		return nil
	}
	scenarioCmd.AddCommand(runCmd)

	// validate subcommand
	validateCmd := cli.NewCommand("validate")
	validateCmd.Short = "Validate a scenario file without executing"
	validateCmd.Args = []cli.Arg{
		{Name: "file", Required: true, Description: "Path to the scenario YAML file"},
	}
	validateCmd.Run = func(ctx *cli.Context) error {
		filePath := ctx.Args[0]

		scenario, err := scenarios.LoadFromFile(filePath)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		fmt.Println(scenario.String())
		fmt.Println("Scenario is valid!")
		return nil
	}
	scenarioCmd.AddCommand(validateCmd)

	// show subcommand
	showCmd := cli.NewCommand("show")
	showCmd.Short = "Show scenario details"
	showCmd.Args = []cli.Arg{
		{Name: "file", Required: true, Description: "Path to the scenario YAML file"},
	}
	showCmd.Run = func(ctx *cli.Context) error {
		filePath := ctx.Args[0]

		scenario, err := scenarios.LoadFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load scenario: %w", err)
		}

		fmt.Println(scenario.String())
		return nil
	}
	scenarioCmd.AddCommand(showCmd)

	return scenarioCmd
}

// parseDuration parses a duration string like "60s", "5m", etc.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// Helper functions

func getStringOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

// Close closes the application and releases resources.
func (a *App) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}

// Version information (set by build flags).
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// RunCLI is the main entry point for the CLI.
func RunCLI() {
	cfg := config.Load()

	app, err := New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
