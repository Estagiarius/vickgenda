package prova

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"vickgenda-cli/internal/models" // Assuming models.Question and models.Test are defined here
)

// Sample questions for simulation
var sampleQuestions = []models.Question{
	{ID: "q1", Subject: "Matemática", QuestionText: "Quanto é 2+2?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, Topic: "aritmética", Tags: []string{"básica"}, AnswerOptions: []string{"3", "4", "5"}, CorrectAnswers: []string{"4"}},
	{ID: "q2", Subject: "Matemática", QuestionText: "Quanto é 5*8?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, Topic: "aritmética", Tags: []string{"básica"}, AnswerOptions: []string{"30", "40", "35"}, CorrectAnswers: []string{"40"}},
	{ID: "q3", Subject: "História", QuestionText: "Quem descobriu o Brasil?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyMedium, Topic: "descobrimentos", Tags: []string{"Brasil"}, CorrectAnswers: []string{"Pedro Álvares Cabral"}},
	{ID: "q4", Subject: "Matemática", QuestionText: "Qual a derivada de x^2?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyHard, Topic: "cálculo", Tags: []string{"avançada"}, CorrectAnswers: []string{"2x"}},
	{ID: "q5", Subject: "Geografia", QuestionText: "Qual a capital da França?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyEasy, Topic: "europa", Tags: []string{"capitais"}, AnswerOptions: []string{"Londres", "Paris", "Madri"}, CorrectAnswers: []string{"Paris"}},
	{ID: "q6", Subject: "Matemática", QuestionText: "Resolva a equação: x + 5 = 10", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyEasy, Topic: "algebra", Tags: []string{"equação"}, CorrectAnswers: []string{"x = 5"}},
	{ID: "q7", Subject: "História", QuestionText: "Em que ano começou a Segunda Guerra Mundial?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyMedium, Topic: "guerras mundiais", Tags: []string{"século XX"}, AnswerOptions: []string{"1939", "1941", "1945"}, CorrectAnswers: []string{"1939"}},
	{ID: "q8", Subject: "Matemática", QuestionText: "Qual o valor de Pi (aproximado)?", QuestionType: models.QuestionTypeMultipleChoice, Difficulty: models.DifficultyMedium, Topic: "geometria", Tags: []string{"constantes"}, AnswerOptions: []string{"3.14", "3.12", "3.16"}, CorrectAnswers: []string{"3.14"}},
	{ID: "q9", Subject: "Matemática", QuestionText: "O que é um número primo?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyMedium, Topic: "teoria dos números", Tags: []string{"definição"}, CorrectAnswers: []string{"Um número natural maior que 1 que não possui outros divisores além de 1 e ele mesmo."}},
	{ID: "q10", Subject: "Matemática", QuestionText: "Qual a área de um círculo de raio r?", QuestionType: models.QuestionTypeEssay, Difficulty: models.DifficultyHard, Topic: "geometria", Tags: []string{"fórmula"}, CorrectAnswers: []string{"πr²"}},
}

// generateCmd representa o comando para gerar uma nova prova.
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Gera uma nova prova",
	Long:  `Cria uma nova prova com base em questões existentes no banco de dados, permitindo especificar diversos critérios de seleção e formatação.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executando o comando 'prova generate'...")

		// Recuperar valores das flags
		title, _ := cmd.Flags().GetString("title")
		subjectFilter, _ := cmd.Flags().GetString("subject") // Renomeado para evitar conflito com prova.Subject
		topicsFilter, _ := cmd.Flags().GetStringSlice("topic")
		difficultiesFilter, _ := cmd.Flags().GetStringSlice("difficulty")
		typesFilter, _ := cmd.Flags().GetStringSlice("type")
		tagsFilter, _ := cmd.Flags().GetStringSlice("tag")
		numQuestionsTotal, _ := cmd.Flags().GetInt("num-questions")
		numEasy, _ := cmd.Flags().GetInt("num-easy")
		numMedium, _ := cmd.Flags().GetInt("num-medium")
		numHard, _ := cmd.Flags().GetInt("num-hard")
		// allowDuplicates, _ := cmd.Flags().GetBool("allow-duplicates") // Implementar se necessário
		randomizeOrder, _ := cmd.Flags().GetBool("randomize-order")
		outputFile, _ := cmd.Flags().GetString("output-file")
		// outputFormat, _ := cmd.Flags().GetString("output-format") // Usaremos TXT simples por enquanto
		instructions, _ := cmd.Flags().GetString("instructions")

		fmt.Println("\n--- Critérios Iniciais para Geração da Prova ---")
		fmt.Printf("Título: %s, Disciplina: %s\n", title, subjectFilter)
		// Adicionar mais prints dos filtros se necessário para debug

		// 1. Filtrar questões (simulação)
		var filteredQuestions []models.Question
		for _, q := range sampleQuestions {
			match := true
			if subjectFilter != "" && !strings.EqualFold(q.Subject, subjectFilter) {
				match = false
			}
			// Changed q.Topics to q.Topic (string) and adapted containsAny to work with string topic
			if len(topicsFilter) > 0 && !contains(topicsFilter, q.Topic) { // Assuming topicsFilter expects single topic match now
				match = false
			}
			if len(difficultiesFilter) > 0 && !contains(difficultiesFilter, q.Difficulty) {
				match = false
			}
			if len(typesFilter) > 0 && !contains(typesFilter, q.QuestionType) { // Changed q.Type to q.QuestionType
				match = false
			}
			if len(tagsFilter) > 0 && !containsAny(q.Tags, tagsFilter) {
				match = false
			}
			if match {
				filteredQuestions = append(filteredQuestions, q)
			}
		}

		if len(filteredQuestions) == 0 {
			fmt.Println("Nenhuma questão foi encontrada com os critérios especificados.")
			return
		}
		fmt.Printf("Total de questões filtradas inicialmente: %d\n", len(filteredQuestions))

		// 2. Selecionar questões
		var selectedQuestions []models.Question
		var randomizationSeed int64 // Para guardar a semente se randomização for usada

		if numQuestionsTotal > 0 {
			// Selecionar aleatoriamente do total filtrado
			if randomizeOrder {
				rand.Seed(time.Now().UnixNano()) // Inicializa a semente
				randomizationSeed = rand.Int63()  // Guarda a semente
				rand.Shuffle(len(filteredQuestions), func(i, j int) {
					filteredQuestions[i], filteredQuestions[j] = filteredQuestions[j], filteredQuestions[i]
				})
			}
			for i := 0; i < numQuestionsTotal && i < len(filteredQuestions); i++ {
				selectedQuestions = append(selectedQuestions, filteredQuestions[i])
			}
		} else if numEasy > 0 || numMedium > 0 || numHard > 0 {
			// Selecionar por dificuldade
			questionsByDifficulty := map[string][]models.Question{
				"fácil":   {},
				"médio":   {},
				"difícil": {},
			}
			for _, q := range filteredQuestions {
				questionsByDifficulty[q.Difficulty] = append(questionsByDifficulty[q.Difficulty], q)
			}

			if randomizeOrder { // Randomizar dentro de cada categoria de dificuldade antes de selecionar
				rand.Seed(time.Now().UnixNano())
				randomizationSeed = rand.Int63()
				for k := range questionsByDifficulty {
					rand.Shuffle(len(questionsByDifficulty[k]), func(i, j int) {
						questionsByDifficulty[k][i], questionsByDifficulty[k][j] = questionsByDifficulty[k][j], questionsByDifficulty[k][i]
					})
				}
			}

			selectFromCategory := func(category string, count int) {
				available := questionsByDifficulty[category]
				for i := 0; i < count && i < len(available); i++ {
					selectedQuestions = append(selectedQuestions, available[i])
				}
			}
			selectFromCategory("fácil", numEasy)
			selectFromCategory("médio", numMedium)
			selectFromCategory("difícil", numHard)
			// Nota: Não estamos tratando --allow-duplicates explicitamente, a seleção acima já evita duplicatas.
			// Se o número total de questões selecionadas por dificuldade for importante,
			// pode ser necessário ajustar ou limitar ao numQuestionsTotal se este também for fornecido.
		} else {
			// Se nenhum critério numérico for dado, selecionar todas as filtradas ou um padrão (ex: 10)
			if randomizeOrder {
				rand.Seed(time.Now().UnixNano())
				randomizationSeed = rand.Int63()
				rand.Shuffle(len(filteredQuestions), func(i, j int) {
					filteredQuestions[i], filteredQuestions[j] = filteredQuestions[j], filteredQuestions[i]
				})
			}
			maxDefaultQuestions := 10
			if len(filteredQuestions) < maxDefaultQuestions {
				maxDefaultQuestions = len(filteredQuestions)
			}
			selectedQuestions = filteredQuestions[:maxDefaultQuestions]
		}

		if len(selectedQuestions) == 0 {
			fmt.Println("Não foi possível selecionar questões com os números especificados a partir das questões filtradas.")
			return
		}

		// Se randomizeOrder foi global e não por dificuldade, e não foi feito antes
		if randomizeOrder && numQuestionsTotal > 0 { // Já feito acima para numQuestionsTotal
			// Se a seleção foi por dificuldade, e queremos randomizar o conjunto final
			// rand.Seed(time.Now().UnixNano()) // Semente já inicializada se randomizeOrder=true
			// randomizationSeed = rand.Int63() // Semente já guardada
			rand.Shuffle(len(selectedQuestions), func(i, j int) {
				selectedQuestions[i], selectedQuestions[j] = selectedQuestions[j], selectedQuestions[i]
			})
		}


		// 3. Criar objeto models.Test
		questionIDs := make([]string, len(selectedQuestions))
		for i, q := range selectedQuestions {
			questionIDs[i] = q.ID
		}

		prova := models.Test{
			ID:                uuid.NewString(),
			Title:             title,
			Subject:           subjectFilter, // Usar o subject do filtro como o principal da prova
			CreatedAt:         time.Now(),
			Instructions:      instructions,
			QuestionIDs:       questionIDs,
			LayoutOptions:     make(map[string]string), // Pode ser preenchido com defaults ou futuras flags
			RandomizationSeed: 0,                       // Definir se a randomização ocorreu
		}
		if randomizeOrder && randomizationSeed != 0 {
			prova.RandomizationSeed = randomizationSeed
		}


		fmt.Printf("\n--- Prova Gerada (Objeto models.Test) ---\n")
		fmt.Printf("%+v\n", prova)
		fmt.Println("------------------------------------")

		// 4. Simular Saída
		if outputFile != "" {
			fmt.Printf("\nSimulando salvamento da prova em: %s\n", outputFile)
			// Aqui ocorreria a escrita no arquivo, usando o outputFormat se necessário.
			// Por agora, apenas a mensagem.
		} else {
			fmt.Printf("\n--- Visualização da Prova (Formato Texto Simples) ---\n")
			fmt.Printf("Título: %s\n", prova.Title)
			fmt.Printf("Disciplina: %s\n", prova.Subject)
			if prova.Instructions != "" {
				fmt.Printf("Instruções: %s\n", prova.Instructions)
			}
			if prova.RandomizationSeed != 0 {
				fmt.Printf("(Questões/alternativas randomizadas com semente: %d)\n", prova.RandomizationSeed)
			}
			fmt.Println("---")

			for i, qID := range prova.QuestionIDs {
				var questionText, questionType string
				var options []string
				// Encontrar a questão original no sampleQuestions para obter detalhes
				// Em um sistema real, buscaria do DB
				originalQuestion := findQuestionByID(qID, sampleQuestions)
				if originalQuestion != nil {
					questionText = originalQuestion.QuestionText // Changed .Text to .QuestionText
					questionType = originalQuestion.QuestionType // Changed .Type to .QuestionType
					if originalQuestion.QuestionType == models.QuestionTypeMultipleChoice { // Changed .Type to .QuestionType
						options = originalQuestion.AnswerOptions // Changed .Options to .AnswerOptions (it's already []string)
					}
				} else {
					questionText = "Texto da questão não encontrado (ID: " + qID + ")"
				}

				fmt.Printf("\nQuestão %d (ID: %s, Tipo: %s): %s\n", i+1, qID, questionType, questionText)
				if len(options) > 0 {
					fmt.Println("Opções:")
					for j, optText := range options {
						fmt.Printf("  %c) %s\n", 'A'+j, optText)
					}
				}
			}
			fmt.Println("\n---------------------------------------------")
		}
		fmt.Println("\nComando 'prova generate' concluído com lógica de simulação.")
	},
}

// Helper para verificar se um slice de strings contém um valor específico (case-insensitive)
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, val) {
			return true
		}
	}
	return false
}

// Helper para verificar se um slice de strings contém qualquer um dos valores de outro slice (case-insensitive)
func containsAny(slice []string, vals []string) bool {
	for _, item := range slice {
		for _, val := range vals {
			if strings.EqualFold(item, val) {
				return true
			}
		}
	}
	return false
}

// Helper para encontrar uma questão por ID em uma lista de questões (simulação)
func findQuestionByID(id string, questions []models.Question) *models.Question {
	for _, q := range questions {
		if q.ID == id {
			return &q
		}
	}
	return nil
}


func init() {
	ProvaCmd.AddCommand(generateCmd)

	// Flags para o comando generate (baseado em docs/specifications/prova_command_spec.md):
	generateCmd.Flags().StringP("title", "t", "", "Título da prova (obrigatório)")
	generateCmd.MarkFlagRequired("title")

	generateCmd.Flags().StringP("subject", "s", "", "Disciplina principal da prova (obrigatório)")
	generateCmd.MarkFlagRequired("subject")

	generateCmd.Flags().StringSlice("topic", []string{}, "Tópicos/assuntos específicos a serem incluídos na prova (opcional)")
	generateCmd.Flags().StringSlice("difficulty", []string{}, "Níveis de dificuldade das questões (ex: facil, medio, dificil) (opcional)")
	generateCmd.Flags().StringSlice("type", []string{}, "Tipos de questões (ex: multipla_escolha, dissertativa) (opcional)")
	generateCmd.Flags().StringSlice("tag", []string{}, "Tags para filtrar questões (opcional)")

	generateCmd.Flags().Int("num-questions", 0, "Número total de questões a serem selecionadas aleatoriamente (opcional)")
	generateCmd.Flags().Int("num-easy", 0, "Número específico de questões fáceis (opcional, usado se --num-questions não for especificado)")
	generateCmd.Flags().Int("num-medium", 0, "Número específico de questões médias (opcional, usado se --num-questions não for especificado)")
	generateCmd.Flags().Int("num-hard", 0, "Número específico de questões difíceis (opcional, usado se --num-questions não for especificado)")

	generateCmd.Flags().Bool("allow-duplicates", false, "Permitir que a mesma questão apareça mais de uma vez (opcional, padrão: false)")
	generateCmd.Flags().Bool("randomize-order", false, "Randomizar a ordem das questões na prova gerada (opcional, padrão: false)")

	generateCmd.Flags().StringP("output-file", "o", "", "Caminho do arquivo onde a prova gerada será salva (opcional)")
	generateCmd.Flags().String("output-format", "txt", "Formato do arquivo de saída (ex: txt, json, pdf) (opcional, padrão: txt)")
	generateCmd.Flags().StringP("instructions", "i", "", "Instruções gerais para a prova (opcional)")

	// Placeholder flags from original template, kept for reference, might be used differently or removed
	// generateCmd.Flags().StringArrayP("question-ids", "q", []string{}, "Lista de IDs de questões a serem incluídas") // This might be an alternative way to specify questions
	// generateCmd.Flags().StringToStringP("layout-options", "l", map[string]string{}, "Opções de formatação (ex: columns=2,header='Nome da Escola')")
	// generateCmd.Flags().Int64("randomization-seed", 0, "Semente para randomização da ordem das questões/alternativas")
}
