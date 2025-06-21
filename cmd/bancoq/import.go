package bancoq

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings" // Added import for strings package
	"time"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	onConflictPolicy string
	isDryRun         bool
)

var bancoqImportCmd = &cobra.Command{
	Use:   "import <CAMINHO_DO_ARQUIVO_JSON>",
	Short: "Importa questões de um arquivo JSON para o banco de dados",
	Long: `Importa um conjunto de questões de um arquivo JSON, conforme o esquema de dados definido.
Este comando permite especificar como tratar conflitos de IDs (falhar, pular, ou atualizar) e
oferece um modo de simulação (dry-run) para verificar o processo sem efetuar alterações no banco.
O arquivo JSON deve conter um array de objetos de questão. Consulte 'docs/schemas/question_import_schema.md'.`,
	Args: cobra.ExactArgs(1),
	Run:  runImportQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqImportCmd)
	bancoqImportCmd.Flags().StringVar(&onConflictPolicy, "on-conflict", "fail", "Política para conflitos de ID: 'fail', 'skip', 'update'")
	bancoqImportCmd.Flags().BoolVar(&isDryRun, "dry-run", false, "Simula a importação sem gravar no banco")
}

// validateQuestionData realiza uma validação detalhada dos campos de uma questão.
func validateQuestionData(q *models.Question, index int) (bool, []string) {
	var validationErrors []string
	prefix := fmt.Sprintf("Questão %d (ID JSON: '%s', Disciplina: '%s')", index+1, q.ID, q.Subject)

	if q.Subject == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("%s: campo 'subject' é obrigatório.", prefix))
	}
	if q.Topic == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("%s: campo 'topic' é obrigatório.", prefix))
	}
	if q.QuestionText == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("%s: campo 'question_text' é obrigatório.", prefix))
	}
	if len(q.CorrectAnswers) == 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("%s: campo 'correct_answers' deve conter pelo menos uma resposta.", prefix))
	}

	if q.Difficulty == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("%s: campo 'difficulty' é obrigatório.", prefix))
	} else if !isValidDifficulty(q.Difficulty) { // Reutilizando de add.go (ou duplicar/mover para local comum)
		validationErrors = append(validationErrors, fmt.Sprintf("%s: 'difficulty' inválido ('%s'). Valores permitidos: easy, medium, hard.", prefix, q.Difficulty))
	}

	if q.QuestionType == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("%s: campo 'question_type' é obrigatório.", prefix))
	} else if !isValidQuestionType(q.QuestionType) { // Reutilizando de add.go
		validationErrors = append(validationErrors, fmt.Sprintf("%s: 'question_type' inválido ('%s'). Valores permitidos: multiple_choice, true_false, essay, short_answer.", prefix, q.QuestionType))
	}

	// Validações específicas para o tipo de questão
	if q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse {
		if len(q.AnswerOptions) == 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: para o tipo '%s', 'answer_options' não pode ser vazio.", prefix, q.QuestionType))
		}
	}
	if q.QuestionType == models.QuestionTypeMultipleChoice {
		for _, ans := range q.CorrectAnswers {
			found := false
			for _, opt := range q.AnswerOptions {
				if ans == opt {
					found = true
					break
				}
			}
			if !found {
				validationErrors = append(validationErrors, fmt.Sprintf("%s: resposta correta '%s' não encontrada em 'answer_options'.", prefix, ans))
			}
		}
	}

	return len(validationErrors) == 0, validationErrors
}

// Duplicadas de add.go para evitar dependência de ciclo ou para manter local.
// Idealmente, estas estariam em um pacote 'validators' ou similar.
func isValidDifficulty(difficulty string) bool {
	return difficulty == models.DifficultyEasy || difficulty == models.DifficultyMedium || difficulty == models.DifficultyHard
}
func isValidQuestionType(qType string) bool {
	return qType == models.QuestionTypeMultipleChoice || qType == models.QuestionTypeTrueFalse || qType == models.QuestionTypeEssay || qType == models.QuestionTypeShortAnswer
}


