package bancoq

import (
	"fmt"
	"math"
	"os"
	"strings"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCommandFlags holds flag values for the list command
var listCommandFlags struct {
	Subject    string
	Topic      string
	Difficulty string
	Type       string
	Author     string
	Tags       []string
	Limit      int
	Page       int
	SortBy     string
	Order      string
}

var bancoqListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista as questões do banco de questões",
	Long: `Exibe uma lista paginada das questões armazenadas no banco de dados.
Permite aplicar diversos filtros para refinar a busca e ordenar os resultados.
Exemplo:
  vickgenda bancoq list --subject "História" --difficulty "medium" --limit 10 --page 2 --sort-by "topic" --order "asc"`,
	Run: runListQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqListCmd)

	// Filter flags
	bancoqListCmd.Flags().StringVar(&listCommandFlags.Subject, "subject", "", "Filtrar por disciplina (ex: \"Matemática\")")
	bancoqListCmd.Flags().StringVar(&listCommandFlags.Topic, "topic", "", "Filtrar por tópico (ex: \"Álgebra Linear\")")
	bancoqListCmd.Flags().StringVar(&listCommandFlags.Difficulty, "difficulty", "", fmt.Sprintf("Filtrar por dificuldade (valores: %s, %s, %s)", models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard))
	bancoqListCmd.Flags().StringVar(&listCommandFlags.Type, "type", "", fmt.Sprintf("Filtrar por tipo de questão (valores: %s, %s, %s, %s)", models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer))
	bancoqListCmd.Flags().StringVar(&listCommandFlags.Author, "author", "", "Filtrar por autor da questão")
	bancoqListCmd.Flags().StringSliceVar(&listCommandFlags.Tags, "tag", []string{}, "Filtrar por tag (atualmente considera apenas a primeira tag fornecida).")

	// Pagination flags
	bancoqListCmd.Flags().IntVar(&listCommandFlags.Limit, "limit", 20, "Número de questões a serem exibidas por página")
	bancoqListCmd.Flags().IntVar(&listCommandFlags.Page, "page", 1, "Número da página a ser visualizada")

	// Sort flags
	bancoqListCmd.Flags().StringVar(&listCommandFlags.SortBy, "sort-by", "created_at", "Coluna para ordenação (opções: id, subject, topic, difficulty, question_type, created_at, last_used_at, author)")
	bancoqListCmd.Flags().StringVar(&listCommandFlags.Order, "order", "desc", "Ordem da ordenação (valores: asc, desc)")
}

// isValidDifficulty checks if the provided difficulty is valid.
func isValidListDifficulty(difficulty string) bool {
	if difficulty == "" { // Allow empty for no filter
		return true
	}
	switch difficulty {
	case models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard:
		return true
	default:
		return false
	}
}

// isValidListQuestionType checks if the provided question type is valid.
func isValidListQuestionType(qType string) bool {
	if qType == "" { // Allow empty for no filter
		return true
	}
	switch qType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer:
		return true
	default:
		return false
	}
}


