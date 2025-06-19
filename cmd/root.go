package cmd

import (
	"fmt"
	"os" // Required for os.Exit

	"vickgenda-cli/cmd/bancoq"       // Direct import for adding command
	vickgendamain "vickgenda-cli/cmd/vickgenda" // Import for GetMainRootCmd
	"vickgenda-cli/internal/db"      // Import for db.InitDB
)

func init() {
	// Initialize the database
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Get the main root command from cmd/vickgenda/main.go
	mainRootCmd := vickgendamain.GetMainRootCmd()

	// Add commands previously managed by this package's rootCmd
	mainRootCmd.AddCommand(bancoq.BancoqCmd)
	// If there were other commands added to the old rootCmd in this package, add them here.
	// For example:
	// mainRootCmd.AddCommand(anotherCmdFromThisPackage)
}

// Note: The local rootCmd, Execute(), and GetRootCmd() have been removed as per instructions.
// The application's entry point and root command management are now centralized
// in cmd/vickgenda/main.go.
