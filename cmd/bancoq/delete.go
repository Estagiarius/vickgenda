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
	Long:  `Remove uma questão do banco de dados, dado o seu ID. Solicita confirmação a menos que a flag --force seja usada.`,
	Args:  cobra.ExactArgs(1),
	Run:   runDeleteQuestion,
}

var forceDelete bool

func init() {
	BancoqCmd.AddCommand(bancoqDeleteCmd)
	bancoqDeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Força a remoção sem confirmação")
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
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("Tem certeza que deseja remover a questão com ID '%s'?", questionID),
			Default: false,
		}
		err := survey.AskOne(confirmPrompt, &confirmed)
		if err != nil { // Handle potential error from survey itself (e.g., non-interactive environment)
			// If survey fails in a non-interactive environment, and force is not set,
			// it's safer to assume no confirmation.
			// However, for this tool, EOF often means "no" or abort.
			fmt.Fprintf(os.Stderr, "Falha ao obter confirmação: %v. Remoção cancelada.\n", err)
			// os.Exit(1); // Or simply return
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
			fmt.Fprintf(os.Stderr, "Erro: Questão com ID '%s' não encontrada para remoção.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao remover questão: %v\n", err)
		}
		os.Exit(1) // Exit with error status if deletion failed
		return
	}

	fmt.Printf("Questão '%s' removida com sucesso.\n", questionID)
}