func runListQuestions(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	// Validate flag values
	if !isValidListDifficulty(listCommandFlags.Difficulty) {
		fmt.Fprintf(os.Stderr, "Valor inválido para --difficulty: '%s'. Use '%s', '%s' ou '%s'.\n", listCommandFlags.Difficulty, models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard)
		os.Exit(1)
	}
	if !isValidListQuestionType(listCommandFlags.Type) {
		fmt.Fprintf(os.Stderr, "Valor inválido para --type: '%s'. Use '%s', '%s', '%s' ou '%s'.\n", listCommandFlags.Type, models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer)
		os.Exit(1)
	}
	order := strings.ToLower(listCommandFlags.Order)
	if order != "asc" && order != "desc" {
		fmt.Fprintf(os.Stderr, "Valor inválido para --order: '%s'. Use 'asc' ou 'desc'.\n", listCommandFlags.Order)
		os.Exit(1)
	}
	// Note: SortBy validation is primarily handled by db.ListQuestions to keep it centralized with DB column names.

	filters := make(map[string]interface{})
	if listCommandFlags.Subject != "" {
		filters["subject"] = listCommandFlags.Subject
	}
	if listCommandFlags.Topic != "" {
		filters["topic"] = listCommandFlags.Topic
	}
	if listCommandFlags.Difficulty != "" {
		filters["difficulty"] = listCommandFlags.Difficulty
	}
	if listCommandFlags.Type != "" {
		filters["question_type"] = listCommandFlags.Type // DB field is question_type
	}
	if listCommandFlags.Author != "" {
		filters["author"] = listCommandFlags.Author
	}
	if len(listCommandFlags.Tags) > 0 {
		// CURRENT LIMITATION: db.ListQuestions currently supports filtering by a single tag string (LIKE %tag%).
		// For true multi-tag (OR/AND) functionality, db.ListQuestions would need modification.
		// For now, we'll use the first tag provided, or join them if the backend supported it differently.
		filters["tags"] = listCommandFlags.Tags[0]
		if len(listCommandFlags.Tags) > 1 {
			fmt.Fprintln(os.Stdout, "Aviso: Múltiplas tags foram fornecidas. Atualmente, a filtragem considera apenas a primeira tag:", listCommandFlags.Tags[0])
		}
	}


	questions, total, err := db.ListQuestions(filters, listCommandFlags.SortBy, order, listCommandFlags.Limit, listCommandFlags.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao listar as questões: %v\n", err)
		os.Exit(1)
	}

	if total == 0 { // Check total count first
		fmt.Println("Nenhuma questão foi encontrada com os filtros aplicados.")
		return
	}
	if len(questions) == 0 && listCommandFlags.Page > 1 { // If not on page 1 and no results for this page
		totalPages := int(math.Ceil(float64(total) / float64(listCommandFlags.Limit)))
		fmt.Printf("Nenhuma questão encontrada na página %d. Total de páginas: %d.\n", listCommandFlags.Page, totalPages)
		fmt.Printf("Total de questões no banco que correspondem aos filtros: %d.\n", total)
		return
	}


	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID Curto", "Disciplina", "Tópico", "Tipo", "Dificuldade", "Início da Questão"})
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetColWidth(60) // Set a reasonable overall column width for QuestionText preview
    table.SetAutoWrapText(false) // Prevent auto-wrapping that might break table structure with long text

	for _, q := range questions {
		idShort := q.ID
		if len(q.ID) > 12 {
			idShort = q.ID[:12] + "..."
		}

		questionTextPreview := strings.ReplaceAll(q.QuestionText, "\n", " ") // Replace newlines for single line preview
		maxPreviewLength := 47 // Max length for the preview text itself
		if len(questionTextPreview) > maxPreviewLength {
			runes := []rune(questionTextPreview)
			if len(runes) > maxPreviewLength {
				questionTextPreview = string(runes[:maxPreviewLength-3]) + "..."
			} else {
				// This case should not be hit if len(questionTextPreview) > maxPreviewLength
				// but as a fallback, use the full string if rune conversion results in shorter.
				questionTextPreview = string(runes)
			}
		}


		row := []string{
			idShort,
			q.Subject,
			q.Topic,
			models.FormatQuestionTypeToPtBR(q.QuestionType),
			models.FormatDifficultyToPtBR(q.Difficulty),
			questionTextPreview,
		}
		table.Append(row)
	}
	table.Render()

	totalPages := 1 // Default to 1 page if limit is 0 or less, or if total is less than limit
	if listCommandFlags.Limit > 0 && total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(listCommandFlags.Limit)))
	}


	fmt.Printf("\nPágina %d de %d. Total de questões correspondentes aos filtros: %d.\n", listCommandFlags.Page, totalPages, total)
}
