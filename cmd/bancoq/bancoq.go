package bancoq

import (
	// "fmt" // DEBUG - No longer needed here if init is removed
	// "vickgenda-cli/cmd" // Removed to break import cycle
	"github.com/spf13/cobra"
)

// BancoqCmd represents the bancoq command
// This will be added to rootCmd in cmd/root.go's init()
var BancoqCmd = &cobra.Command{
	Use:   "bancoq",
	Short: "Gerencia o banco de questões.",
	Long:  `O comando 'bancoq' permite adicionar, listar, visualizar, editar e remover questões do banco de dados.`,
	// Run: func(cmd *cobra.Command, args []string) { fmt.Println("bancoq called") },
}

// init() is removed from here.
// The BancoqCmd is added to rootCmd in cmd/root.go's init function.
