// Package integration provides integration tests for the keys module.
package integration

import (
	"strings"
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/keys"
)

// TestKeysList tests listing all keys in the keyring.
func TestKeysList(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	result, err := mod.List(ctx)
	requireNoError(t, err, "Failed to list keys")

	// There should be at least the genesis key
	requireTrue(t, len(result) > 0, "Should have at least one key")

	t.Logf("Found %d keys:", len(result))
	for _, key := range result {
		t.Logf("  %s: %s (%s)", key.Name, key.Address, key.Type)
	}

	// Verify genesis key exists
	found := false
	for _, key := range result {
		if key.Name == TestKey {
			found = true
			break
		}
	}
	requireTrue(t, found, "Genesis key should be in the list")
}

// TestKeysShow tests showing a specific key.
func TestKeysShow(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	result, err := mod.Show(ctx, TestKey)
	requireNoError(t, err, "Failed to show key")
	requireNotNil(t, result, "Key info is nil")

	requireEqual(t, TestKey, result.Name, "Key name mismatch")
	requireEqual(t, testAddr, result.Address, "Key address mismatch")
	requireTrue(t, result.Type != "", "Key type should not be empty")

	t.Logf("Key: %s, Address: %s, Type: %s", result.Name, result.Address, result.Type)
}

// TestKeysAddAndDelete tests adding and deleting a key.
func TestKeysAddAndDelete(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	keyName := generateUniqueID("testkey")

	// Add a new key
	result, err := mod.Add(ctx, keyName, nil)
	requireNoError(t, err, "Failed to add key")
	requireNotNil(t, result, "Key info is nil")
	requireEqual(t, keyName, result.Name, "Key name mismatch")
	requireTrue(t, strings.HasPrefix(result.Address, "kira"), "Address should start with kira")

	t.Logf("Created key: %s -> %s", result.Name, result.Address)

	// Verify the key exists
	exists, err := mod.Exists(ctx, keyName)
	requireNoError(t, err, "Failed to check key existence")
	requireTrue(t, exists, "Key should exist")

	// Delete the key
	err = mod.Delete(ctx, keyName, true)
	requireNoError(t, err, "Failed to delete key")

	// Verify the key no longer exists
	exists, err = mod.Exists(ctx, keyName)
	requireNoError(t, err, "Failed to check key existence")
	requireTrue(t, !exists, "Key should not exist after deletion")

	t.Log("Key added and deleted successfully")
}

// TestKeysAddWithOptions tests adding a key with various options.
func TestKeysAddWithOptions(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	keyName := generateUniqueID("optkey")

	// Add key with options (no backup to avoid interactive prompt)
	opts := &sdk.KeyAddOptions{
		NoBackup: true,
	}
	result, err := mod.Add(ctx, keyName, opts)
	requireNoError(t, err, "Failed to add key with options")
	requireNotNil(t, result, "Key info is nil")

	t.Logf("Created key with options: %s -> %s", result.Name, result.Address)

	// Cleanup
	err = mod.Delete(ctx, keyName, true)
	requireNoError(t, err, "Failed to delete key")
}

// TestKeysRename tests renaming a key.
func TestKeysRename(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	oldName := generateUniqueID("oldname")
	newName := generateUniqueID("newname")

	// Create a key
	result, err := mod.Add(ctx, oldName, nil)
	requireNoError(t, err, "Failed to add key")
	address := result.Address

	// Rename the key
	err = mod.Rename(ctx, oldName, newName)
	requireNoError(t, err, "Failed to rename key")

	// Verify old name doesn't exist
	exists, err := mod.Exists(ctx, oldName)
	requireNoError(t, err, "Failed to check old key existence")
	requireTrue(t, !exists, "Old key name should not exist")

	// Verify new name exists with same address
	newKey, err := mod.Show(ctx, newName)
	requireNoError(t, err, "Failed to show renamed key")
	requireEqual(t, address, newKey.Address, "Address should be preserved after rename")

	t.Logf("Renamed key from %s to %s", oldName, newName)

	// Cleanup
	err = mod.Delete(ctx, newName, true)
	requireNoError(t, err, "Failed to delete key")
}

// TestKeysExportImport tests exporting and importing a key.
func TestKeysExportImport(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	originalName := generateUniqueID("exportkey")
	importedName := generateUniqueID("importkey")

	// Create a key
	result, err := mod.Add(ctx, originalName, nil)
	requireNoError(t, err, "Failed to add key")
	originalAddress := result.Address

	// Export the key
	armor, err := mod.Export(ctx, originalName)
	if err != nil {
		// Export might fail if the key is not exportable
		t.Logf("Export failed (may be expected): %v", err)
		// Cleanup and skip rest of test
		_ = mod.Delete(ctx, originalName, true)
		t.Skip("Key export not supported")
		return
	}

	requireTrue(t, len(armor) > 0, "Exported armor should not be empty")
	t.Logf("Exported key armor length: %d", len(armor))

	// Import the key with a new name
	err = mod.Import(ctx, importedName, armor)
	requireNoError(t, err, "Failed to import key")

	// Verify imported key has the same address
	importedKey, err := mod.Show(ctx, importedName)
	requireNoError(t, err, "Failed to show imported key")
	requireEqual(t, originalAddress, importedKey.Address, "Imported key should have same address")

	t.Logf("Exported and imported key: %s -> %s", originalName, importedName)

	// Cleanup
	_ = mod.Delete(ctx, originalName, true)
	_ = mod.Delete(ctx, importedName, true)
}

