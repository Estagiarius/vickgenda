package squad4

import (
	"fmt"
	"github.com/spf13/cobra"
)

// DashboardCmd represents the dashboard command
var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Exibe o painel principal com informações resumidas.",
	Long:  `O comando 'dashboard' é o ponto de entrada visual principal, mostrando um resumo de tarefas, eventos e outras informações relevantes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("dashboard called, logic to be implemented.")
		return nil
	},
}

// RelembrarCmd represents the relembrar command
var RelembrarCmd = &cobra.Command{
	Use:   "relembrar",
	Short: "Gerencia lembretes e recordações.",
	Long:  `O comando 'relembrar' ajuda você a não esquecer de informações importantes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("relembrar called, logic to be implemented.")
		return nil
	},
}

// FocoCmd represents the foco command
var FocoCmd = &cobra.Command{
	Use:   "foco",
	Short: "Ativa o modo de foco para trabalho ininterrupto.",
	Long:  `O comando 'foco' inicia uma sessão de trabalho focado, minimizando distrações.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("foco called, logic to be implemented.")
		return nil
	},
}

// RelatorioCmd represents the relatorio command
var RelatorioCmd = &cobra.Command{
	Use:   "relatorio",
	Short: "Gera relatórios sobre produtividade e atividades.",
	Long:  `O comando 'relatorio' fornece insights sobre seu uso do Vickgenda e progresso.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("relatorio called, logic to be implemented.")
		return nil
	},
}

// InitSquad4Commands initializes and adds all Squad 4 commands to the provided root command.
func InitSquad4Commands(rootCmd *cobra.Command) {
	// Add top-level commands from Squad 4
	rootCmd.AddCommand(DashboardCmd)
	rootCmd.AddCommand(RelembrarCmd) // RelembrarCmd itself has subcommands (adicionar, listar)
	rootCmd.AddCommand(FocoCmd)       // FocoCmd itself has subcommands (iniciar)
	rootCmd.AddCommand(RelatorioCmd)  // RelatorioCmd itself has subcommands (produtividade, etc.)
}
