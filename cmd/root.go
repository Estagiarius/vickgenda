package cmd

import (
	// "fmt" // No longer needed for db init error
	// "os"  // No longer needed for db init error

	"vickgenda-cli/cmd/bancoq" // Direct import for adding command
	"vickgenda-cli/cmd/cli"    // Import the new cli package
	"vickgenda-cli/cmd/prova"  // Import for prova.ProvaCmd
	// "vickgenda-cli/internal/db" // No longer needed for db.InitDB here
)

func init() {
	// The database initialization is now handled in cmd/vickgenda/main.go's main().

	// Get the main root command from cmd/cli/cli.go
	mainRootCmd := cli.GetRootCmd()

	// Add commands previously managed by this package's rootCmd
	mainRootCmd.AddCommand(bancoq.BancoqCmd)
	mainRootCmd.AddCommand(prova.ProvaCmd)
}

// Note: The local rootCmd, Execute(), and GetRootCmd() have been removed previously.
// The application's entry point and root command management are now centralized
// in the cli package and cmd/vickgenda/main.go.
