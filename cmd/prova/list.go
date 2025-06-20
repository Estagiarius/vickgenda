package prova

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vickgenda-cli/internal/models" // Assuming models.Test is defined here
)

// Sample tests for simulation
var sampleGeneratedProvas = []models.Test{
	{ID: "prova123", Title: "Prova de Matemática Básica", Subject: "Matemática", CreatedAt: time.Now().Add(-24 * time.Hour), QuestionIDs: []string{"q1", "q2", "q6"}, Instructions: "Leia com atenção."},
	{ID: "prova456", Title: "Avaliação de História do Brasil", Subject: "História", CreatedAt: time.Now().Add(-48 * time.Hour), QuestionIDs: []string{"q3", "q7"}, Instructions: "Responda de forma clara."},
	{ID: "prova789", Title: "Teste Surpresa de Geografia", Subject: "Geografia", CreatedAt: time.Now(), QuestionIDs: []string{"q5"}},
	{ID: "prova101", Title: "Prova Avançada de Cálculo", Subject: "Matemática", CreatedAt: time.Now().Add(-72 * time.Hour), QuestionIDs: []string{"q4", "q10", "q8"}, Instructions: "Justifique suas respostas."},
	{ID: "prova202", Title: "Revisão de Tópicos Matemáticos", Subject: "Matemática", CreatedAt: time.Now().Add(-12 * time.Hour), QuestionIDs: []string{"q1", "q8", "q9"}, Instructions: "Boa sorte!"},
}

// listCmd representa o comando para listar provas geradas.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista as provas geradas",
	Long:  `Exibe uma lista de todas as provas que foram geradas e estão atualmente armazenadas no sistema. Permite filtrar por disciplina e controlar a paginação e ordenação dos resultados.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executando o comando 'prova list' com lógica de simulação...")

		// Recuperar valores das flags
		subjectFilter, _ := cmd.Flags().GetString("subject")
		limit, _ := cmd.Flags().GetInt("limit")
		page, _ := cmd.Flags().GetInt("page")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		order, _ := cmd.Flags().GetString("order")

		// 1. Filtrar provas
		var filteredProvas []models.Test
		if subjectFilter != "" {
			for _, p := range sampleGeneratedProvas {
				if strings.EqualFold(p.Subject, subjectFilter) {
					filteredProvas = append(filteredProvas, p)
				}
			}
		} else {
			filteredProvas = make([]models.Test, len(sampleGeneratedProvas))
			copy(filteredProvas, sampleGeneratedProvas)
		}

		if len(filteredProvas) == 0 {
			fmt.Printf("Nenhuma prova encontrada com o filtro de disciplina: '%s'.\n", subjectFilter)
			return
		}

		// 2. Validar e aplicar ordenação
		validSortBy := []string{"created_at", "title", "subject"}
		isValidSortBy := false
		for _, valid := range validSortBy {
			if sortBy == valid {
				isValidSortBy = true
				break
			}
		}
		if !isValidSortBy {
			fmt.Printf("Critério de ordenação inválido: '%s'. Utilizando 'created_at' como padrão.\n", sortBy)
			sortBy = "created_at"
		}
		if order != "asc" && order != "desc" {
			fmt.Printf("Ordem de classificação inválida: '%s'. Utilizando 'desc' como padrão.\n", order)
			order = "desc"
		}

		sort.Slice(filteredProvas, func(i, j int) bool {
			p1 := filteredProvas[i]
			p2 := filteredProvas[j]
			var less bool
			switch sortBy {
			case "title":
				less = strings.ToLower(p1.Title) < strings.ToLower(p2.Title)
			case "subject":
				less = strings.ToLower(p1.Subject) < strings.ToLower(p2.Subject)
			case "created_at":
				less = p1.CreatedAt.Before(p2.CreatedAt)
			default: // Should not happen due to validation
				less = p1.CreatedAt.Before(p2.CreatedAt)
			}
			if order == "desc" {
				return !less
			}
			return less
		})

		// 3. Aplicar Paginação
		totalProvas := len(filteredProvas)
		if limit <= 0 { // Default limit if not positive
			limit = 10
		}
		if page <= 0 { // Default page if not positive
			page = 1
		}
		startIndex := (page - 1) * limit
		endIndex := startIndex + limit

		if startIndex >= totalProvas {
			fmt.Printf("Página %d fora do alcance. Total de provas: %d (limite por página: %d).\n", page, totalProvas, limit)
			fmt.Println("Nenhuma prova para exibir nesta página.")
			return
		}
		if endIndex > totalProvas {
			endIndex = totalProvas
		}

		provasPaginadas := filteredProvas[startIndex:endIndex]
		totalPages := (totalProvas + limit - 1) / limit


		// 4. Exibir Resultados
		fmt.Printf("\n--- Lista de Provas Geradas (Página %d de %d) ---\n", page, totalPages)
		fmt.Println("----------------------------------------------------------------------------------------------------")
		fmt.Printf("%-10s | %-35s | %-15s | %-20s | %s\n", "ID", "Título", "Disciplina", "Data de Criação", "Nº Questões")
		fmt.Println("----------------------------------------------------------------------------------------------------")

		if len(provasPaginadas) == 0 { // Should be caught by startIndex check, but good for safety
			fmt.Println("Nenhuma prova encontrada para os critérios especificados.")
		} else {
			for _, p := range provasPaginadas {
				fmt.Printf("%-10s | %-35s | %-15s | %-20s | %d\n",
					p.ID,
					truncateString(p.Title, 33),
					truncateString(p.Subject, 13),
					p.CreatedAt.Format("02/01/2006 15:04:05"),
					len(p.QuestionIDs))
			}
		}
		fmt.Println("----------------------------------------------------------------------------------------------------")
		fmt.Printf("Exibindo %d de %d provas. Ordenado por: %s (%s).\n", len(provasPaginadas), totalProvas, sortBy, order)

		fmt.Println("\nComando 'prova list' concluído com lógica de simulação.")
	},
}

// Helper para truncar string e adicionar "..." se for maior que o comprimento máximo
func truncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength-3] + "..."
	}
	return s
}


func init() {
	ProvaCmd.AddCommand(listCmd)
	// Flags para o comando list (baseado em docs/specifications/prova_command_spec.md):
	listCmd.Flags().StringP("subject", "s", "", "Filtrar provas pela disciplina (opcional)")
	listCmd.Flags().IntP("limit", "l", 10, "Limitar o número de provas listadas por página (opcional, padrão: 10)")
	listCmd.Flags().IntP("page", "p", 1, "Número da página a ser exibida (opcional, padrão: 1)")
	listCmd.Flags().String("sort-by", "created_at", "Critério de ordenação (ex: created_at, title, subject) (opcional, padrão: created_at)")
	listCmd.Flags().String("order", "desc", "Ordem de classificação ('asc' para ascendente, 'desc' para descendente) (opcional, padrão: desc)")
}
