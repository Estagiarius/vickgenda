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
	Long:  `Exibe uma lista paginada das questões armazenadas no banco de dados. Permite aplicar diversos filtros para refinar a busca e ordenar os resultados conforme especificado.`,
	Run:   runListQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqListCmd)

	// Filter flags
	bancoqListCmd.Flags().StringVar(&listFlags.Subject, "subject", "", "Filtrar por disciplina (ex: \"Matemática\")")
	bancoqListCmd.Flags().StringVar(&listFlags.Topic, "topic", "", "Filtrar por tópico (ex: \"Álgebra Linear\")")
	bancoqListCmd.Flags().StringVar(&listFlags.Difficulty, "difficulty", "", "Filtrar por dificuldade (valores: easy, medium, hard)")
	bancoqListCmd.Flags().StringVar(&listFlags.Type, "type", "", "Filtrar por tipo de questão (valores: multiple_choice, true_false, essay, short_answer)")
	bancoqListCmd.Flags().StringVar(&listFlags.Author, "author", "", "Filtrar por autor da questão")
	bancoqListCmd.Flags().StringSliceVar(&listFlags.Tags, "tag", []string{}, "Filtrar por tag (pode ser usado múltiplas vezes para buscar questões com qualquer uma das tags)")

	// Pagination flags
	bancoqListCmd.Flags().IntVar(&listFlags.Limit, "limit", 20, "Número de questões a serem exibidas por página")
	bancoqListCmd.Flags().IntVar(&listFlags.Page, "page", 1, "Número da página a ser visualizada")

	// Sort flags
	bancoqListCmd.Flags().StringVar(&listFlags.SortBy, "sort-by", "created_at", "Coluna para ordenação (opções: id, subject, topic, difficulty, question_type, created_at, last_used_at, author)")
	bancoqListCmd.Flags().StringVar(&listFlags.Order, "order", "desc", "Ordem da ordenação (valores: asc, desc)")
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
		// TODO: Validar valor de dificuldade contra models.DifficultyEasy, etc.
		filters["difficulty"] = listFlags.Difficulty
	}
	if listFlags.Type != "" {
		// TODO: Validar valor de tipo contra models.QuestionTypeMultipleChoice, etc.
		filters["question_type"] = listFlags.Type // Campo no BD é question_type
	}
	if listFlags.Author != "" {
		filters["author"] = listFlags.Author
	}
	if len(listFlags.Tags) > 0 {
		// db.ListQuestions espera uma string única para tags se for uma busca LIKE simples.
		// Se db.ListQuestions for melhorado para tratar múltiplas tags (ex: AND ou OR), isto pode mudar.
		// Por enquanto, se múltiplas --tag flags são usadas, podemos juntá-las ou usar apenas a primeira.
		// A query atual `tags LIKE ?` implica busca por substring.
		// Então, se múltiplas tags são providas, usaremos apenas a primeira por agora.
		// Esta parte de db.ListQuestions pode precisar de melhoria para filtragem multi-tag.
		filters["tags"] = listFlags.Tags[0] // Passa apenas a primeira tag por enquanto.
	}

	// Validate sort order
	order := strings.ToLower(listFlags.Order)
	if order != "asc" && order != "desc" {
		fmt.Fprintf(os.Stderr, "Valor inválido para --order: '%s'. Use 'asc' ou 'desc'.\n", listFlags.Order)
		os.Exit(1)
	}

	// Basic validation for sort-by column (already handled more robustly in db.ListQuestions)
	// We can add a similar client-side check if desired.

	questions, total, err := db.ListQuestions(filters, listFlags.SortBy, order, listFlags.Limit, listFlags.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao listar as questões: %v\n", err)
		os.Exit(1)
	}

	if len(questions) == 0 {
		fmt.Println("Nenhuma questão foi encontrada com os filtros aplicados.")
		if total > 0 { // This case implies current page has no items but others might
			totalPages := int(math.Ceil(float64(total) / float64(listFlags.Limit)))
			fmt.Printf("Total de questões no banco que correspondem aos filtros (fora da paginação atual): %d. Total de páginas: %d.\n", total, totalPages)
		}
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Disciplina", "Tópico", "Tipo", "Dificuldade", "Início da Questão"}) // Mantidos em Português conforme arquivo
	table.SetBorder(true)
	table.SetRowLine(true)

	for _, q := range questions {
		idShort := q.ID
		if len(q.ID) > 8 { // Limitar tamanho do ID exibido para não quebrar a tabela
			idShort = q.ID[:8] + "..."
		}

		questionTextShort := strings.ReplaceAll(q.QuestionText, "\n", " ")
		if len(questionTextShort) > 50 { // Limitar tamanho do texto da questão exibido
			runes := []rune(questionTextShort) // Lidar com caracteres multi-byte
			if len(runes) > 50 {
				questionTextShort = string(runes[:47]) + "..."
			} else {
				questionTextShort = string(runes)
			}
		}

		// A tradução para exibição já é feita no original, mantendo.
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
	if listFlags.Limit > 0 { // Evitar divisão por zero se limit for 0 ou negativo
		totalPages = int(math.Ceil(float64(total) / float64(listFlags.Limit)))
	} else if total > 0 { // Se limit não é positivo mas há itens, considerar como 1 página gigante
		totalPages = 1
	}

	fmt.Printf("\nPágina %d de %d. Total de questões correspondentes aos filtros: %d.\n", listFlags.Page, totalPages, total)
}
