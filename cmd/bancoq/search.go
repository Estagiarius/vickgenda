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

// searchCommandFlags holds flag values specifically for the search command.
// It mirrors listCommandFlags for consistency in filtering, pagination, and sorting.
var searchCommandFlags struct {
	Subject      string
	Topic        string
	Difficulty   string
	Type         string
	Author       string
	Tags         []string
	Limit        int
	Page         int
	SortBy       string
	Order        string
	SearchFields []string // Specific to search: fields to search within
}

var bancoqSearchCmd = &cobra.Command{
	Use:   "search <TERMO_DE_BUSCA>",
	Short: "Busca questões por palavras-chave no banco de dados",
	Long: `Realiza uma busca textual por questões no banco de dados utilizando as palavras-chave fornecidas.
A busca pode ser direcionada a campos específicos usando a flag --field.
Filtros adicionais (disciplina, tópico, etc.) podem ser combinados com o termo de busca.
Exemplo:
  vickgenda bancoq search "teorema de pitágoras" --subject "Matemática" --field "question_text" --field "topic"`,
	Args: cobra.ExactArgs(1), // Requer exatamente um argumento para o termo de busca
	Run:  runSearchQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqSearchCmd)

	// Standard filter flags (mirrors list.go for consistency)
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.Subject, "subject", "", "Filtrar por disciplina")
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.Topic, "topic", "", "Filtrar por tópico")
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.Difficulty, "difficulty", "", fmt.Sprintf("Filtrar por dificuldade (%s, %s, %s)", models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard))
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.Type, "type", "", fmt.Sprintf("Filtrar por tipo (%s, %s, %s, %s)", models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer))
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.Author, "author", "", "Filtrar por autor")
	bancoqSearchCmd.Flags().StringSliceVar(&searchCommandFlags.Tags, "tag", []string{}, "Filtrar por tag (atualmente considera apenas a primeira tag fornecida)")

	// Pagination flags
	bancoqSearchCmd.Flags().IntVar(&searchCommandFlags.Limit, "limit", 20, "Número de questões por página")
	bancoqSearchCmd.Flags().IntVar(&searchCommandFlags.Page, "page", 1, "Número da página")

	// Sort flags
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.SortBy, "sort-by", "created_at", "Coluna para ordenação (padrão: created_at)") // 'relevance' could be added if FTS is implemented
	bancoqSearchCmd.Flags().StringVar(&searchCommandFlags.Order, "order", "desc", "Ordem (asc, desc)")

	// Search specific flag
	bancoqSearchCmd.Flags().StringSliceVar(&searchCommandFlags.SearchFields, "field", []string{"question_text", "subject", "topic"}, "Campos para busca textual (ex: question_text, subject, topic, tags, all)")
}

// isValidSearchDifficulty - wrapper for list's validator, allows empty
func isValidSearchDifficulty(difficulty string) bool {
	return isValidListDifficulty(difficulty) // from list.go (or a common validator package)
}

// isValidSearchQuestionType - wrapper for list's validator, allows empty
func isValidSearchQuestionType(qType string) bool {
	return isValidListQuestionType(qType) // from list.go (or a common validator package)
}


