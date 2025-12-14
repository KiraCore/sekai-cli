// Package keys provides key management functionality.
package keys

import (
	"context"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides key management functionality.
// It wraps the KeysClient from the SDK client.
type Module struct {
	client sdk.Client
}

// New creates a new keys module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Add creates a new key with the given name.
func (m *Module) Add(ctx context.Context, name string, opts *sdk.KeyAddOptions) (*sdk.KeyInfo, error) {
	return m.client.Keys().Add(ctx, name, opts)
}

// Delete removes a key by name.
func (m *Module) Delete(ctx context.Context, name string, force bool) error {
	return m.client.Keys().Delete(ctx, name, force)
}

// List returns all keys in the keyring.
func (m *Module) List(ctx context.Context) ([]sdk.KeyInfo, error) {
	return m.client.Keys().List(ctx)
}

// Show returns information about a specific key.
func (m *Module) Show(ctx context.Context, name string) (*sdk.KeyInfo, error) {
	return m.client.Keys().Show(ctx, name)
}

// Export exports a key as ASCII-armored string.
func (m *Module) Export(ctx context.Context, name string) (string, error) {
	return m.client.Keys().Export(ctx, name)
}

// Import imports a key from ASCII-armored string.
func (m *Module) Import(ctx context.Context, name, armor string) error {
	return m.client.Keys().Import(ctx, name, armor)
}

// Rename renames a key.
func (m *Module) Rename(ctx context.Context, oldName, newName string) error {
	return m.client.Keys().Rename(ctx, oldName, newName)
}

// Mnemonic returns the mnemonic for a key (if available).
func (m *Module) Mnemonic(ctx context.Context, name string) (string, error) {
	return m.client.Keys().Mnemonic(ctx, name)
}

// GetAddress returns the address for a key.
func (m *Module) GetAddress(ctx context.Context, name string) (string, error) {
	info, err := m.client.Keys().Show(ctx, name)
	if err != nil {
		return "", err
	}
	return info.Address, nil
}

// Exists checks if a key exists.
func (m *Module) Exists(ctx context.Context, name string) (bool, error) {
	_, err := m.client.Keys().Show(ctx, name)
	if err != nil {
		if err == sdk.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateOptions configures key creation.
type CreateOptions struct {
	// Recover indicates whether to recover from mnemonic
	Recover bool

	// Mnemonic is the BIP39 mnemonic (required if Recover is true)
	Mnemonic string

	// HDPath is the HD derivation path (default: m/44'/118'/0'/0/0)
	HDPath string

	// Algorithm is the key algorithm (default: secp256k1)
	Algorithm string

	// NoBackup skips mnemonic display
	NoBackup bool

	// Index is the account index for HD derivation
	Index uint32
}

// ToSDKOptions converts CreateOptions to sdk.KeyAddOptions.
func (o *CreateOptions) ToSDKOptions() *sdk.KeyAddOptions {
	if o == nil {
		return nil
	}
	return &sdk.KeyAddOptions{
		Recover:   o.Recover,
		Mnemonic:  o.Mnemonic,
		HDPath:    o.HDPath,
		Algorithm: o.Algorithm,
		NoBackup:  o.NoBackup,
		Index:     o.Index,
	}
}

// Create creates a new key with the given options.
func (m *Module) Create(ctx context.Context, name string, opts *CreateOptions) (*sdk.KeyInfo, error) {
	var sdkOpts *sdk.KeyAddOptions
	if opts != nil {
		sdkOpts = opts.ToSDKOptions()
	}
	return m.Add(ctx, name, sdkOpts)
}

// Recover recovers a key from a mnemonic.
func (m *Module) Recover(ctx context.Context, name, mnemonic string) (*sdk.KeyInfo, error) {
	return m.Add(ctx, name, &sdk.KeyAddOptions{
		Recover:  true,
		Mnemonic: mnemonic,
	})
}

// ImportHex imports a hex-encoded private key.
func (m *Module) ImportHex(ctx context.Context, name, hexKey, keyType string) error {
	return m.client.Keys().ImportHex(ctx, name, hexKey, keyType)
}

// ListKeyTypes returns all supported key algorithms.
func (m *Module) ListKeyTypes(ctx context.Context) ([]string, error) {
	return m.client.Keys().ListKeyTypes(ctx)
}

// Migrate migrates keys from amino to protobuf format.
func (m *Module) Migrate(ctx context.Context) error {
	return m.client.Keys().Migrate(ctx)
}

// Parse converts address from hex to bech32 or vice versa.
func (m *Module) Parse(ctx context.Context, address string) (*sdk.ParsedAddress, error) {
	return m.client.Keys().Parse(ctx, address)
}
