package bancoq

import (
	"fmt"
	"math"
	"os"
	// "strconv" // Removed as unused
	"strings"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models" // For potential use, though db returns models.Question

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Struct to hold list command flags
var listFlags struct {
	Subject    string
	Topic      string
	Difficulty string
	Type       string
	Author     string
	Tags       []string // Changed from Tag (string slice) to Tags for clarity
	Limit      int
	Page       int
	SortBy     string
	Order      string
}

var bancoqListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista as questões do banco de questões",
	Long:  `Lista as questões existentes no banco de dados, com opções de filtro, ordenação e paginação.`,
	Run:   runListQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqListCmd)

	// Filter flags
	bancoqListCmd.Flags().StringVar(&listFlags.Subject, "subject", "", "Filtrar por matéria (ex: \"Matemática\")")
	bancoqListCmd.Flags().StringVar(&listFlags.Topic, "topic", "", "Filtrar por tópico (ex: \"Álgebra\")")
	bancoqListCmd.Flags().StringVar(&listFlags.Difficulty, "difficulty", "", "Filtrar por dificuldade (easy, medium, hard)")
	bancoqListCmd.Flags().StringVar(&listFlags.Type, "type", "", "Filtrar por tipo de questão (multiple_choice, true_false, etc.)")
	bancoqListCmd.Flags().StringVar(&listFlags.Author, "author", "", "Filtrar por autor")
	bancoqListCmd.Flags().StringSliceVar(&listFlags.Tags, "tag", []string{}, "Filtrar por tag (pode ser usado múltiplas vezes)")

	// Pagination flags
	bancoqListCmd.Flags().IntVar(&listFlags.Limit, "limit", 20, "Número de questões por página")
	bancoqListCmd.Flags().IntVar(&listFlags.Page, "page", 1, "Número da página")

	// Sort flags
	bancoqListCmd.Flags().StringVar(&listFlags.SortBy, "sort-by", "created_at", "Coluna para ordenação (id, subject, topic, difficulty, question_type, created_at, last_used_at, author)")
	bancoqListCmd.Flags().StringVar(&listFlags.Order, "order", "desc", "Ordem da ordenação (asc, desc)")
}

func runListQuestions(cmd *cobra.Command, args []string) {
	if err := db.InitDB(); err != nil { // Ensure DB is initialized
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	filters := make(map[string]interface{})
	if listFlags.Subject != "" {
		filters["subject"] = listFlags.Subject
	}
	if listFlags.Topic != "" {
		filters["topic"] = listFlags.Topic
	}
	if listFlags.Difficulty != "" {
		// TODO: Validate difficulty value against models.DifficultyEasy, etc.
		filters["difficulty"] = listFlags.Difficulty
	}
	if listFlags.Type != "" {
		// TODO: Validate type value against models.QuestionTypeMultipleChoice, etc.
		filters["question_type"] = listFlags.Type // DB field is question_type
	}
	if listFlags.Author != "" {
		filters["author"] = listFlags.Author
	}
	if len(listFlags.Tags) > 0 {
		// db.ListQuestions expects a single string for tags if it's a simple LIKE search.
		// If db.ListQuestions is enhanced to handle multiple tags (e.g. AND or OR), this might change.
		// For now, if multiple --tag flags are used, we could join them or handle only the first.
		// The current db.ListQuestions `tags LIKE ?` implies it searches for a substring.
		// So, if multiple tags are provided, we might just use the first one for this simple filter.
		// Or, if the intention is "contains any of these tags", then the DB query needs adjustment.
		// For this implementation, let's assume db.ListQuestions handles a slice of tags or we take the first one.
		// The provided db.ListQuestions filters["tags"] = "%"+valStr+"%"
		// This means it only supports one tag string for filtering.
		// We'll pass only the first tag if multiple are given by user for now.
		filters["tags"] = listFlags.Tags[0] // Pass the first tag for now.
                                            // This part of db.ListQuestions might need enhancement for multi-tag filtering.
	}

	// Validate sort order
	order := strings.ToLower(listFlags.Order)
	if order != "asc" && order != "desc" {
		fmt.Fprintf(os.Stderr, "Valor inválido para --order: %s. Use 'asc' ou 'desc'.\n", listFlags.Order)
		os.Exit(1)
	}

	// Basic validation for sort-by column (already handled more robustly in db.ListQuestions)
	// We can add a similar client-side check if desired.

	questions, total, err := db.ListQuestions(filters, listFlags.SortBy, order, listFlags.Limit, listFlags.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao listar questões: %v\n", err)
		os.Exit(1)
	}

	if len(questions) == 0 {
		fmt.Println("Nenhuma questão encontrada com os filtros aplicados.")
		if total > 0 { // This case implies current page has no items but others might
			totalPages := int(math.Ceil(float64(total) / float64(listFlags.Limit)))
			fmt.Printf("Total de questões no banco: %d. Total de páginas: %d.\n", total, totalPages)
		}
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Matéria", "Tópico", "Tipo", "Dificuldade", "Início da Questão"})
	table.SetBorder(true) // Set Border to true
	table.SetRowLine(true) // Enable row line

	for _, q := range questions {
		idShort := q.ID
		if len(q.ID) > 8 {
			idShort = q.ID[:8]
		}

		questionTextShort := strings.ReplaceAll(q.QuestionText, "\n", " ")
		if len(questionTextShort) > 50 {
			questionTextShort = questionTextShort[:50] + "..."
		}

		// Translate difficulty and type for display if needed
		displayDifficulty := q.Difficulty
		switch q.Difficulty {
		case models.DifficultyEasy: displayDifficulty = "Fácil"
		case models.DifficultyMedium: displayDifficulty = "Médio"
		case models.DifficultyHard: displayDifficulty = "Difícil"
		}

		displayType := q.QuestionType
		switch q.QuestionType {
		case models.QuestionTypeMultipleChoice: displayType = "Múltipla Escolha"
		case models.QuestionTypeTrueFalse: displayType = "Verdadeiro/Falso"
		case models.QuestionTypeEssay: displayType = "Dissertativa"
		case models.QuestionTypeShortAnswer: displayType = "Resposta Curta"
		}


		row := []string{
			idShort,
			q.Subject,
			q.Topic,
			displayType,
			displayDifficulty,
			questionTextShort,
		}
		table.Append(row)
	}
	table.Render()

	totalPages := 0
	if listFlags.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(listFlags.Limit)))
	}
	fmt.Printf("\nPágina %d de %d. Total de questões correspondentes: %d.\n", listFlags.Page, totalPages, total)
}
