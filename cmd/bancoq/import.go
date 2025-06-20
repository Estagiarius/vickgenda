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
O arquivo JSON deve conter um array de objetos de questão. Consulte a documentação para o esquema exato.`,
	Args: cobra.ExactArgs(1),
	Run:  runImportQuestions,
}

func init() {
	BancoqCmd.AddCommand(bancoqImportCmd)
	bancoqImportCmd.Flags().StringVar(&onConflictPolicy, "on-conflict", "fail", "Política de tratamento para conflitos de ID: 'fail' (falhar), 'skip' (pular), 'update' (atualizar)")
	bancoqImportCmd.Flags().BoolVar(&isDryRun, "dry-run", false, "Simula a importação sem gravar alterações no banco de dados")
}

func runImportQuestions(cmd *cobra.Command, args []string) {
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	filePath := args[0]
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao obter o caminho absoluto do arquivo '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	// Validate onConflictPolicy
	validPolicies := map[string]bool{"fail": true, "skip": true, "update": true}
	if !validPolicies[onConflictPolicy] {
		fmt.Fprintf(os.Stderr, "Política --on-conflict inválida: '%s'. Valores permitidos: 'fail', 'skip', 'update'.\n", onConflictPolicy)
		os.Exit(1)
	}

	fmt.Printf("Importando questões de: %s\n", absFilePath)
	if isDryRun {
		fmt.Println("ATENÇÃO: Executando em modo de simulação (Dry Run). Nenhuma alteração será persistida no banco de dados.")
	}
	fmt.Printf("Política de conflito de ID: %s\n\n", onConflictPolicy)

	jsonFile, err := os.Open(absFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir o arquivo JSON '%s': %v\n", absFilePath, err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o arquivo JSON '%s': %v\n", absFilePath, err)
		os.Exit(1)
	}

	var questionsToImport []models.Question
	if err = json.Unmarshal(byteValue, &questionsToImport); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao decodificar o JSON do arquivo '%s': %v\n", absFilePath, err)
		fmt.Fprintln(os.Stderr, "Verifique se o JSON está formatado corretamente como um array de objetos de questão.")
		os.Exit(1)
	}

	successfulImports := 0
	failedImports := 0
	skippedImports := 0
	updatedImports := 0
	var errorDetails []string

	if len(questionsToImport) == 0 {
		fmt.Println("Nenhuma questão foi encontrada no arquivo JSON fornecido.")
		return
	}

	for i, q := range questionsToImport {
		fmt.Printf("Processando questão %d de %d: Disciplina '%s'...\n", i+1, len(questionsToImport), q.Subject)

		// Basic Validation
		if q.Subject == "" || q.Topic == "" || q.Difficulty == "" || q.QuestionText == "" || len(q.CorrectAnswers) == 0 || q.QuestionType == "" {
			errorMsg := fmt.Sprintf("Questão %d (Disciplina: '%s') inválida: campos obrigatórios faltando (Subject, Topic, Difficulty, QuestionText, CorrectAnswers, QuestionType).", i+1, q.Subject)
			errorDetails = append(errorDetails, errorMsg)
			failedImports++
			fmt.Printf("  Erro: %s\n", errorMsg)
			continue
		}
		// TODO: Adicionar validação para os valores de Difficulty e QuestionType contra constantes conhecidas

		isNewQuestion := true // Flag para determinar se devemos criar ou atualizar
		originalQuestionID := q.ID // Preservar ID original do JSON para mensagens

		if q.ID == "" {
			q.ID = uuid.NewString()
			fmt.Printf("  ID não fornecido na questão; novo ID gerado: %s\n", q.ID)
		} else {
			originalQuestion, errDbGet := db.GetQuestion(q.ID)
			if errDbGet == nil { // Questão com este ID existe
				fmt.Printf("  A questão com ID '%s' já existe no banco de dados.\n", q.ID)
				isNewQuestion = false // Não é nova, é um conflito potencial
				switch onConflictPolicy {
				case "fail":
					errorMsg := fmt.Sprintf("Questão ID '%s' (Disciplina: '%s') já existe. Falha devido à política 'fail'.", q.ID, q.Subject)
					errorDetails = append(errorDetails, errorMsg)
					failedImports++
					fmt.Printf("    %s\n", errorMsg)
					continue
				case "skip":
					skippedImports++
					fmt.Printf("    Questão ID '%s' (Disciplina: '%s') ignorada devido à política 'skip'.\n", q.ID, q.Subject)
					continue
				case "update":
					fmt.Printf("    Questão ID '%s' (Disciplina: '%s') será atualizada (política 'update').\n", q.ID, q.Subject)
					// Preservar CreatedAt original se não especificado ou zero nos dados de importação
					if q.CreatedAt.IsZero() && !originalQuestion.CreatedAt.IsZero() {
						q.CreatedAt = originalQuestion.CreatedAt
					}
					// LastUsedAt do arquivo de importação sobrescreverá, ou será zero se não presente
					// Author do arquivo de importação sobrescreverá

					if !isDryRun {
						if errUpdate := db.UpdateQuestion(q); errUpdate != nil {
							errorMsg := fmt.Sprintf("Erro ao atualizar a questão ID '%s' (Disciplina: '%s'): %v", q.ID, q.Subject, errUpdate)
							errorDetails = append(errorDetails, errorMsg)
							failedImports++
							fmt.Printf("      Erro na atualização: %v\n", errUpdate)
							continue
						}
					}
					updatedImports++
					fmt.Printf("    Questão ID '%s' (Disciplina: '%s') %s.\n", q.ID, q.Subject, tern(isDryRun, "seria atualizada", "foi atualizada"))
					continue // Próxima questão após tratar atualização
				}
			} else if !errors.Is(errDbGet, sql.ErrNoRows) && !strings.Contains(errDbGet.Error(), "not found") {
				// Erro inesperado ao verificar se a questão existe
				errorMsg := fmt.Sprintf("Erro ao verificar a existência da questão ID '%s' (Disciplina: '%s'): %v", q.ID, q.Subject, errDbGet)
				errorDetails = append(errorDetails, errorMsg)
				failedImports++
				fmt.Printf("    %s\n", errorMsg)
				continue
			}
			// Se ErrNoRows ou "not found", significa que o ID do JSON não existe, então prosseguir para criar.
		}

		// Definir CreatedAt se não fornecido no JSON (e é uma nova questão ou atualização não definiu)
		if q.CreatedAt.IsZero() {
			q.CreatedAt = time.Now()
		}

		// Realizar Importação (Criar nova questão se isNewQuestion for true)
		if isNewQuestion { // Criar apenas se for genuinamente nova ou ID do JSON não foi encontrado
			if !isDryRun {
				_, errDbCreate := db.CreateQuestion(q)
				if errDbCreate != nil {
					errorMsg := fmt.Sprintf("Erro ao criar questão ID '%s' (Disciplina: '%s', ID Original JSON: '%s'): %v", q.ID, q.Subject, originalQuestionID, errDbCreate)
					errorDetails = append(errorDetails, errorMsg)
					failedImports++
					fmt.Printf("    Erro na criação: %v\n", errDbCreate)
					continue
				}
				fmt.Printf("  Questão (Disciplina: '%s') importada com ID %s.\n", q.Subject, q.ID)
			} else {
				fmt.Printf("  [Dry Run] Questão (Disciplina: '%s') seria importada com ID %s (ID Original JSON: '%s').\n", q.Subject, q.ID, originalQuestionID)
			}
			successfulImports++
		}
	}

	fmt.Println("\n--- Relatório da Importação ---")
	if isDryRun {
		fmt.Println("Execução em modo de simulação (Dry Run). Nenhuma alteração foi persistida no banco de dados.")
	}
	fmt.Printf("Total de questões no arquivo: %d\n", len(questionsToImport))
	fmt.Printf("Importadas com sucesso (novas ou que seriam novas): %d\n", successfulImports)
	if onConflictPolicy == "update" {
		fmt.Printf("Atualizadas (ou que seriam atualizadas): %d\n", updatedImports)
	}
	if onConflictPolicy == "skip" {
		fmt.Printf("Ignoradas (conflito de ID com política 'skip'): %d\n", skippedImports)
	}
	fmt.Printf("Falhas na importação (erros ou conflito com política 'fail'): %d\n", failedImports)

	if len(errorDetails) > 0 {
		fmt.Println("\nDetalhes dos erros/falhas:")
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
