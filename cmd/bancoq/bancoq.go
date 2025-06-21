package bancoq

import (
	"fmt"
	"os"

	"vickgenda-cli/internal/db" // Importar o pacote db

	"github.com/spf13/cobra"
)

// BancoqCmd representa o comando bancoq
var BancoqCmd = &cobra.Command{
	Use:   "bancoq",
	Short: "Gerencia o banco de questões",
	Long: `O comando 'bancoq' é o ponto de entrada para todas as operações relacionadas ao banco de questões.
Ele permite adicionar, editar, excluir, listar, visualizar, buscar e importar questões.
Utilize os subcomandos para realizar as ações específicas. Por exemplo, 'vickgenda bancoq add' para adicionar uma nova questão.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Inicializa o banco de dados antes de qualquer subcomando de bancoq ser executado.
		// Usar "" para o caminho fará com que InitDB use o caminho padrão.
		if err := db.InitDB(""); err != nil {
			// Imprimir o erro para stderr para que seja visível mesmo se o log não estiver configurado.
			fmt.Fprintf(os.Stderr, "Erro crítico ao inicializar o banco de dados: %v\n", err)
			// Retornar o erro fará com que o Cobra pare a execução do comando.
			return fmt.Errorf("falha ao inicializar o banco de dados: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Se 'bancoq' for chamado sem subcomandos, mostrar ajuda.
		return cmd.Help()
	},
}

// A adição de BancoqCmd ao rootCmd é feita em cmd/root.go ou cmd/cli/cli.go
