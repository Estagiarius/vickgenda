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
	Long: `Importa um array de questões de um arquivo JSON, conforme o esquema definido.
Permite tratamento de conflitos e simulação (dry-run).
O arquivo JSON deve conter um array de objetos de questão. Consulte a documentação para o esquema exato.`,
	Args: cobra.ExactArgs(1),
	Run:  runImportQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqImportCmd)
	bancoqImportCmd.Flags().StringVar(&onConflictPolicy, "on-conflict", "fail", "Política de conflito de ID: fail, skip, update")
	bancoqImportCmd.Flags().BoolVar(&isDryRun, "dry-run", false, "Simula a importação sem gravar no banco de dados")
}

func runImportQuestions(cmd *cobra.Command, args []string) {
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	filePath := args[0]
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao obter caminho absoluto do arquivo '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	// Validate onConflictPolicy
	validPolicies := map[string]bool{"fail": true, "skip": true, "update": true}
	if !validPolicies[onConflictPolicy] {
		fmt.Fprintf(os.Stderr, "Política --on-conflict inválida: '%s'. Use 'fail', 'skip', ou 'update'.\n", onConflictPolicy)
		os.Exit(1)
	}

	fmt.Printf("Importando questões de: %s\n", absFilePath)
	if isDryRun {
		fmt.Println("ATENÇÃO: Executando em modo Dry Run. Nenhuma alteração será feita no banco de dados.")
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
		fmt.Fprintf(os.Stderr, "Erro ao decodificar JSON do arquivo '%s': %v\n", absFilePath, err)
		fmt.Fprintln(os.Stderr, "Verifique se o JSON está formatado corretamente como um array de questões.")
		os.Exit(1)
	}

	successfulImports := 0
	failedImports := 0
	skippedImports := 0
	updatedImports := 0
	var errorDetails []string

	if len(questionsToImport) == 0 {
		fmt.Println("Nenhuma questão encontrada no arquivo JSON.")
		return
	}

	for i, q := range questionsToImport {
		fmt.Printf("Processando questão %d de %d: Assunto '%s'...\n", i+1, len(questionsToImport), q.Subject)

		// Basic Validation
		if q.Subject == "" || q.Topic == "" || q.Difficulty == "" || q.QuestionText == "" || len(q.CorrectAnswers) == 0 || q.QuestionType == "" {
			errorMsg := fmt.Sprintf("Questão %d (Assunto: '%s') inválida: campos obrigatórios faltando (Subject, Topic, Difficulty, QuestionText, CorrectAnswers, QuestionType).", i+1, q.Subject)
			errorDetails = append(errorDetails, errorMsg)
			failedImports++
			fmt.Printf("  Erro: %s\n", errorMsg)
			continue
		}
		// TODO: Add validation for Difficulty and QuestionType values against known constants

		isNewQuestion := true // Flag to determine if we should create or update
		originalQuestionID := q.ID // Preserve original ID from JSON for messages

		if q.ID == "" {
			q.ID = uuid.NewString()
			fmt.Printf("  ID não fornecido, novo ID gerado: %s\n", q.ID)
		} else {
			originalQuestion, errDbGet := db.GetQuestion(q.ID)
			if errDbGet == nil { // Question with this ID exists
				fmt.Printf("  Questão com ID '%s' já existe no banco de dados.\n", q.ID)
				isNewQuestion = false // It's not new, it's a potential conflict
				switch onConflictPolicy {
				case "fail":
					errorMsg := fmt.Sprintf("Questão ID '%s' (Assunto: '%s') já existe. Falha devido à política 'fail'.", q.ID, q.Subject)
					errorDetails = append(errorDetails, errorMsg)
					failedImports++
					fmt.Printf("    %s\n", errorMsg)
					continue
				case "skip":
					skippedImports++
					fmt.Printf("    Questão ID '%s' (Assunto: '%s') pulada devido à política 'skip'.\n", q.ID, q.Subject)
					continue
				case "update":
					fmt.Printf("    Questão ID '%s' (Assunto: '%s') será atualizada (política 'update').\n", q.ID, q.Subject)
					// Preserve CreatedAt from original if not specified or zero in the import data
					if q.CreatedAt.IsZero() && !originalQuestion.CreatedAt.IsZero() {
						q.CreatedAt = originalQuestion.CreatedAt
					}
					// LastUsedAt from import file will overwrite, or be zero if not present
					// Author from import file will overwrite

					if !isDryRun {
						if errUpdate := db.UpdateQuestion(q); errUpdate != nil {
							errorMsg := fmt.Sprintf("Erro ao atualizar questão ID '%s' (Assunto: '%s'): %v", q.ID, q.Subject, errUpdate)
							errorDetails = append(errorDetails, errorMsg)
							failedImports++
							fmt.Printf("      Erro na atualização: %v\n", errUpdate)
							continue
						}
					}
					updatedImports++
					fmt.Printf("    Questão ID '%s' (Assunto: '%s') %s.\n", q.ID, q.Subject, tern(isDryRun, "seria atualizada", "foi atualizada"))
					continue // Move to next question after handling update
				}
			} else if !errors.Is(errDbGet, sql.ErrNoRows) && !strings.Contains(errDbGet.Error(), "not found") {
				// An unexpected error occurred while checking if the question exists
				errorMsg := fmt.Sprintf("Erro ao verificar existência da questão ID '%s' (Assunto: '%s'): %v", q.ID, q.Subject, errDbGet)
				errorDetails = append(errorDetails, errorMsg)
				failedImports++
				fmt.Printf("    %s\n", errorMsg)
				continue
			}
			// If ErrNoRows or "not found", it means ID from JSON doesn't exist, so proceed to create.
			// isNewQuestion remains true (or effectively, as it's not a conflict case that stops creation)
		}

		// Set CreatedAt if not provided in JSON (and it's a new question or update didn't set it)
		if q.CreatedAt.IsZero() {
			q.CreatedAt = time.Now()
		}

		// Perform Import (Create new question if isNewQuestion is true)
		if isNewQuestion { // Only create if it's genuinely new or ID from JSON wasn't found
			if !isDryRun {
				_, errDbCreate := db.CreateQuestion(q)
				if errDbCreate != nil {
					errorMsg := fmt.Sprintf("Erro ao criar questão ID '%s' (Assunto: '%s', ID Original JSON: '%s'): %v", q.ID, q.Subject, originalQuestionID, errDbCreate)
					errorDetails = append(errorDetails, errorMsg)
					failedImports++
					fmt.Printf("    Erro na criação: %v\n", errDbCreate)
					continue
				}
				fmt.Printf("  Questão (Assunto: '%s') importada com ID %s.\n", q.Subject, q.ID)
			} else {
				fmt.Printf("  [Dry Run] Questão (Assunto: '%s') seria importada com ID %s (ID Original JSON: '%s').\n", q.Subject, q.ID, originalQuestionID)
			}
			successfulImports++
		}
	}

	fmt.Println("\n--- Relatório da Importação ---")
	if isDryRun {
		fmt.Println("Execução em modo Dry Run. Nenhuma alteração foi feita no banco de dados.")
	}
	fmt.Printf("Total de questões no arquivo: %d\n", len(questionsToImport))
	fmt.Printf("Importadas com sucesso (novas ou que seriam novas): %d\n", successfulImports)
	if onConflictPolicy == "update" {
		fmt.Printf("Atualizadas (ou que seriam atualizadas): %d\n", updatedImports)
	}
	if onConflictPolicy == "skip" {
		fmt.Printf("Puladas (conflito de ID): %d\n", skippedImports)
	}
	fmt.Printf("Falhas na importação: %d\n", failedImports)

	if len(errorDetails) > 0 {
		fmt.Println("\nDetalhes dos erros:")
		for _, detail := range errorDetails {
			fmt.Printf("  - %s\n", detail)
		}
	}
	fmt.Println("-------------------------------")
}

// Simple ternary helper for logging
func tern(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}
