package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"vickgenda-cli/cmd/bancoq" // Direct import for adding command
)

func init() {
	// fmt.Println("DEBUG: init() in cmd/root.go called") // DEBUG line removed
	// Add subcommands here
	rootCmd.AddCommand(bancoq.BancoqCmd)
}

var rootCmd = &cobra.Command{
	Use:   "vickgenda",
	Short: "Vickgenda CLI é uma ferramenta para ajudar professores a gerenciar questões e provas.",
	Long: `Vickgenda CLI é uma aplicação de linha de comando para gerenciar
bancos de questões pedagógicas, gerar provas e outras tarefas relacionadas à docência.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// fmt.Println("DEBUG: cmd.Execute() in root.go called") // DEBUG line removed
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// GetRootCmd returns the root command. Used to add subcommands from other packages.
func GetRootCmd() *cobra.Command {
	// fmt.Println("DEBUG: cmd.GetRootCmd() in root.go called") // DEBUG line removed
	return rootCmd
}
