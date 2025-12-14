// Package cli provides a lightweight CLI framework for sekai-cli.
// It is designed to have zero external dependencies.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Command represents a CLI command or subcommand.
type Command struct {
	// Name is the command name as used on the command line.
	Name string

	// Aliases are alternative names for the command.
	Aliases []string

	// Short is a short description shown in help.
	Short string

	// Long is a longer description shown in command help.
	Long string

	// Usage provides usage examples.
	Usage string

	// Args describes the expected arguments.
	Args []Arg

	// Flags are the command's flags.
	Flags []Flag

	// Run is the function to execute when this command is called.
	// If nil, help is shown when the command is invoked without subcommands.
	Run RunFunc

	// SubCommands are nested commands.
	SubCommands []*Command

	// Hidden hides the command from help output.
	Hidden bool

	// parent is the parent command (set automatically).
	parent *Command
}

// Arg describes a command argument.
type Arg struct {
	// Name is the argument name for display purposes.
	Name string

	// Required indicates if the argument is required.
	Required bool

	// Description is the argument description.
	Description string
}

// Flag represents a command-line flag.
type Flag struct {
	// Name is the long flag name (without --).
	Name string

	// Short is the short flag name (single character, without -).
	Short string

	// Default is the default value.
	Default string

	// Required indicates if the flag is required.
	Required bool

	// Usage is the flag description.
	Usage string

	// Hidden hides the flag from help output.
	Hidden bool
}

// RunFunc is the function signature for command execution.
type RunFunc func(ctx *Context) error

// Context provides context for command execution.
type Context struct {
	// Command is the command being executed.
	Command *Command

	// Args are the positional arguments.
	Args []string

	// Flags are the parsed flag values.
	Flags map[string]string

	// Stdin is the standard input.
	Stdin io.Reader

	// Stdout is the standard output.
	Stdout io.Writer

	// Stderr is the standard error.
	Stderr io.Writer

	// parent context for accessing parent command data.
	parent *Context
}

// NewCommand creates a new command with the given name.
func NewCommand(name string) *Command {
	return &Command{
		Name:        name,
		SubCommands: make([]*Command, 0),
		Flags:       make([]Flag, 0),
		Args:        make([]Arg, 0),
	}
}

// AddCommand adds a subcommand.
func (c *Command) AddCommand(sub *Command) {
	sub.parent = c
	c.SubCommands = append(c.SubCommands, sub)
}

// AddCommands adds multiple subcommands.
func (c *Command) AddCommands(subs ...*Command) {
	for _, sub := range subs {
		c.AddCommand(sub)
	}
}

// AddFlag adds a flag to the command.
func (c *Command) AddFlag(f Flag) {
	c.Flags = append(c.Flags, f)
}

// AddArg adds an argument to the command.
func (c *Command) AddArg(a Arg) {
	c.Args = append(c.Args, a)
}

// Execute parses arguments and executes the command.
func (c *Command) Execute(args []string) error {
	return c.ExecuteContext(&Context{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}, args)
}

// ExecuteContext parses arguments and executes with the given context.
func (c *Command) ExecuteContext(ctx *Context, args []string) error {
	ctx.Command = c
	ctx.Flags = make(map[string]string)

	// Set defaults for all flags
	for _, f := range c.Flags {
		if f.Default != "" {
			ctx.Flags[f.Name] = f.Default
		}
	}

	// Check for subcommand FIRST (before parsing flags)
	// This ensures "keys --help" goes to keys subcommand, not root
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		subName := args[0]

		// Check for help as first arg
		if subName == "help" {
			if len(args) > 1 {
				// Help for specific subcommand
				for _, sub := range c.SubCommands {
					if sub.Name == args[1] || contains(sub.Aliases, args[1]) {
						return sub.showHelp(ctx)
					}
				}
			}
			return c.showHelp(ctx)
		}

		// Look for matching subcommand
		for _, sub := range c.SubCommands {
			if sub.Name == subName || contains(sub.Aliases, subName) {
				// Parse parent flags first (only flags before subcommand)
				// We'll pass remaining args to subcommand
				subCtx := &Context{
					Stdin:  ctx.Stdin,
					Stdout: ctx.Stdout,
					Stderr: ctx.Stderr,
					parent: ctx,
				}
				return sub.ExecuteContext(subCtx, args[1:])
			}
		}
	}

	// Parse arguments (no subcommand matched or args start with flags)
	remaining, err := c.parseArgs(ctx, args)
	if err != nil {
		return err
	}

	// Check for help flag
	if ctx.Flags["help"] == "true" {
		return c.showHelp(ctx)
	}

	// Check again for subcommand in remaining args (after flags)
	if len(remaining) > 0 {
		subName := remaining[0]

		for _, sub := range c.SubCommands {
			if sub.Name == subName || contains(sub.Aliases, subName) {
				subCtx := &Context{
					Stdin:  ctx.Stdin,
					Stdout: ctx.Stdout,
					Stderr: ctx.Stderr,
					parent: ctx,
				}
				return sub.ExecuteContext(subCtx, remaining[1:])
			}
		}
	}

	// Set remaining args
	ctx.Args = remaining

	// Validate required flags
	for _, f := range c.Flags {
		if f.Required && ctx.Flags[f.Name] == "" {
			return fmt.Errorf("required flag --%s not provided", f.Name)
		}
	}

	// Validate required args
	for i, a := range c.Args {
		if a.Required && (i >= len(ctx.Args) || ctx.Args[i] == "") {
			return fmt.Errorf("required argument <%s> not provided", a.Name)
		}
	}

	// Execute command
	if c.Run != nil {
		return c.Run(ctx)
	}

	// No run function and no subcommand matched - show help
	return c.showHelp(ctx)
}

