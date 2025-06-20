package prova

import "github.com/spf13/cobra"

// ProvaCmd representa o comando raiz para gerenciar provas.
var ProvaCmd = &cobra.Command{
	Use:   "prova",
	Short: "Gerencia provas e avaliações",
	Long:  `Permite gerar, listar, visualizar, deletar e exportar provas e avaliações educacionais. Utilize os subcomandos para realizar as ações desejadas.`,
}

// Execute adiciona todos os comandos filhos ao comando raiz e define flags apropriadamente.
// Esta é chamada por main.main(). Ela só precisa acontecer uma vez para o rootCmd.
// Execute executa o comando raiz.
func Execute() error {
	return ProvaCmd.Execute()
}

func init() {
	// Flags globais para o comando 'prova' podem ser adicionadas aqui.
	// Exemplo: ProvaCmd.PersistentFlags().StringP("config", "c", "", "Arquivo de configuração para o comando prova")
}
