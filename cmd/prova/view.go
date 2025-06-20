package prova

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vickgenda-cli/internal/models"
)

// --- Reusing sample data structures (similar to list.go and generate.go) ---

// Sample tests for simulation (can be shared or redefined if state is an issue)
var viewSampleGeneratedProvas = []models.Test{
	{ID: "prova123", Title: "Prova de Matemática Básica", Subject: "Matemática", CreatedAt: time.Now().Add(-24 * time.Hour), QuestionIDs: []string{"q1", "q2", "q6", "qMissing"}, Instructions: "Leia com atenção cada questão."},
	{ID: "prova456", Title: "Avaliação de História do Brasil", Subject: "História", CreatedAt: time.Now().Add(-48 * time.Hour), QuestionIDs: []string{"q3", "q7"}, Instructions: "Responda de forma clara e objetiva."},
	{ID: "prova789", Title: "Teste Surpresa de Geografia", Subject: "Geografia", CreatedAt: time.Now(), QuestionIDs: []string{"q5"}},
}

// Sample questions for simulation (can be shared or redefined)
var viewSampleQuestions = []models.Question{
	{ID: "q1", Subject: "Matemática", QuestionText: "Quanto é 2+2?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, Topic: "aritmética", Tags: []string{"básica"}, AnswerOptions: []string{"3", "4", "5"}, CorrectAnswers: []string{"4"}},
	{ID: "q2", Subject: "Matemática", QuestionText: "Quanto é 5*8?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, Topic: "aritmética", Tags: []string{"básica"}, AnswerOptions: []string{"30", "40", "35"}, CorrectAnswers: []string{"40"}},
	{ID: "q3", Subject: "História", QuestionText: "Quem descobriu o Brasil?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyMedium, Topic: "descobrimentos", Tags: []string{"Brasil"}, CorrectAnswers: []string{"Pedro Álvares Cabral"}},
	{ID: "q4", Subject: "Matemática", QuestionText: "Qual a derivada de x^2?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyHard, Topic: "cálculo", Tags: []string{"avançada"}, CorrectAnswers: []string{"2x"}},
	{ID: "q5", Subject: "Geografia", QuestionText: "Qual a capital da França?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, Topic: "europa", Tags: []string{"capitais"}, AnswerOptions: []string{"Londres", "Paris", "Madri"}, CorrectAnswers: []string{"Paris"}},
	{ID: "q6", Subject: "Matemática", QuestionText: "Resolva a equação: x + 5 = 10", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyEasy, Topic: "algebra", Tags: []string{"equação"}, CorrectAnswers: []string{"x = 5"}},
	{ID: "q7", Subject: "História", QuestionText: "Em que ano começou a Segunda Guerra Mundial?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyMedium, Topic: "guerras mundiais", Tags: []string{"século XX"}, AnswerOptions: []string{"1939", "1941", "1945"}, CorrectAnswers: []string{"1939"}},
	{ID: "q8", Subject: "Matemática", QuestionText: "Qual o valor de Pi (aproximado)?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyMedium, Topic: "geometria", Tags: []string{"constantes"}, AnswerOptions: []string{"3.14", "3.12", "3.16"}, CorrectAnswers: []string{"3.14"}},
}

// Helper to find a Test by ID
func findTestByID(id string, tests []models.Test) *models.Test {
	for i := range tests {
		if tests[i].ID == id {
			return &tests[i]
		}
	}
	return nil
}

// Helper to find a Question by ID
func findViewQuestionByID(id string, questions []models.Question) *models.Question {
	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}
	return nil
}
// --- End of sample data structures ---

// viewCmd representa o comando para visualizar uma prova específica.
var viewCmd = &cobra.Command{
	Use:   "view <id_prova>",
	Short: "Visualiza os detalhes de uma prova específica",
	Long:  `Carrega e exibe todas as informações de uma prova específica, incluindo suas questões, com base no ID fornecido. Permite formatar a saída e opcionalmente mostrar as respostas.`,
	Args:  cobra.ExactArgs(1), // Espera exatamente um argumento: o ID da prova.
	Run: func(cmd *cobra.Command, args []string) {
		provaID := args[0] // Already validated by cobra.ExactArgs(1)
		showAnswers, _ := cmd.Flags().GetBool("show-answers")
		outputFormat, _ := cmd.Flags().GetString("output-format")

		fmt.Printf("Executando o comando 'prova view' para a Prova ID: %s (Formato: %s, Mostrar Respostas: %t)\n",
			provaID, outputFormat, showAnswers)

		// 1. Validar output-format (focando em 'txt')
		if outputFormat != "txt" {
			fmt.Printf("AVISO: Formato de saída '%s' ainda não é totalmente suportado. Exibindo em formato de texto.\n", outputFormat)
			// outputFormat = "txt" // Forçar para txt se quisermos ser estritos
		}

		// 2. Encontrar a prova
		prova := findTestByID(provaID, viewSampleGeneratedProvas)
		if prova == nil {
			fmt.Printf("Erro: Prova com ID '%s' não foi encontrada.\n", provaID)
			return
		}

		// 3. Preparar para buscar questões
		var fetchedQuestions []*models.Question
		questionsFoundMap := make(map[string]*models.Question) // Para fácil acesso e evitar duplicatas se ID for repetido

		fmt.Println("\n[Simulação] Buscando detalhes das questões da prova...")
		for _, qID := range prova.QuestionIDs {
			if _, exists := questionsFoundMap[qID]; exists { // Evitar buscar a mesma questão múltiplas vezes se ID estiver duplicado na prova
				continue
			}
			question := findViewQuestionByID(qID, viewSampleQuestions)
			if question != nil {
				fetchedQuestions = append(fetchedQuestions, question)
				questionsFoundMap[qID] = question
			} else {
				fmt.Printf("AVISO: A questão com ID '%s' (listada na prova) não foi encontrada no banco de questões de simulação.\n", qID)
				// Adicionar um placeholder ou tratar como erro crítico dependendo do requisito
				fetchedQuestions = append(fetchedQuestions, &models.Question{ID: qID, QuestionText: fmt.Sprintf("Questão com ID '%s' não encontrada.", qID), QuestionType: "desconhecido"})
			}
		}

		// 4. Formatar e Exibir Saída (formato TXT)
		fmt.Printf("\n--- Detalhes da Prova: %s ---\n", prova.Title)
		fmt.Printf("ID da Prova: %s\n", prova.ID)
		fmt.Printf("Disciplina: %s\n", prova.Subject)
		fmt.Printf("Data de Criação: %s\n", prova.CreatedAt.Format("02/01/2006 15:04:05"))
		if prova.Instructions != "" {
			fmt.Printf("Instruções: %s\n", prova.Instructions)
		}
		if prova.RandomizationSeed > 0 {
			fmt.Printf("Semente de Randomização: %d\n", prova.RandomizationSeed)
		}
		fmt.Println("------------------------------------")

		if len(prova.QuestionIDs) == 0 {
			fmt.Println("Esta prova não contém questões.")
		} else {
			fmt.Printf("\n--- Questões (%d) ---\n", len(prova.QuestionIDs))
			for i, qID := range prova.QuestionIDs {
				question := questionsFoundMap[qID] // Usar o mapa para pegar a questão na ordem correta
				if question == nil { // Segurança, embora o loop anterior deva popular
					fmt.Printf("\n%d. Questão ID '%s': [ERRO INTERNO - Detalhes não puderam ser carregados]\n", i+1, qID)
					continue
				}

				fmt.Printf("\n%d. (ID: %s) %s\n", i+1, question.ID, question.QuestionText)
				fmt.Printf("   Tipo: %s, Dificuldade: %s, Tópico: %s, Tags: %s\n",
					models.FormatQuestionTypeToPtBR(question.QuestionType), models.FormatDifficultyToPtBR(question.Difficulty), question.Topic, strings.Join(question.Tags, ", "))

				if question.QuestionType == models.QuestionTypeMultipleChoice || question.QuestionType == models.QuestionTypeTrueFalse {
					if len(question.AnswerOptions) > 0 {
						fmt.Println("   Opções:")
						for j, opt := range question.AnswerOptions {
							prefix := fmt.Sprintf("     %c)", 'A'+j)
							if showAnswers {
								isCorrectOption := false
								for _, correctAns := range question.CorrectAnswers {
									if strings.EqualFold(opt, correctAns) {
										isCorrectOption = true
										break
									}
								}
								if isCorrectOption {
									prefix += " [*]"
								} else {
									prefix += " [ ]"
								}
							}
							fmt.Printf("%s %s\n", prefix, opt)
						}
					} else {
						fmt.Println("   AVISO: Questão de múltipla escolha sem opções definidas.")
					}
				}

				if showAnswers {
					if len(question.CorrectAnswers) > 0 {
						if question.QuestionType == models.QuestionTypeEssay || (question.QuestionType == models.QuestionTypeMultipleChoice && len(question.AnswerOptions) == 0) {
							fmt.Printf("   Resposta(s) Correta(s): %s\n", strings.Join(question.CorrectAnswers, " | "))
						} else if question.QuestionType == models.QuestionTypeMultipleChoice && len(question.AnswerOptions) > 0 {
							// fmt.Printf("   Gabarito (texto): %s\n", strings.Join(question.CorrectAnswers, " | "))
						}
					} else {
						fmt.Println("   AVISO: Resposta(s) correta(s) não disponível(is) para esta questão.")
					}
				}
			}
		}
		fmt.Println("\n------------------------------------")
		fmt.Println("\nComando 'prova view' concluído com lógica de simulação.")
	},
}

func init() {
	ProvaCmd.AddCommand(viewCmd)
	// Flags para o comando view (baseado em docs/specifications/prova_command_spec.md):
	viewCmd.Flags().BoolP("show-answers", "a", false, "Exibir as respostas das questões na visualização (opcional, padrão: false)")
	viewCmd.Flags().StringP("output-format", "f", "txt", "Formato de saída para a visualização (ex: txt, json, markdown) (opcional, padrão: txt)")
}
