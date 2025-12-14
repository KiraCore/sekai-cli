// Package integration provides integration tests for the layer2 module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/layer2"
)

// TestLayer2AllDapps tests querying all dapps.
func TestLayer2AllDapps(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := layer2.New(client)
	result, err := mod.AllDapps(ctx)
	requireNoError(t, err, "Failed to query all dapps")

	t.Logf("All dapps: %s", string(result))
}

// TestLayer2TransferDapps tests querying transfer dapps.
func TestLayer2TransferDapps(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := layer2.New(client)
	result, err := mod.TransferDapps(ctx)
	requireNoError(t, err, "Failed to query transfer dapps")

	t.Logf("Transfer dapps: %s", string(result))
}
