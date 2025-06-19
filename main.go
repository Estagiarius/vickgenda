package main

import (

	"vickgenda-cli/cmd" // Updated to use the new cmd package
)

func main() {
	// db.InitDB() // Initialization will be handled by individual commands or a PersistentPreRun later
	cmd.Execute()
=======
	"fmt"
	"os"

	// Ensure this import path matches your module name in go.mod + path to package
	"professor-cli/internal/squad4"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "professor-cli",
	Short: "Professor CLI é uma ferramenta para auxiliar professores em suas tarefas diárias.",
	Long: `Uma interface de linha de comando para ajudar professores com gerenciamento de tarefas,
agenda, notas, foco, e mais.`,
	// Uncomment the following line if you want to print help when no subcommand is provided
	// Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func main() {
	// Initialize Squad 4 commands and add them to the root command
	squad4.InitSquad4Commands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar o comando: '%s'\n", err)
		os.Exit(1)
	}

}
