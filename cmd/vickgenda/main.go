package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd representa o comando base quando chamado sem nenhum subcomando
var rootCmd = &cobra.Command{
	Use:   "vickgenda",
	Short: "Vickgenda é uma aplicação CLI de agenda.",
	Long: `Uma ferramenta de linha de comando para ajudar no gerenciamento de tarefas, aulas e produtividade.
Vickgenda é uma aplicação CLI compreensiva construída com Go,
Cobra, e Bubble Tea, desenhada para ajudar você a gerenciar suas tarefas e eventos
eficientemente diretamente do seu terminal.`, // Descrição Long combinada: original com novo pt-BR
	// Descomente a linha a seguir se sua aplicação simples
	// tiver uma ação associada a ela:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// versionCmd representa o comando de versão
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostra a versão do Vickgenda",
	Long:  `Todo software tem versões. Esta é a do Vickgenda.`, // Traduzido
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Vickgenda version 0.1.0")
	},
}

// Execute adiciona todos os comandos filhos ao comando raiz e define as flags apropriadamente.
// Isto é chamado por main.main(). Só precisa acontecer uma vez para o rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	// Adiciona versionCmd ao rootCmd
	rootCmd.AddCommand(versionCmd)
	Execute()
}