func runSearchQuestions(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	searchQuery := args[0]
	if strings.TrimSpace(searchQuery) == "" {
		fmt.Fprintln(os.Stderr, "Erro: O termo de busca não pode ser vazio.")
		os.Exit(1)
	}

	// Validate flags
	if !isValidSearchDifficulty(searchCommandFlags.Difficulty) {
		fmt.Fprintf(os.Stderr, "Valor inválido para --difficulty: '%s'. Use '%s', '%s' ou '%s'.\n", searchCommandFlags.Difficulty, models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard)
		os.Exit(1)
	}
	if !isValidSearchQuestionType(searchCommandFlags.Type) {
		fmt.Fprintf(os.Stderr, "Valor inválido para --type: '%s'. Use '%s', '%s', '%s' ou '%s'.\n", searchCommandFlags.Type, models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer)
		os.Exit(1)
	}
	order := strings.ToLower(searchCommandFlags.Order)
	if order != "asc" && order != "desc" {
		fmt.Fprintf(os.Stderr, "Valor inválido para --order: '%s'. Use 'asc' ou 'desc'.\n", searchCommandFlags.Order)
		os.Exit(1)
	}


	filters := make(map[string]interface{})

	// Standard filters
	if searchCommandFlags.Subject != "" {
		filters["subject"] = searchCommandFlags.Subject
	}
	if searchCommandFlags.Topic != "" {
		filters["topic"] = searchCommandFlags.Topic
	}
	if searchCommandFlags.Difficulty != "" {
		filters["difficulty"] = searchCommandFlags.Difficulty
	}
	if searchCommandFlags.Type != "" {
		filters["question_type"] = searchCommandFlags.Type
	}
	if searchCommandFlags.Author != "" {
		filters["author"] = searchCommandFlags.Author
	}
	if len(searchCommandFlags.Tags) > 0 {
		filters["tags"] = searchCommandFlags.Tags[0] // Consistent with list.go, uses first tag
		if len(searchCommandFlags.Tags) > 1 {
			fmt.Fprintln(os.Stdout, "Aviso: Múltiplas tags foram fornecidas. Atualmente, a filtragem considera apenas a primeira tag:", searchCommandFlags.Tags[0])
		}
	}

	// Search specific filters
	filters["search_query"] = searchQuery

	processedSearchFields := []string{}
	userProvidedAll := false
	for _, f := range searchCommandFlags.SearchFields {
		if strings.ToLower(f) == "all" {
			userProvidedAll = true
			break
		}
		processedSearchFields = append(processedSearchFields, strings.ToLower(f))
	}

	if userProvidedAll {
		// These are the fields db.ListQuestions knows how to search if "search_fields" contains them.
		filters["search_fields"] = []string{"id", "subject", "topic", "question_text", "source", "tags", "author", "difficulty", "question_type"}
	} else if len(processedSearchFields) > 0 {
		filters["search_fields"] = processedSearchFields
	} else {
		// Default if --field is not used or is empty, cobra default should handle this, but as a safeguard:
		filters["search_fields"] = []string{"question_text", "subject", "topic"}
	}

	questions, total, err := db.ListQuestions(filters, searchCommandFlags.SortBy, order, searchCommandFlags.Limit, searchCommandFlags.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao buscar questões: %v\n", err)
		os.Exit(1)
	}

	if total == 0 {
		fmt.Println("Nenhuma questão foi encontrada para o termo de busca e filtros aplicados.")
		return
	}
    if len(questions) == 0 && searchCommandFlags.Page > 1 {
        totalPages := int(math.Ceil(float64(total) / float64(searchCommandFlags.Limit)))
		fmt.Printf("Nenhuma questão encontrada na página %d. Total de páginas: %d.\n", searchCommandFlags.Page, totalPages)
		fmt.Printf("Total de questões no banco que correspondem aos filtros e termo de busca: %d.\n", total)
		return
    }


	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID Curto", "Disciplina", "Tópico", "Tipo", "Dificuldade", "Início da Questão"})
	table.SetBorder(true)
	table.SetRowLine(true)
    table.SetColWidth(60) 
    table.SetAutoWrapText(false)


	for _, q := range questions {
		idShort := q.ID
		if len(q.ID) > 12 {
			idShort = q.ID[:12] + "..."
		}
		questionTextPreview := strings.ReplaceAll(q.QuestionText, "\n", " ")
		maxPreviewLength := 47
		if len(questionTextPreview) > maxPreviewLength {
			runes := []rune(questionTextPreview)
			if len(runes) > maxPreviewLength {
				questionTextPreview = string(runes[:maxPreviewLength-3]) + "..."
			} else {
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

	totalPages := 1
	if searchCommandFlags.Limit > 0 && total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(searchCommandFlags.Limit)))
	}
	fmt.Printf("\nPágina %d de %d. Total de questões correspondentes aos filtros e termo de busca: %d.\n", searchCommandFlags.Page, totalPages, total)
}
