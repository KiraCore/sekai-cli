// Package main is the entry point for sekai-cli.
package main

import (
	"github.com/kiracore/sekai-cli/internal/app"
)

// Version information (set by build flags).
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	app.RunCLI()
}
