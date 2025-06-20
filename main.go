package main

import (
	"fmt"
	"os"

	"vickgenda-cli/cmd/cli"    // For cli.SetupRootCmd, cli.Execute
	"vickgenda-cli/internal/db" // For db.InitDB

	// Import packages for side effects (to run their init() functions)
	_ "vickgenda-cli/cmd"               // For cmd/root.go init()
	_ "vickgenda-cli/cmd/bancoq"        // For cmd/bancoq init() which adds to cmd.BancoqCmd
	_ "vickgenda-cli/cmd/prova"         // For cmd/prova init()
	_ "vickgenda-cli/cmd/vickgenda"     // For cmd/vickgenda/auth.go and setup.go init()
	_ "vickgenda-cli/internal/squad4"   // For squad4 commands init (if any)
	_ "vickgenda-cli/internal/commands/agenda" // If agenda commands need to register themselves
	_ "vickgenda-cli/internal/commands/aula"
	_ "vickgenda-cli/internal/commands/notas"
	_ "vickgenda-cli/internal/commands/rotina"
	_ "vickgenda-cli/internal/commands/student"
	_ "vickgenda-cli/internal/commands/tarefa"
	// Add other command packages here if they have init() functions that register commands
)

func main() {
	// Initialize the database early.
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	// Setup the root command from the cli package
	// (this adds commands defined in cli, like resolveCmd and dashboardCmd)
	cli.SetupRootCmd()

	// Note: Commands from other packages (cmd/root, cmd/vickgenda/auth, cmd/vickgenda/setup)
	// are added to cli.GetRootCmd() within their own init() functions,
	// which are run due to the blank imports above.

	// Execute the root command from the cli package
	if err := cli.Execute(); err != nil {
		// Cobra's Execute usually prints errors itself.
		os.Exit(1)
	}
}
