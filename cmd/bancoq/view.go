package bancoq

import (
	"database/sql" // Keep one
	"errors"       // Keep one
	"fmt"
	"os"
	"strings"
	// "time" // Stays commented out

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var bancoqViewCmd = &cobra.Command{
	Use:   "view <ID_DA_QUESTAO>",
	Short: "Visualiza todos os detalhes de uma questão específica",
	Long:  `Exibe todos os detalhes de uma questão específica do banco de dados, identificada pelo seu ID. As informações são apresentadas em formato de lista de definições.`,
	Args:  cobra.ExactArgs(1), // Ensures exactly one argument - the ID - is provided
	Run:   runViewQuestion,
}

func init() {
	BancoqCmd.AddCommand(bancoqViewCmd)
	// No flags for this command yet, but could add --show-answers or format options later.
	// Exemplo: bancoqViewCmd.Flags().Bool("gabarito", false, "Exibir também o gabarito/respostas corretas")
}

func runViewQuestion(cmd *cobra.Command, args []string) {
	if err := db.InitDB(""); err != nil { // Ensure DB is initialized
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	questionID := args[0]

	question, err := db.GetQuestion(questionID)
	if err != nil {
		// Check if the error is because the question was not found
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") { // db.GetQuestion returns a wrapped error
			fmt.Fprintf(os.Stderr, "Erro: A questão com ID '%s' não foi encontrada.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao buscar a questão com ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return
	}

	fmt.Printf("Detalhes da Questão ID: %s\n\n", question.ID)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(true)
	// Configurações para parecer uma lista de definições
	table.SetBorder(false)
	table.SetColumnSeparator(":")
	table.SetHeaderLine(false)
	table.SetCenterSeparator("")
	table.SetTablePadding("  ") // Adiciona um pouco de padding
    table.SetAlignment(tablewriter.ALIGN_LEFT)


	data := [][]string{
		{"ID", question.ID}, // Removido ":" para consistência com tablewriter
		{"Disciplina", question.Subject},
		{"Tópico", question.Topic},
		{"Dificuldade", models.FormatDifficultyToPtBR(question.Difficulty)},
		{"Tipo", models.FormatQuestionTypeToPtBR(question.QuestionType)},
		{"Texto da Questão", question.QuestionText},
	}

	if len(question.AnswerOptions) > 0 {
		optionsFormatted := ""
		for i, opt := range question.AnswerOptions {
			optionsFormatted += fmt.Sprintf("%c) %s", 'A'+i, opt) // Usando letras para opções
			if i < len(question.AnswerOptions)-1 {
				optionsFormatted += "\n"
			}
		}
		data = append(data, []string{"Opções de Resposta", optionsFormatted})
	}

	correctAnswersFormatted := ""
	if len(question.CorrectAnswers) > 0 {
		for i, ans := range question.CorrectAnswers {
			correctAnswersFormatted += fmt.Sprintf("- %s", ans)
			if i < len(question.CorrectAnswers)-1 {
				correctAnswersFormatted += "\n"
			}
		}
	} else {
		correctAnswersFormatted = "(Não especificado)"
	}
	data = append(data, []string{"Respostas Corretas", correctAnswersFormatted})

	if question.Source != "" {
		data = append(data, []string{"Fonte", question.Source})
	}
	if len(question.Tags) > 0 {
		data = append(data, []string{"Tags", strings.Join(question.Tags, ", ")})
	}
	if question.Author != "" {
		data = append(data, []string{"Autor", question.Author})
	}
	// Year is not a field in models.Question, removing this line.
	// if question.Year > 0 {
	// 	data = append(data, []string{"Ano", fmt.Sprintf("%d",question.Year)})
	// }

	data = append(data, []string{"Criada em", question.CreatedAt.Format("02/01/2006 15:04:05 MST")})
	// UpdatedAt is not a field in models.Question
	// data = append(data, []string{"Atualizada em", question.UpdatedAt.Format("02/01/2006 15:04:05 MST")})
	data = append(data, []string{"Usada pela Última Vez", models.FormatLastUsedAt(question.LastUsedAt)})
	// UsageCount is not a field in models.Question
	// data = append(data, []string{"Contagem de Uso", fmt.Sprintf("%d", question.UsageCount)})
	// IsPublic is not a field in models.Question
	// data = append(data, []string{"Pública", fmt.Sprintf("%t", question.IsPublic)})

	for _, v := range data {
        table.Append(v)
    }
	table.Render()
}