// parseArgs parses flags and returns remaining positional arguments.
// Flags can appear anywhere (before or after positional args).
func (c *Command) parseArgs(ctx *Context, args []string) ([]string, error) {
	var remaining []string
	skipNext := false

	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}

		arg := args[i]

		// Check for -- (end of flags)
		if arg == "--" {
			remaining = append(remaining, args[i+1:]...)
			break
		}

		// Note: We continue parsing flags even after positional args
		// This allows "cmd arg --flag value" syntax
		// Removed: seenPositional check that stopped flag parsing

		// Long flag
		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			value := "true"

			// Check for = syntax
			if idx := strings.Index(name, "="); idx >= 0 {
				value = name[idx+1:]
				name = name[:idx]
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Next arg is the value (for non-boolean flags)
				if !c.isBoolFlag(name) {
					value = args[i+1]
					skipNext = true
				}
			}

			if !c.hasFlag(name) && name != "help" {
				return nil, fmt.Errorf("unknown flag: --%s", name)
			}

			ctx.Flags[name] = value
			continue
		}

		// Short flag
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]

			// Handle combined short flags like -abc
			for j, r := range short {
				name := c.longNameForShort(string(r))
				if name == "" {
					// Check if it's "h" for help
					if string(r) == "h" {
						ctx.Flags["help"] = "true"
						continue
					}
					return nil, fmt.Errorf("unknown flag: -%s", string(r))
				}

				value := "true"
				// If last char and next arg exists and doesn't start with -, use it as value
				if j == len(short)-1 && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					if !c.isBoolFlag(name) {
						value = args[i+1]
						skipNext = true
					}
				}

				ctx.Flags[name] = value
			}
			continue
		}

		// Positional argument
		remaining = append(remaining, arg)
	}

	return remaining, nil
}

// hasFlag checks if the command has a flag with the given name.
func (c *Command) hasFlag(name string) bool {
	for _, f := range c.Flags {
		if f.Name == name {
			return true
		}
	}
	return false
}

// isBoolFlag checks if a flag is boolean (doesn't take a value).
func (c *Command) isBoolFlag(name string) bool {
	// Known boolean flags (don't take values)
	boolFlags := map[string]bool{
		"help":    true,
		"force":   true,
		"yes":     true,
		"recover": true,
	}
	if boolFlags[name] {
		return true
	}

	// Flags with non-boolean defaults are not boolean
	for _, f := range c.Flags {
		if f.Name == name {
			if f.Default != "" && f.Default != "true" && f.Default != "false" {
				return false
			}
			// If no default, treat as non-boolean (takes value)
			return false
		}
	}
	return false
}

// longNameForShort returns the long flag name for a short flag.
func (c *Command) longNameForShort(short string) string {
	for _, f := range c.Flags {
		if f.Short == short {
			return f.Name
		}
	}
	return ""
}

// showHelp displays help for the command.
func (c *Command) showHelp(ctx *Context) error {
	fmt.Fprintln(ctx.Stdout, c.helpText())
	return nil
}