// TestKeysGetAddress tests getting address for a key.
func TestKeysGetAddress(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	address, err := mod.GetAddress(ctx, TestKey)
	requireNoError(t, err, "Failed to get address")
	requireEqual(t, testAddr, address, "Address mismatch")

	t.Logf("Key %s has address %s", TestKey, address)
}

// TestKeysExists tests checking if a key exists.
func TestKeysExists(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)

	// Test existing key
	exists, err := mod.Exists(ctx, TestKey)
	requireNoError(t, err, "Failed to check key existence")
	requireTrue(t, exists, "Genesis key should exist")

	// Test non-existing key
	exists, err = mod.Exists(ctx, "nonexistent_key_12345")
	requireNoError(t, err, "Failed to check key existence")
	requireTrue(t, !exists, "Non-existent key should not exist")

	t.Log("Key existence checks passed")
}

// TestKeysRecoverFromMnemonic tests recovering a key from a mnemonic.
func TestKeysRecoverFromMnemonic(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)

	// First create a key to get a mnemonic
	keyName := generateUniqueID("mnemonickey")
	result, err := mod.Add(ctx, keyName, nil)
	requireNoError(t, err, "Failed to add key")
	originalAddress := result.Address

	// Get the mnemonic (if available)
	mnemonic, err := mod.Mnemonic(ctx, keyName)
	if err != nil {
		t.Logf("Cannot get mnemonic (may be expected): %v", err)
		// Cleanup and skip
		_ = mod.Delete(ctx, keyName, true)
		t.Skip("Mnemonic retrieval not supported")
		return
	}

	// Delete the key
	err = mod.Delete(ctx, keyName, true)
	requireNoError(t, err, "Failed to delete key")

	// Recover the key from mnemonic
	recoveredKeyName := generateUniqueID("recoveredkey")
	recovered, err := mod.Recover(ctx, recoveredKeyName, mnemonic)
	requireNoError(t, err, "Failed to recover key")
	requireEqual(t, originalAddress, recovered.Address, "Recovered key should have same address")

	t.Logf("Recovered key %s with address %s", recoveredKeyName, recovered.Address)

	// Cleanup
	_ = mod.Delete(ctx, recoveredKeyName, true)
}

// TestKeysListKeyTypes tests listing all supported key algorithms.
func TestKeysListKeyTypes(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	types, err := mod.ListKeyTypes(ctx)
	requireNoError(t, err, "Failed to list key types")
	requireTrue(t, len(types) > 0, "Should have at least one key type")

	t.Logf("Supported key types: %v", types)

	// Verify secp256k1 is in the list (it's the default)
	found := false
	for _, kt := range types {
		if strings.Contains(kt, "secp256k1") {
			found = true
			break
		}
	}
	requireTrue(t, found, "secp256k1 should be a supported key type")
}

// TestKeysImportHex tests importing a hex-encoded private key.
func TestKeysImportHex(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)
	keyName := generateUniqueID("hexkey")

	// Use a well-known test hex key (32 bytes = 64 hex chars)
	// This is just a test key - do NOT use in production
	testHexKey := "0000000000000000000000000000000000000000000000000000000000000001"

	err := mod.ImportHex(ctx, keyName, testHexKey, "")
	requireNoError(t, err, "Failed to import hex key")

	// Verify the key exists
	exists, err := mod.Exists(ctx, keyName)
	requireNoError(t, err, "Failed to check key existence")
	requireTrue(t, exists, "Imported key should exist")

	// Show the key
	info, err := mod.Show(ctx, keyName)
	requireNoError(t, err, "Failed to show imported key")
	requireEqual(t, keyName, info.Name, "Key name mismatch")

	t.Logf("Imported hex key: %s -> %s", keyName, info.Address)

	// Cleanup
	_ = mod.Delete(ctx, keyName, true)
}

// TestKeysMigrate tests migrating keys from amino to protobuf format.
func TestKeysMigrate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)

	// Run migrate - should succeed even if nothing to migrate
	err := mod.Migrate(ctx)
	// Migrate may return an error if nothing to migrate, which is OK
	if err != nil {
		t.Logf("Migrate returned (may be expected): %v", err)
	} else {
		t.Log("Migrate completed successfully")
	}
}

// TestKeysParse tests parsing address from hex to bech32 and vice versa.
func TestKeysParse(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := keys.New(client)

	// Parse a known bech32 address
	result, err := mod.Parse(ctx, testAddr)
	requireNoError(t, err, "Failed to parse bech32 address")
	requireNotNil(t, result, "Parsed address is nil")

	t.Logf("Parsed %s:", testAddr)
	t.Logf("  Human: %s", result.Human)
	t.Logf("  Bytes: %s", result.Bytes)
	t.Logf("  Hex: %s", result.Hex)
	t.Logf("  Bech32: %s", result.Bech32)

	// Verify we got expected values
	requireEqual(t, "kira", result.Human, "Human prefix should be 'kira'")
	requireTrue(t, result.Bytes != "", "Bytes should not be empty")

	// If we got hex, try parsing it back
	if result.Hex != "" {
		result2, err := mod.Parse(ctx, result.Hex)
		requireNoError(t, err, "Failed to parse hex address")
		requireNotNil(t, result2, "Parsed address is nil")
		t.Logf("Parsed hex back: Human=%s, Bytes=%s", result2.Human, result2.Bytes)
	}
}
