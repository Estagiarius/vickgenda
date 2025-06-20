package prova

import "github.com/spf13/cobra"

// ProvaCmd representa o comando raiz para gerenciar provas.
var ProvaCmd = &cobra.Command{
	Use:   "prova",
	Short: "Gerencia provas e avaliações",
	Long:  `Permite gerar, listar, visualizar, deletar e exportar provas criadas a partir do banco de questões.`,
}

// Execute adiciona todos os comandos filhos ao comando raiz e define flags apropriadamente.
// Esta é chamada por main.main(). Ela só precisa acontecer uma vez para o rootCmd.
func Execute() error {
	return ProvaCmd.Execute()
}

func init() {
	// Aqui você pode adicionar inicializações, como flags globais para o comando prova.
}