// helpText generates the help text for the command.
func (c *Command) helpText() string {
	var sb strings.Builder

	// Description
	if c.Long != "" {
		sb.WriteString(c.Long)
	} else if c.Short != "" {
		sb.WriteString(c.Short)
	}
	sb.WriteString("\n\n")

	// Usage
	sb.WriteString("Usage:\n")
	sb.WriteString("  ")
	sb.WriteString(c.fullName())

	if len(c.SubCommands) > 0 {
		sb.WriteString(" [command]")
	}

	if len(c.Flags) > 0 {
		sb.WriteString(" [flags]")
	}

	if len(c.Args) > 0 {
		for _, a := range c.Args {
			if a.Required {
				sb.WriteString(" <")
			} else {
				sb.WriteString(" [")
			}
			sb.WriteString(a.Name)
			if a.Required {
				sb.WriteString(">")
			} else {
				sb.WriteString("]")
			}
		}
	}

	sb.WriteString("\n")

	// Examples
	if c.Usage != "" {
		sb.WriteString("\nExamples:\n")
		sb.WriteString(c.Usage)
		sb.WriteString("\n")
	}

	// Available commands
	if len(c.SubCommands) > 0 {
		sb.WriteString("\nAvailable Commands:\n")

		// Calculate max name length for alignment
		maxLen := 0
		for _, sub := range c.SubCommands {
			if !sub.Hidden && len(sub.Name) > maxLen {
				maxLen = len(sub.Name)
			}
		}

		for _, sub := range c.SubCommands {
			if !sub.Hidden {
				sb.WriteString(fmt.Sprintf("  %-*s  %s\n", maxLen, sub.Name, sub.Short))
			}
		}
	}

	// Arguments
	if len(c.Args) > 0 {
		sb.WriteString("\nArguments:\n")
		maxLen := 0
		for _, a := range c.Args {
			if len(a.Name) > maxLen {
				maxLen = len(a.Name)
			}
		}

		for _, a := range c.Args {
			required := ""
			if a.Required {
				required = " (required)"
			}
			sb.WriteString(fmt.Sprintf("  %-*s  %s%s\n", maxLen, a.Name, a.Description, required))
		}
	}

	// Flags
	if len(c.Flags) > 0 {
		sb.WriteString("\nFlags:\n")

		for _, f := range c.Flags {
			if f.Hidden {
				continue
			}

			flagStr := "  "
			if f.Short != "" {
				flagStr += fmt.Sprintf("-%s, ", f.Short)
			} else {
				flagStr += "    "
			}
			flagStr += fmt.Sprintf("--%s", f.Name)

			if f.Default != "" && f.Default != "true" && f.Default != "false" {
				flagStr += fmt.Sprintf(" (default: %s)", f.Default)
			}

			sb.WriteString(fmt.Sprintf("%-30s  %s\n", flagStr, f.Usage))
		}
	}

	// Help hint
	if len(c.SubCommands) > 0 {
		sb.WriteString(fmt.Sprintf("\nUse \"%s [command] --help\" for more information about a command.\n", c.fullName()))
	}

	return sb.String()
}

// fullName returns the full command path (e.g., "sekai-cli bank send").
func (c *Command) fullName() string {
	if c.parent != nil {
		return c.parent.fullName() + " " + c.Name
	}
	return c.Name
}

// Root returns the root command.
func (c *Command) Root() *Command {
	if c.parent != nil {
		return c.parent.Root()
	}
	return c
}

// GetFlag gets a flag value from context, checking parent contexts.
func (ctx *Context) GetFlag(name string) string {
	if v, ok := ctx.Flags[name]; ok {
		return v
	}
	if ctx.parent != nil {
		return ctx.parent.GetFlag(name)
	}
	return ""
}

// GetArg gets a positional argument by index.
func (ctx *Context) GetArg(index int) string {
	if index < 0 || index >= len(ctx.Args) {
		return ""
	}
	return ctx.Args[index]
}

// Printf prints formatted output to stdout.
func (ctx *Context) Printf(format string, args ...interface{}) {
	fmt.Fprintf(ctx.Stdout, format, args...)
}

// Println prints a line to stdout.
func (ctx *Context) Println(args ...interface{}) {
	fmt.Fprintln(ctx.Stdout, args...)
}

// Errorf prints formatted error to stderr.
func (ctx *Context) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(ctx.Stderr, format, args...)
}

// contains checks if a slice contains a string.
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
