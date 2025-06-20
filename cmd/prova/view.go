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
	{ID: "q1", Subject: "Matemática", Text: "Quanto é 2+2?", Type: "multipla_escolha", Difficulty: "fácil", Topics: []string{"aritmética"}, Tags: []string{"básica"}, Options: []models.QuestionOption{{Text: "3", IsCorrect: false}, {Text: "4", IsCorrect: true}, {Text: "5", IsCorrect: false}}, CorrectAnswers: []string{"4"}},
	{ID: "q2", Subject: "Matemática", Text: "Quanto é 5*8?", Type: "multipla_escolha", Difficulty: "fácil", Topics: []string{"aritmética"}, Tags: []string{"básica"}, Options: []models.QuestionOption{{Text: "30", IsCorrect: false}, {Text: "40", IsCorrect: true}, {Text: "35", IsCorrect: false}}, CorrectAnswers: []string{"40"}},
	{ID: "q3", Subject: "História", Text: "Quem descobriu o Brasil?", Type: "dissertativa", Difficulty: "médio", Topics: []string{"descobrimentos"}, Tags: []string{"Brasil"}, CorrectAnswers: []string{"Pedro Álvares Cabral"}},
	{ID: "q4", Subject: "Matemática", Text: "Qual a derivada de x^2?", Type: "dissertativa", Difficulty: "difícil", Topics: []string{"cálculo"}, Tags: []string{"avançada"}, CorrectAnswers: []string{"2x"}},
	{ID: "q5", Subject: "Geografia", Text: "Qual a capital da França?", Type: "multipla_escolha", Difficulty: "fácil", Topics: []string{"europa"}, Tags: []string{"capitais"}, Options: []models.QuestionOption{{Text: "Londres", IsCorrect: false}, {Text: "Paris", IsCorrect: true}, {Text: "Madri", IsCorrect: false}}, CorrectAnswers: []string{"Paris"}},
	{ID: "q6", Subject: "Matemática", Text: "Resolva a equação: x + 5 = 10", Type: "dissertativa", Difficulty: "fácil", Topics: []string{"algebra"}, Tags: []string{"equação"}, CorrectAnswers: []string{"x = 5"}},
	{ID: "q7", Subject: "História", Text: "Em que ano começou a Segunda Guerra Mundial?", Type: "multipla_escolha", Difficulty: "médio", Topics: []string{"guerras mundiais"}, Tags: []string{"século XX"}, Options: []models.QuestionOption{{Text: "1939", IsCorrect: true}, {Text: "1941", IsCorrect: false}, {Text: "1945", IsCorrect: false}}, CorrectAnswers: []string{"1939"}},
	{ID: "q8", Subject: "Matemática", Text: "Qual o valor de Pi (aproximado)?", Type: "multipla_escolha", Difficulty: "médio", Topics: []string{"geometria"}, Tags: []string{"constantes"}, Options: []models.QuestionOption{{Text: "3.14", IsCorrect: true}, {Text: "3.12", IsCorrect: false}, {Text: "3.16", IsCorrect: false}}, CorrectAnswers: []string{"3.14"}},
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
				fetchedQuestions = append(fetchedQuestions, &models.Question{ID: qID, Text: fmt.Sprintf("Questão com ID '%s' não encontrada.", qID), Type: "desconhecido"})
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

				fmt.Printf("\n%d. (ID: %s) %s\n", i+1, question.ID, question.Text)
				fmt.Printf("   Tipo: %s, Dificuldade: %s, Tópicos: %s, Tags: %s\n",
					question.Type, question.Difficulty, strings.Join(question.Topics, ", "), strings.Join(question.Tags, ", "))

				if question.Type == "multipla_escolha" || question.Type == "verdadeiro_falso" { // Adaptar para "true_false" se existir
					if len(question.Options) > 0 {
						fmt.Println("   Opções:")
						for j, opt := range question.Options {
							prefix := fmt.Sprintf("     %c)", 'A'+j)
							if showAnswers {
								isCorrectOption := false
								for _, correctAns := range question.CorrectAnswers {
									// Assumindo que CorrectAnswers para múltipla escolha guarda o TEXTO da opção correta
									if strings.EqualFold(opt.Text, correctAns) {
										isCorrectOption = true
										break
									}
								}
								if isCorrectOption {
									prefix += " [*]" // Marcador para resposta correta
								} else {
									prefix += " [ ]"
								}
							}
							fmt.Printf("%s %s\n", prefix, opt.Text)
						}
					} else {
						fmt.Println("   AVISO: Questão de múltipla escolha sem opções definidas.")
					}
				}

				if showAnswers {
					if len(question.CorrectAnswers) > 0 {
						// Para dissertativas, ou para mostrar explicitamente a(s) resposta(s) correta(s)
						// mesmo para múltipla escolha (além do marcador [*])
						if question.Type == "dissertativa" || (question.Type == "multipla_escolha" && len(question.Options) == 0) {
							fmt.Printf("   Resposta(s) Correta(s): %s\n", strings.Join(question.CorrectAnswers, " | "))
						} else if question.Type == "multipla_escolha" && len(question.Options) > 0 {
							// Já mostrado com [*], mas pode adicionar um sumário se quiser
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
