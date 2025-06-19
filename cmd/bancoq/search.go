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

var searchFlags struct {
	Subject      string
	Topic        string
	Difficulty   string
	Type         string
	Author       string
	Tags         []string // For --tag filter
	Limit        int
	Page         int
	SortBy       string
	Order        string
	SearchFields []string // For --field flag
}

var bancoqSearchCmd = &cobra.Command{
	Use:   "search <TERMO_DE_BUSCA>",
	Short: "Procura questões por palavras-chave",
	Long: `Procura questões no banco de dados por palavras-chave em campos de texto especificados.
Por padrão, busca em "question_text", "subject" e "topic". Use --field para especificar outros campos.`,
	Args: cobra.ExactArgs(1),
	Run:  runSearchQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqSearchCmd)

	// Standard filter flags (same as list)
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Subject, "subject", "", "Filtrar por matéria (ex: \"Matemática\")")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Topic, "topic", "", "Filtrar por tópico (ex: \"Álgebra\")")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Difficulty, "difficulty", "", "Filtrar por dificuldade (easy, medium, hard)")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Type, "type", "", "Filtrar por tipo de questão (multiple_choice, etc.)")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Author, "author", "", "Filtrar por autor")
	bancoqSearchCmd.Flags().StringSliceVar(&searchFlags.Tags, "tag", []string{}, "Filtrar por tag específica (correspondência exata na lista de tags da questão)") // This --tag is for specific filtering

	// Pagination flags
	bancoqSearchCmd.Flags().IntVar(&searchFlags.Limit, "limit", 20, "Número de questões por página")
	bancoqSearchCmd.Flags().IntVar(&searchFlags.Page, "page", 1, "Número da página")

	// Sort flags
	bancoqSearchCmd.Flags().StringVar(&searchFlags.SortBy, "sort-by", "created_at", "Coluna para ordenação") // Default might change to relevance score if FTS is used later
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Order, "order", "desc", "Ordem da ordenação (asc, desc)")

	// Search specific flag
	bancoqSearchCmd.Flags().StringSliceVar(&searchFlags.SearchFields, "field", []string{"question_text", "subject", "topic"}, "Campos para buscar (ex: question_text, subject, topic, tags, source, author, all)")
}

func runSearchQuestions(cmd *cobra.Command, args []string) {
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	searchQuery := args[0]
	if strings.TrimSpace(searchQuery) == "" {
		fmt.Fprintln(os.Stderr, "Erro: Termo de busca não pode ser vazio.")
		os.Exit(1)
	}

	filters := make(map[string]interface{})

	// Standard filters
	if searchFlags.Subject != "" {
		filters["subject"] = searchFlags.Subject
	}
	if searchFlags.Topic != "" {
		filters["topic"] = searchFlags.Topic
	}
	if searchFlags.Difficulty != "" {
		filters["difficulty"] = searchFlags.Difficulty
	}
	if searchFlags.Type != "" {
		filters["question_type"] = searchFlags.Type // DB field is question_type
	}
	if searchFlags.Author != "" {
		filters["author"] = searchFlags.Author
	}
	if len(searchFlags.Tags) > 0 {
		// The `db.ListQuestions` handles `filters["tags"]` with a LIKE clause: `tags LIKE %tagValue%`
		// If multiple --tag flags are provided, we might want to pass them all and let db layer decide how to combine (e.g. AND or OR)
		// For now, consistent with `list`, we'll pass the first one if that's how db.ListQuestions handles it.
		// The current db.ListQuestions implementation for `case "tags":` uses `"%"+valStr+"%"` which means it expects a single string.
		filters["tags"] = searchFlags.Tags[0]
	}

	// Search specific filters
	filters["search_query"] = searchQuery

	processedSearchFields := []string{}
	userProvidedAll := false
	for _, f := range searchFlags.SearchFields {
		if strings.ToLower(f) == "all" {
			userProvidedAll = true
			break
		}
		// Basic validation for field names (can correspond to db.ListQuestions's validSearchFields)
		// For simplicity, we trust user input here or rely on db.ListQuestions's own validation.
		processedSearchFields = append(processedSearchFields, strings.ToLower(f))
	}

	if userProvidedAll {
		filters["search_fields"] = []string{"question_text", "subject", "topic", "tags", "source", "author"} // Default "all" fields
	} else if len(processedSearchFields) > 0 {
		filters["search_fields"] = processedSearchFields
	} else {
		// Default if --field is not used or is empty after processing (though cobra default helps)
		filters["search_fields"] = []string{"question_text", "subject", "topic"}
	}


	order := strings.ToLower(searchFlags.Order)
	if order != "asc" && order != "desc" {
		fmt.Fprintf(os.Stderr, "Valor inválido para --order: %s. Use 'asc' ou 'desc'.\n", searchFlags.Order)
		os.Exit(1)
	}

	questions, total, err := db.ListQuestions(filters, searchFlags.SortBy, order, searchFlags.Limit, searchFlags.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao procurar questões: %v\n", err)
		os.Exit(1)
	}

	if len(questions) == 0 {
		fmt.Println("Nenhuma questão encontrada para o termo de busca e filtros aplicados.")
		if total > 0 {
			totalPages := int(math.Ceil(float64(total) / float64(searchFlags.Limit)))
			fmt.Printf("Total de questões no banco que correspondem aos filtros (antes da busca por termo): %d. Total de páginas: %d.\n", total, totalPages)
		}
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Matéria", "Tópico", "Tipo", "Dificuldade", "Início da Questão"})
	table.SetBorder(true)
	table.SetRowLine(true)

	for _, q := range questions {
		idShort := q.ID
		if len(q.ID) > 8 {
			idShort = q.ID[:8]
		}
		questionTextShort := strings.ReplaceAll(q.QuestionText, "\n", " ")
		if len(questionTextShort) > 50 {
			questionTextShort = questionTextShort[:50] + "..."
		}
		row := []string{
			idShort,
			q.Subject,
			q.Topic,
			models.FormatQuestionTypeToPtBR(q.QuestionType),
			models.FormatDifficultyToPtBR(q.Difficulty),
			questionTextShort,
		}
		table.Append(row)
	}
	table.Render()

	totalPages := 0
	if searchFlags.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(searchFlags.Limit)))
	}
	fmt.Printf("\nPágina %d de %d. Total de questões correspondentes: %d.\n", searchFlags.Page, totalPages, total)
}