func runImportQuestions(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	filePath := args[0]
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao obter caminho absoluto de '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	validPolicies := map[string]bool{"fail": true, "skip": true, "update": true}
	if !validPolicies[onConflictPolicy] {
		fmt.Fprintf(os.Stderr, "Política --on-conflict inválida: '%s'. Use 'fail', 'skip', ou 'update'.\n", onConflictPolicy)
		os.Exit(1)
	}

	fmt.Printf("Importando questões de: %s\n", absFilePath)
	if isDryRun {
		fmt.Println("ATENÇÃO: Modo de SIMULAÇÃO (Dry Run). Nenhuma alteração será feita no banco.")
	}
	fmt.Printf("Política de conflito de ID: %s\n\n", onConflictPolicy)

	jsonFile, err := os.Open(absFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir arquivo JSON '%s': %v\n", absFilePath, err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler arquivo JSON '%s': %v\n", absFilePath, err)
		os.Exit(1)
	}

	var questionsToImport []models.Question
	if err = json.Unmarshal(byteValue, &questionsToImport); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao decodificar JSON de '%s': %v\nVerifique se o JSON é um array de objetos de questão.\n", absFilePath, err)
		os.Exit(1)
	}

	successfulCreations, failedImports, skippedImports, successfulUpdates := 0, 0, 0, 0
	var importErrorDetails []string

	if len(questionsToImport) == 0 {
		fmt.Println("Nenhuma questão encontrada no arquivo JSON.")
		return
	}

	for i, q := range questionsToImport {
		fmt.Printf("Processando questão %d/%d: Subj: '%s', Topic: '%s'...\n", i+1, len(questionsToImport), q.Subject, q.Topic)

		valid, valErrors := validateQuestionData(&q, i)
		if !valid {
			importErrorDetails = append(importErrorDetails, valErrors...)
			failedImports++
			fmt.Printf("  Erro de validação: %s\n", strings.Join(valErrors, "; "))
			continue
		}

		isNewQuestion := true
		originalJSONID := q.ID // ID como no arquivo JSON, pode ser ""

		if q.ID == "" {
			q.ID = uuid.NewString() // Gerar novo ID se não fornecido
			fmt.Printf("  ID não fornecido; novo ID gerado: %s\n", q.ID)
		} else {
			// Verificar se ID do JSON já existe no banco
			existingQuestion, dbErr := db.GetQuestion(q.ID)
			if dbErr == nil { // Questão com este ID já existe
				isNewQuestion = false
				fmt.Printf("  Conflito: Questão com ID '%s' (Subj: '%s') já existe no banco.\n", q.ID, existingQuestion.Subject)
				switch onConflictPolicy {
				case "fail":
					errStr := fmt.Sprintf("Falha devido à política 'fail' para ID '%s'.", q.ID)
					importErrorDetails = append(importErrorDetails, errStr)
					failedImports++
					fmt.Printf("    %s\n", errStr)
					continue
				case "skip":
					skippedImports++
					fmt.Printf("    Ignorada (política 'skip').\n")
					continue
				case "update":
					fmt.Printf("    Será atualizada (política 'update').\n")
					// Preservar CreatedAt original do banco se não especificado no JSON
					if q.CreatedAt.IsZero() && !existingQuestion.CreatedAt.IsZero() {
						q.CreatedAt = existingQuestion.CreatedAt
					} else if q.CreatedAt.IsZero() { // Se ambos CreatedAt (JSON e banco) são zero, usar Now()
						q.CreatedAt = time.Now()
					}
					// LastUsedAt e Author do JSON sobrescrevem os do banco.
					if !isDryRun {
						if errUpdate := db.UpdateQuestion(q); errUpdate != nil {
							errStr := fmt.Sprintf("Erro ao ATUALIZAR ID '%s': %v", q.ID, errUpdate)
							importErrorDetails = append(importErrorDetails, errStr)
							failedImports++
							fmt.Printf("      Erro na atualização: %v\n", errUpdate)
							continue
						}
					}
					successfulUpdates++
					fmt.Printf("    Questão ID '%s' %s.\n", q.ID, tern(isDryRun, "seria ATUALIZADA", "ATUALIZADA"))
					continue // Próxima questão
				}
			} else if !errors.Is(dbErr, sql.ErrNoRows) { // Erro inesperado ao verificar
				errStr := fmt.Sprintf("Erro ao verificar ID '%s' no banco: %v", q.ID, dbErr)
				importErrorDetails = append(importErrorDetails, errStr)
				failedImports++
				fmt.Printf("    %s\n", errStr)
				continue
			}
			// Se sql.ErrNoRows, ID do JSON não existe, então é nova para o banco.
		}

		if q.CreatedAt.IsZero() { // Se é nova e CreatedAt não veio do JSON
			q.CreatedAt = time.Now()
		}

		if isNewQuestion { // Somente criar se for realmente nova para o banco
			if !isDryRun {
				if _, errCreate := db.CreateQuestion(q); errCreate != nil {
					errStr := fmt.Sprintf("Erro ao CRIAR questão (ID JSON: '%s', ID Gerado/Usado: '%s'): %v", originalJSONID, q.ID, errCreate)
					importErrorDetails = append(importErrorDetails, errStr)
					failedImports++
					fmt.Printf("    Erro na criação: %v\n", errCreate)
					continue
				}
			}
			successfulCreations++
			fmt.Printf("  Questão (ID JSON: '%s') %s com ID %s.\n", originalJSONID, tern(isDryRun, "seria CRIADA", "CRIADA"), q.ID)
		}
	}

	fmt.Println("\n--- Relatório da Importação ---")
	if isDryRun {
		fmt.Println("MODO DE SIMULAÇÃO (Dry Run). Nenhuma alteração foi persistida.")
	}
	fmt.Printf("Total de questões no arquivo: %d\n", len(questionsToImport))
	fmt.Printf("Criadas com sucesso: %d\n", successfulCreations)
	fmt.Printf("Atualizadas com sucesso (política 'update'): %d\n", successfulUpdates)
	fmt.Printf("Ignoradas (política 'skip'): %d\n", skippedImports)
	fmt.Printf("Falhas (erro de validação, erro no DB, ou política 'fail'): %d\n", failedImports)

	if len(importErrorDetails) > 0 {
		fmt.Println("\nDetalhes dos erros/falhas:")
		for _, detail := range importErrorDetails {
			fmt.Printf("  - %s\n", detail)
		}
	}
	fmt.Println("-------------------------------")
}

func tern(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}
