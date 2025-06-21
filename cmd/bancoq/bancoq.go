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
	Short: "Gerencia o banco de questões",
	Long:  `O comando 'bancoq' é o ponto de entrada para todas as operações relacionadas ao banco de questões.
Ele permite adicionar, editar, excluir, listar, visualizar, buscar e importar questões.
Utilize os subcomandos para realizar as ações específicas. Por exemplo, 'bancoq add' para adicionar uma nova questão.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("bancoq called, logic to be implemented.")
		return nil
	},
}

// init() is removed from here.
// The BancoqCmd is added to rootCmd in cmd/root.go's init function.
