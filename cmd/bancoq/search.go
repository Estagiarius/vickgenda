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
	Short: "Busca questões por palavras-chave no banco de dados",
	Long: `Realiza uma busca textual por questões no banco de dados utilizando as palavras-chave fornecidas.
Por padrão, a busca é realizada nos campos de texto principais da questão (texto, disciplina, tópico).
Utilize a flag --field para especificar outros campos ou 'all' para todos os campos textuais indexados.
Filtros adicionais por disciplina, tópico, dificuldade, etc., podem ser aplicados.`,
	Args: cobra.ExactArgs(1), // Requer exatamente um argumento para o termo de busca
	Run:  runSearchQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqSearchCmd)

	// Standard filter flags (same as list)
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Subject, "subject", "", "Filtrar por disciplina (ex: \"Matemática\")")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Topic, "topic", "", "Filtrar por tópico (ex: \"Álgebra Linear\")")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Difficulty, "difficulty", "", "Filtrar por dificuldade (valores: easy, medium, hard)")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Type, "type", "", "Filtrar por tipo de questão (valores: multiple_choice, true_false, essay, short_answer)")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Author, "author", "", "Filtrar por autor da questão")
	bancoqSearchCmd.Flags().StringSliceVar(&searchFlags.Tags, "tag", []string{}, "Filtrar por tag específica (correspondência exata na lista de tags da questão; pode ser usado múltiplas vezes)")

	// Pagination flags
	bancoqSearchCmd.Flags().IntVar(&searchFlags.Limit, "limit", 20, "Número de questões a serem exibidas por página")
	bancoqSearchCmd.Flags().IntVar(&searchFlags.Page, "page", 1, "Número da página a ser visualizada")

	// Sort flags
	bancoqSearchCmd.Flags().StringVar(&searchFlags.SortBy, "sort-by", "created_at", "Coluna para ordenação (padrão: created_at; 'relevance' pode ser suportado futuramente com FTS)")
	bancoqSearchCmd.Flags().StringVar(&searchFlags.Order, "order", "desc", "Ordem da ordenação (valores: asc, desc)")

	// Search specific flag
	bancoqSearchCmd.Flags().StringSliceVar(&searchFlags.SearchFields, "field", []string{"question_text", "subject", "topic"}, "Campos para realizar a busca textual (ex: question_text, subject, topic, tags, source, author, all)")
}

func runSearchQuestions(cmd *cobra.Command, args []string) {
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	searchQuery := args[0]
	if strings.TrimSpace(searchQuery) == "" {
		fmt.Fprintln(os.Stderr, "Erro: O termo de busca não pode ser vazio.")
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
		fmt.Fprintf(os.Stderr, "Valor inválido para --order: '%s'. Use 'asc' ou 'desc'.\n", searchFlags.Order)
		os.Exit(1)
	}

	questions, total, err := db.ListQuestions(filters, searchFlags.SortBy, order, searchFlags.Limit, searchFlags.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao buscar questões: %v\n", err)
		os.Exit(1)
	}

	if len(questions) == 0 {
		fmt.Println("Nenhuma questão foi encontrada para o termo de busca e filtros aplicados.")
		if total > 0 {
			totalPages := int(math.Ceil(float64(total) / float64(searchFlags.Limit)))
			fmt.Printf("Total de questões no banco que correspondem aos filtros (antes da busca por termo): %d. Total de páginas: %d.\n", total, totalPages)
		}
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	// Table headers are already in Portuguese.
	table.SetHeader([]string{"ID", "Matéria", "Tópico", "Tipo", "Dificuldade", "Início da Questão"})
	table.SetBorder(true)
	table.SetRowLine(true)

	for _, q := range questions {
		idShort := q.ID
		if len(q.ID) > 8 { // Limitar ID para exibição
			idShort = q.ID[:8] + "..."
		}
		questionTextShort := strings.ReplaceAll(q.QuestionText, "\n", " ")
		if len(questionTextShort) > 50 { // Limitar texto para exibição
			runes := []rune(questionTextShort) // Lidar com caracteres multi-byte
			if len(runes) > 50 {
				questionTextShort = string(runes[:47]) + "..."
			} else {
				questionTextShort = string(runes)
			}
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
	if searchFlags.Limit > 0 { // Evitar divisão por zero
		totalPages = int(math.Ceil(float64(total) / float64(searchFlags.Limit)))
	} else if total > 0 { // Se limite não for positivo mas houver itens, considerar 1 página
		totalPages = 1
	}
	fmt.Printf("\nPágina %d de %d. Total de questões correspondentes aos filtros e termo de busca: %d.\n", searchFlags.Page, totalPages, total)
}
