package vickgenda // Changed package name

import (
	"fmt"

	"github.com/spf13/cobra"
	"vickgenda-cli/cmd/cli" // Added import for cli package
)

// setupCmd representa o comando de configuração
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Executa o assistente de configuração inicial do Vickgenda",
	Long: `Inicia um assistente interativo para configurar o Vickgenda
para o primeiro uso. Isso pode incluir a configuração da autenticação do usuário,
conexões de banco de dados (se houver) e preferências padrão.`, // Traduzido
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Assistente de configuração iniciado. Os passos de configuração serão implementados aqui.")
	},
}

// A função init adiciona o setupCmd ao rootCmd
func init() {
	cli.GetRootCmd().AddCommand(setupCmd)
}
