package bancoq

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"vickgenda-cli/internal/db"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var bancoqDeleteCmd = &cobra.Command{
	Use:   "delete <ID_DA_QUESTAO>",
	Short: "Remove uma questão do banco de dados",
	Long:  `Remove permanentemente uma questão específica do banco de dados, utilizando o seu ID. Por padrão, solicita confirmação antes de excluir, a menos que a flag --force seja utilizada.`,
	Args:  cobra.ExactArgs(1),
	Run:   runDeleteQuestion,
}

var forceDelete bool

func init() {
	BancoqCmd.AddCommand(bancoqDeleteCmd)
	bancoqDeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Força a remoção sem pedir confirmação")
}

func runDeleteQuestion(cmd *cobra.Command, args []string) {
	if err := db.InitDB(); err != nil { // Ensure DB is initialized
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	questionID := args[0]

	// Confirmation step (if force is false)
	confirmed := forceDelete // If forceDelete is true, confirmed is true
	if !forceDelete {
		// Primeiro, tentar buscar a questão para exibir seu texto (ou parte dele) na confirmação.
		// Isso torna a confirmação mais segura para o usuário.
		var questionTextPreview string
		q, err := db.GetQuestionByID(questionID) // Supondo que esta função exista e retorne a questão ou um erro.
		if err != nil || q == nil {
			// Se não encontrar ou der erro, ainda perguntar, mas sem o texto.
			questionTextPreview = fmt.Sprintf("com ID '%s' (detalhes não puderam ser carregados)", questionID)
		} else {
			if len(q.Text) > 50 { // Limitar o preview do texto
				questionTextPreview = fmt.Sprintf("'%s...' (ID: %s)", q.Text[:50], questionID)
			} else {
				questionTextPreview = fmt.Sprintf("'%s' (ID: %s)", q.Text, questionID)
			}
		}

		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("Tem certeza que deseja remover a questão %s?", questionTextPreview),
			Default: false,
			Help:    "Esta ação é irreversível e removerá permanentemente a questão do banco de dados.",
		}
		err = survey.AskOne(confirmPrompt, &confirmed) // Reatribuir err
		if err != nil {
			fmt.Fprintf(os.Stderr, "Falha ao obter confirmação: %v. Remoção cancelada.\n", err)
			return
		}
	}

	if !confirmed {
		fmt.Println("Remoção cancelada pelo usuário.")
		return
	}

	// Attempt to delete the question
	err := db.DeleteQuestion(questionID) // Reatribuir err
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Idealmente, db.DeleteQuestion ou db.GetQuestionByID retornaria um erro específico
			fmt.Fprintf(os.Stderr, "Erro: A questão com ID '%s' não foi encontrada para remoção.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao remover a questão: %v\n", err)
		}
		os.Exit(1) // Exit with error status if deletion failed
		return
	}

	fmt.Printf("Questão com ID '%s' removida com sucesso.\n", questionID)
}
