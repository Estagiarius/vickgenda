package bancoq

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var bancoqViewCmd = &cobra.Command{
	Use:   "view <ID_DA_QUESTAO>",
	Short: "Visualiza todos os detalhes de uma questão específica",
	Long:  `Visualiza todos os detalhes de uma questão específica, dado o seu ID.`,
	Args:  cobra.ExactArgs(1), // Ensures exactly one argument - the ID - is provided
	Run:   runViewQuestion,
}

func init() {
	BancoqCmd.AddCommand(bancoqViewCmd)
	// No flags for this command yet, but could add --show-answers or format options later.
}

func runViewQuestion(cmd *cobra.Command, args []string) {
	if err := db.InitDB(); err != nil { // Ensure DB is initialized
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	questionID := args[0]

	question, err := db.GetQuestion(questionID)
	if err != nil {
		// Check if the error is because the question was not found
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") { // db.GetQuestion returns a wrapped error
			fmt.Fprintf(os.Stderr, "Erro: Questão com ID '%s' não encontrada.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao buscar questão ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return
	}

	fmt.Printf("Detalhes da Questão ID: %s\n\n", question.ID)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(true) // Allow text wrapping
	// table.SetHeaderAlignment(tablewriter.ALIGN_LEFT) // Not strictly needed for key-value
	// table.SetAlignment(tablewriter.ALIGN_LEFT) // Default alignment is left

	// For a key-value display, we don't need headers, but we want two columns.
	// We can achieve this by just appending pairs.
	// To make it look more like a definition list, we can manually format or use specific table settings.
	// table.SetBorder(false) // No outer border
	// table.SetColumnSeparator(":")
	// table.SetHeaderLine(false) // No line after headers (if headers were used)
	// table.SetCenterSeparator("") // No line between rows if we want compact list

	data := [][]string{
		{"ID:", question.ID},
		{"Matéria:", question.Subject},
		{"Tópico:", question.Topic},
		{"Dificuldade:", models.FormatDifficultyToPtBR(question.Difficulty)},
		{"Tipo:", models.FormatQuestionTypeToPtBR(question.QuestionType)},
		{"Texto da Questão:", question.QuestionText},
	}

	if len(question.AnswerOptions) > 0 {
		// Display each option on a new line within the cell
		optionsFormatted := ""
		for i, opt := range question.AnswerOptions {
			optionsFormatted += fmt.Sprintf("%d. %s", i+1, opt)
			if i < len(question.AnswerOptions)-1 {
				optionsFormatted += "\n"
			}
		}
		data = append(data, []string{"Opções de Resposta:", optionsFormatted})
	}

	// Display each correct answer on a new line
	correctAnswersFormatted := ""
	for i, ans := range question.CorrectAnswers {
		correctAnswersFormatted += fmt.Sprintf("- %s", ans)
		if i < len(question.CorrectAnswers)-1 {
			correctAnswersFormatted += "\n"
		}
	}
	data = append(data, []string{"Respostas Corretas:", correctAnswersFormatted})

	if question.Source != "" {
		data = append(data, []string{"Fonte:", question.Source})
	}
	if len(question.Tags) > 0 {
		data = append(data, []string{"Tags:", strings.Join(question.Tags, ", ")})
	}
	if question.Author != "" {
		data = append(data, []string{"Autor:", question.Author})
	}

	data = append(data, []string{"Criada em:", question.CreatedAt.Format(time.RFC1123Z)})
	data = append(data, []string{"Usada pela Última Vez:", models.FormatLastUsedAt(question.LastUsedAt)})

	for _, v := range data {
        table.Append(v)
    }
	table.Render()
}
