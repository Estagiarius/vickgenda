package squad4

import (
	"github.com/spf13/cobra"
)

// InitSquad4Commands initializes and adds all Squad 4 commands to the provided root command.
func InitSquad4Commands(rootCmd *cobra.Command) {
	// Add top-level commands from Squad 4
	rootCmd.AddCommand(DashboardCmd)
	rootCmd.AddCommand(RelembrarCmd) // RelembrarCmd itself has subcommands (adicionar, listar)
	rootCmd.AddCommand(FocoCmd)       // FocoCmd itself has subcommands (iniciar)
	rootCmd.AddCommand(RelatorioCmd)  // RelatorioCmd itself has subcommands (produtividade, etc.)
}
