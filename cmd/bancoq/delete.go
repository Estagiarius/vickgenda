package bancoq

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"vickgenda-cli/internal/db"

	"strings"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models" // Para models.Question no preview

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var bancoqDeleteCmd = &cobra.Command{
	Use:   "delete <ID_DA_QUESTAO>",
	Short: "Remove uma questão do banco de dados",
	Long: `Remove permanentemente uma questão específica do banco de dados, utilizando o seu ID.
Por padrão, solicita confirmação antes de excluir. Use a flag --force para pular a confirmação.
Exemplo:
  vickgenda bancoq delete 123e4567-e89b-12d3-a456-426614174000
  vickgenda bancoq delete 123e4567-e89b-12d3-a456-426614174000 --force`,
	Args: cobra.ExactArgs(1), // Garante que exatamente um argumento (o ID) seja fornecido
	Run:  runDeleteQuestion,
}

var forceDelete bool

func init() {
	BancoqCmd.AddCommand(bancoqDeleteCmd)
	bancoqDeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Força a remoção sem pedir confirmação")
}

func runDeleteQuestion(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	questionID := args[0]
	if strings.TrimSpace(questionID) == "" {
		fmt.Fprintln(os.Stderr, "Erro: O ID da questão não pode ser vazio.")
		os.Exit(1)
	}

	// Confirmation step (unless --force is used)
	confirmed := forceDelete
	if !forceDelete {
		var questionPreviewMsg string
		question, err := db.GetQuestion(questionID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Se a questão não existe, não há nada para deletar.
				// Informar o usuário e sair pode ser uma boa UX.
				fmt.Fprintf(os.Stderr, "Aviso: A questão com ID '%s' não foi encontrada. Nada para remover.\n", questionID)
				return // Sair sem erro, pois o estado desejado (questão não existe) já é verdade.
			}
			// Para outros erros ao buscar, a confirmação será mais genérica.
			questionPreviewMsg = fmt.Sprintf("com ID '%s' (detalhes não puderam ser carregados devido a erro: %v)", questionID, err)
		} else {
			// Formatar preview da questão
			textPreview := question.QuestionText
			if len(textPreview) > 70 { // Limitar o preview do texto
				runes := []rune(textPreview)
				if len(runes) > 70 {
					textPreview = string(runes[:67]) + "..."
				} else {
					textPreview = string(runes)
				}
			}
			questionPreviewMsg = fmt.Sprintf("'%s' (ID: %s, Disciplina: %s, Tópico: %s)",
				textPreview, question.ID, question.Subject, question.Topic)
		}

		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("Tem certeza que deseja remover permanentemente a questão %s?", questionPreviewMsg),
			Default: false,
			Help:    "Esta ação é irreversível.",
		}
		// Reatribuir err para o erro do survey.AskOne
		err = survey.AskOne(confirmPrompt, &confirmed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Falha ao obter confirmação: %v. Remoção cancelada.\n", err)
			os.Exit(1) // Sair se o prompt de confirmação falhar
			return
		}
	}

	if !confirmed {
		fmt.Println("Remoção cancelada pelo usuário.")
		return
	}

	// Attempt to delete the question
	err := db.DeleteQuestion(questionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Este caso é um pouco redundante se a verificação acima (ao buscar para preview) já tratou.
			// Mas é uma boa salvaguarda se a questão for deletada entre o preview e a confirmação.
			fmt.Fprintf(os.Stderr, "Erro: A questão com ID '%s' não foi encontrada para remoção (pode ter sido removida por outro processo).\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao remover a questão com ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return
	}

	fmt.Printf("Questão com ID '%s' removida com sucesso.\n", questionID)
}
