package prova

import (
	"fmt"
	// "os" // Required for os.WriteFile in actual implementation, used for placeholder message here - REMOVED
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vickgenda-cli/internal/models"
)

// --- Reusing sample data structures (similar to view.go) ---

var sampleGeneratedProvasForExport = []models.Test{
	{ID: "exp123", Title: "Prova de Matemática para Exportar", Subject: "Matemática", CreatedAt: time.Now().Add(-24 * time.Hour), QuestionIDs: []string{"qExp1", "qExp2", "qMissingExp"}, Instructions: "Exportar esta prova."},
	{ID: "exp456", Title: "Avaliação de História para Exportar", Subject: "História", CreatedAt: time.Now().Add(-48 * time.Hour), QuestionIDs: []string{"qExp3"}, Instructions: "Verificar formato de exportação."},
}

var sampleQuestionsForExport = []models.Question{
	{ID: "qExp1", Subject: "Matemática", QuestionText: "Qual o resultado de 10+10?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, AnswerOptions: []string{"15", "20", "25"}, CorrectAnswers: []string{"20"}},
	{ID: "qExp2", Subject: "Matemática", QuestionText: "O que é uma equação?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyMedium, CorrectAnswers: []string{"Uma igualdade envolvendo uma ou mais incógnitas."}},
	{ID: "qExp3", Subject: "História", QuestionText: "Quem foi o primeiro presidente do Brasil?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyHard, AnswerOptions: []string{"Deodoro da Fonseca", "Prudente de Morais"}, CorrectAnswers: []string{"Deodoro da Fonseca"}},
}

// Helper to find a Test by ID
func findTestByIDForExport(id string, tests []models.Test) *models.Test {
	for i := range tests {
		if tests[i].ID == id {
			return &tests[i]
		}
	}
	return nil
}

// Helper to find a Question by ID
func findQuestionByIDForExport(id string, questions []models.Question) *models.Question {
	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}
	return nil
}
// --- End of sample data ---

// exportCmd representa o comando para exportar uma prova.
var exportCmd = &cobra.Command{
	Use:   "export <id_prova> <filepath>",
	Short: "Exporta uma prova para um arquivo",
	Long:  `Salva os dados de uma prova específica em um arquivo, no formato especificado (ex: TXT, JSON, PDF). Inclui todas as questões e, opcionalmente, suas respostas.`,
	Args:  cobra.ExactArgs(2), // Espera dois argumentos: ID da prova e caminho do arquivo.
	Run: func(cmd *cobra.Command, args []string) {
		provaID := args[0]   // Validated by cobra.ExactArgs(2)
		filepath := args[1] // Validated by cobra.ExactArgs(2)

		exportFormat, _ := cmd.Flags().GetString("format") // Renamed to avoid conflict
		showAnswers, _ := cmd.Flags().GetBool("show-answers")

		fmt.Printf("Executando o comando 'prova export' para a Prova ID: %s\n", provaID)
		fmt.Printf("Caminho do arquivo de saída: %s, Formato: %s, Incluir Respostas: %t\n", filepath, exportFormat, showAnswers)

		// 1. Validar formato de exportação (focando em 'txt')
		if exportFormat != "txt" {
			// For other formats like json, pdf, html, specific libraries or logic would be needed.
			// For now, we'll warn and default to a text-like representation if attempting others.
			fmt.Printf("AVISO: O formato de exportação '%s' não é totalmente suportado para escrita em arquivo nesta simulação. O conteúdo gerado será textual.\n", exportFormat)
			// Potentially default to .txt extension for filepath if format is not txt.
		}

		// 2. Encontrar a prova
		prova := findTestByIDForExport(provaID, sampleGeneratedProvasForExport)
		if prova == nil {
			fmt.Printf("Erro: Prova com ID '%s' não foi encontrada.\n", provaID)
			return
		}

		// 3. Buscar detalhes das questões
		var fetchedQuestions []*models.Question
		questionsMap := make(map[string]*models.Question)

		// fmt.Println("\n[Simulação] Buscando detalhes das questões da prova para exportação...")
		for _, qID := range prova.QuestionIDs {
			if _, exists := questionsMap[qID]; exists {
				continue
			}
			question := findQuestionByIDForExport(qID, sampleQuestionsForExport)
			if question != nil {
				fetchedQuestions = append(fetchedQuestions, question)
				questionsMap[qID] = question // Store in map for ordered access later
			} else {
				fmt.Printf("AVISO: A questão com ID '%s' (listada na prova) não foi encontrada no banco de questões de simulação.\n", qID)
				// Create a placeholder to maintain order and indicate missing data
				placeholderQuestion := &models.Question{ID: qID, QuestionText: fmt.Sprintf("[Questão com ID '%s' não encontrada]", qID), QuestionType: "desconhecido"}
				fetchedQuestions = append(fetchedQuestions, placeholderQuestion)
				questionsMap[qID] = placeholderQuestion
			}
		}

		// Reorder fetchedQuestions according to prova.QuestionIDs
		orderedFetchedQuestions := make([]*models.Question, len(prova.QuestionIDs))
		for i, qID := range prova.QuestionIDs {
			orderedFetchedQuestions[i] = questionsMap[qID]
		}


		// 4. Formatar conteúdo para exportação (usando a nova função)
		// Passando `orderedFetchedQuestions` para manter a ordem da prova.
		formattedContent, err := formatTestContentForExport(prova, orderedFetchedQuestions, exportFormat, showAnswers)
		if err != nil {
			fmt.Printf("Erro ao formatar o conteúdo da prova: %v\n", err)
			return
		}

		// 5. Simular Escrita em Arquivo
		fmt.Printf("\n--- Simulação de Exportação para Arquivo ---\n")
		fmt.Printf("Arquivo: %s\n", filepath)
		fmt.Printf("Formato: %s\n", exportFormat)
		fmt.Println("Conteúdo (primeiras ~250 caracteres):")

		contentPreview := formattedContent
		if len(contentPreview) > 250 {
			contentPreview = contentPreview[:250] + "..."
		}
		fmt.Println(contentPreview)
		fmt.Println("-----------------------------------------")

		// Placeholder para a escrita real do arquivo:
		// if exportFormat == "txt" { // Ou outros formatos baseados em texto
		// 	err := os.WriteFile(filepath, []byte(formattedContent), 0644)
		// 	if err != nil {
		// 		fmt.Fprintf(os.Stderr, "Erro (simulado) ao escrever prova no arquivo '%s': %v\n", filepath, err)
		// 		return // ou cmd.SilenceUsage = true; return err
		// 	}
		// 	fmt.Printf("\nProva exportada com sucesso para: %s\n", filepath)
		// } else if exportFormat == "json" {
		//    // jsonBytes, _ := json.MarshalIndent(provaComQuestoes, "", "  ")
		//    // os.WriteFile(filepath, jsonBytes, 0644) ...
		// } // etc. for pdf, html

		fmt.Println("\nComando 'prova export' concluído com lógica de simulação.")
	},
}

// Helper function to format test content (similar to view.go logic)
// This could be moved to a shared formatter package later.
func formatTestContentForExport(test *models.Test, questions []*models.Question, formatType string, showAnswers bool) (string, error) {
	if test == nil {
		return "", fmt.Errorf("prova (test) não pode ser nula")
	}

	// For now, only "txt" is implemented for direct string generation
	if formatType != "txt" {
		// In a real scenario, you might return an error or handle other formats.
		// For this simulation, we'll just note it and proceed with txt-like formatting.
		// No need to print here, calling function can decide.
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Título: %s\n", test.Title))
	sb.WriteString(fmt.Sprintf("Disciplina: %s\n", test.Subject))
	if test.Instructions != "" {
		sb.WriteString(fmt.Sprintf("Instruções: %s\n", test.Instructions))
	}
	if test.RandomizationSeed > 0 {
		sb.WriteString(fmt.Sprintf("(Randomizado com semente: %d)\n", test.RandomizationSeed))
	}
	sb.WriteString("------------------------------------\n\n")

	if len(questions) == 0 { // Based on the actual questions passed
		sb.WriteString("Esta prova não contém questões ou as questões não puderam ser carregadas.\n")
	} else {
		for i, question := range questions {
			if question == nil { // Should not happen if list is pre-filtered or placeholders used
				sb.WriteString(fmt.Sprintf("%d. [Erro - Questão nula na lista]\n\n", i+1))
				continue
			}
			sb.WriteString(fmt.Sprintf("%d. (ID: %s) %s\n", i+1, question.ID, question.QuestionText))

			if strings.HasPrefix(question.QuestionText, "[Questão com ID") && strings.HasSuffix(question.QuestionText, "não encontrada]") {
				sb.WriteString("\n"); // Extra space for missing questions
				continue;
			}


			if question.QuestionType == models.QuestionTypeMultipleChoice {
				if len(question.AnswerOptions) > 0 {
					// sb.WriteString("   Opções:\n") // Removed for cleaner export
					for j, opt := range question.AnswerOptions {
						prefix := fmt.Sprintf("  %c)", 'A'+j)
						marker := "" // No marker for plain export unless answers are shown
						if showAnswers {
							marker = "[ ]" // Default for showAnswers
							for _, correctAns := range question.CorrectAnswers {
								// Assuming CorrectAnswers for MC stores the text of the correct option
								if strings.EqualFold(opt, correctAns) {
									marker = "[*]"
									break
								}
							}
						}
						sb.WriteString(fmt.Sprintf("%s %s %s\n", prefix, marker, opt))
					}
				} else {
					// sb.WriteString("   AVISO: Questão de múltipla escolha sem opções definidas.\n")
				}
			}

			if showAnswers {
				if len(question.CorrectAnswers) > 0 {
					if question.QuestionType == models.QuestionTypeEssay || (question.QuestionType == models.QuestionTypeMultipleChoice && len(question.AnswerOptions) == 0) {
						sb.WriteString(fmt.Sprintf("   Resposta Correta: %s\n", strings.Join(question.CorrectAnswers, " | ")))
					} else if question.QuestionType == models.QuestionTypeMultipleChoice && len(question.AnswerOptions) > 0 && markerHasNotShownAllAnswers(question) {
						// If markers didn't show all, or for a summary
						sb.WriteString(fmt.Sprintf("   Gabarito: %s\n", strings.Join(question.CorrectAnswers, " | ")))
					}
				} else if question.QuestionType != models.QuestionTypeMultipleChoice || len(question.AnswerOptions) == 0 { // Avoid for MC with options if no correct answer is set
					// sb.WriteString("   AVISO: Resposta correta não disponível.\n")
				}
			}
			sb.WriteString("\n") // Add a blank line after each question for readability
		}
	}
	sb.WriteString("------------------------------------\n")
	return sb.String(), nil
}

// Helper to check if all correct answers for MC were already marked (e.g. if CorrectAnswers contains option letters not texts)
// This is a placeholder, actual logic might be more complex if CorrectAnswers stores A,B,C instead of full text.
func markerHasNotShownAllAnswers(q *models.Question) bool {
	// If CorrectAnswers stores text and we marked based on text, this might be true.
	// If CorrectAnswers stores letters 'A', 'B', etc., then markers are the primary way.
	// For this simulation, assume if CorrectAnswers has entries, and we are showing answers,
	// we might want to list them explicitly if they are not just option texts.
	return true // Simplified: always show explicit gabarito if available and showAnswers is true for MC
}


func init() {
	ProvaCmd.AddCommand(exportCmd)
	// Flags para o comando export (baseado em docs/specifications/prova_command_spec.md):
	exportCmd.Flags().StringP("format", "f", "txt", "Formato do arquivo de exportação (ex: txt, json, pdf, html) (opcional, padrão: txt)")
	exportCmd.Flags().Bool("show-answers", false, "Incluir as respostas das questões no arquivo exportado (opcional, padrão: false)")
	// A flag "template" foi mencionada no setup inicial mas não no doc. Mantendo as do doc.
	// exportCmd.Flags().String("template", "", "Caminho para um template customizado de exportação (ex: para PDF ou HTML)")
}
