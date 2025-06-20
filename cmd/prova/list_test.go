package prova

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"github.com/spf13/cobra" // Moved from bottom
	"github.com/spf13/pflag"  // Moved from bottom
	"vickgenda-cli/internal/models" // Moved from bottom (and uncommented)
)

// executeProvaListCommand executes the 'prova list' command with given arguments
// and returns the stdout/stderr output and any error.
func executeProvaListCommand(args ...string) (string, error) {
	cmdToTest := ProvaCmd // ProvaCmd should have listCmd added to it.

	b := new(bytes.Buffer)
	cmdToTest.SetOut(b)
	cmdToTest.SetErr(b)

	fullArgs := append([]string{"list"}, args...)
	cmdToTest.SetArgs(fullArgs)

	// Reset flags for listCmd to default values before each execution.
	var listCmdInstance *cobra.Command
	for _, cmd := range ProvaCmd.Commands() {
		if cmd.Name() == "list" {
			listCmdInstance = cmd
			break
		}
	}
	if listCmdInstance != nil {
		listCmdInstance.Flags().Visit(func(f *pflag.Flag) {
			f.Value.Set(f.DefValue)
		})
	}

	err := cmdToTest.Execute()
	return b.String(), err
}

// TestProvaListDefault tests the default output of 'prova list'.
func TestProvaListDefault(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	output, err := executeProvaListCommand()
	if err != nil {
		t.Fatalf("Esperado nenhum erro para listagem padrão, obteve: %v\nSaída: %s", err, output)
	}

	// Table header is already in Portuguese in list.go
	if !strings.Contains(output, "ID         | Título                              | Disciplina      | Data de Criação      | Nº Questões") {
		t.Errorf("Saída não contém o cabeçalho da tabela esperado. Obteve: %s", output)
	}

	if !strings.Contains(output, "prova123") {
		t.Errorf("Saída não contém o ID de prova de exemplo esperado 'prova123'. Obteve: %s", output)
	}
	if !strings.Contains(output, "prova456") {
		t.Errorf("Saída não contém o ID de prova de exemplo esperado 'prova456'. Obteve: %s", output)
	}

	idx789 := strings.Index(output, "prova789")
	idx123 := strings.Index(output, "prova123")

	if idx789 == -1 || idx123 == -1 || idx789 > idx123 {
		t.Errorf("Ordem de classificação padrão (created_at desc) parece incorreta ou itens não encontrados. Índice 'prova789': %d, índice 'prova123': %d. Saída:\n%s", idx789, idx123, output)
	}
}

// TestProvaListFilterBySubject tests filtering by subject.
func TestProvaListFilterBySubject(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	output, err := executeProvaListCommand("--subject", "História")
	if err != nil {
		t.Fatalf("Erro ao filtrar por disciplina 'História': %v\nSaída: %s", err, output)
	}

	if !strings.Contains(output, "prova456") {
		t.Errorf("Esperado 'prova456' (História) na saída, obteve: %s", output)
	}
	if strings.Contains(output, "prova123") {
		t.Errorf("Não esperado 'prova123' (Matemática) na saída filtrada por 'História'. Obteve: %s", output)
	}

	outputNoMatch, _ := executeProvaListCommand("--subject", "QuímicaAvançada")
	// Check for the specific Portuguese message from list.go
	if !strings.Contains(outputNoMatch, "Nenhuma prova encontrada com o filtro de disciplina: 'QuímicaAvançada'.") &&
	   !strings.Contains(outputNoMatch, "Nenhuma prova encontrada para os critérios especificados.") {
		t.Errorf("Esperada mensagem 'Nenhuma prova encontrada' para 'QuímicaAvançada', obteve: %s", outputNoMatch)
	}
}

// TestProvaListSort tests sorting functionality.
func TestProvaListSort(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	output, err := executeProvaListCommand("--sort-by", "title", "--order", "asc")
	if err != nil {
		t.Fatalf("Erro ao ordenar por título asc: %v\nSaída: %s", err, output)
	}

	idxAvaliacao := strings.Index(output, "prova456")
	idxProvaAvancada := strings.Index(output, "prova101")
	idxProvaBasica := strings.Index(output, "prova123")

	if !(idxAvaliacao < idxProvaAvancada && idxProvaAvancada < idxProvaBasica) {
		t.Errorf("Esperados títulos em ordem ascendente. Obteve índices: Avaliação (%d), Prova Avançada (%d), Prova Básica (%d). Saída:\n%s", idxAvaliacao, idxProvaAvancada, idxProvaBasica, output)
	}

	outputInvalidSort, _ := executeProvaListCommand("--sort-by", "campoInvalido")
	// Check for the specific Portuguese message from list.go
	if !strings.Contains(outputInvalidSort, "Critério de ordenação inválido: 'campoInvalido'. Utilizando 'created_at' como padrão.") {
		t.Errorf("Esperado aviso para critério de ordenação inválido. Obteve: %s", outputInvalidSort)
	}
}

// TestProvaListPagination tests pagination functionality.
func TestProvaListPagination(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	outputP1L1, err := executeProvaListCommand("--limit", "1", "--page", "1")
	if err != nil {
		t.Fatalf("Erro com limit 1 page 1: %v\nSaída: %s", err, outputP1L1)
	}
	if !strings.Contains(outputP1L1, "prova789") {
		t.Errorf("Esperado 'prova789' na página 1 com limite 1. Obteve: %s", outputP1L1)
	}
	if strings.Contains(outputP1L1, "prova202") {
		t.Errorf("Não esperado 'prova202' na página 1 com limite 1. Obteve: %s", outputP1L1)
	}
	// Check for Portuguese pagination string, e.g., "Página 1 de 5"
	if !strings.Contains(outputP1L1, "Página 1 de 5") {
		t.Errorf("Informação de paginação incorreta para página 1 limite 1. Obteve: %s", outputP1L1)
	}

	outputP2L1, err := executeProvaListCommand("--limit", "1", "--page", "2")
	if err != nil {
		t.Fatalf("Erro com limit 1 page 2: %v\nSaída: %s", err, outputP2L1)
	}
	if !strings.Contains(outputP2L1, "prova202") {
		t.Errorf("Esperado 'prova202' na página 2 com limite 1. Obteve: %s", outputP2L1)
	}
	if strings.Contains(outputP2L1, "prova789") {
		t.Errorf("Não esperado 'prova789' na página 2 com limite 1. Obteve: %s", outputP2L1)
	}
	if !strings.Contains(outputP2L1, "Página 2 de 5") {
		t.Errorf("Informação de paginação incorreta para página 2 limite 1. Obteve: %s", outputP2L1)
	}

	outputOOB, _ := executeProvaListCommand("--limit", "1", "--page", "10")
	// Check for Portuguese out of bounds message from list.go
	if !strings.Contains(outputOOB, "Página 10 fora do alcance") && !strings.Contains(outputOOB, "Nenhuma prova para exibir nesta página.") {
		t.Errorf("Esperada mensagem 'Página fora do alcance' ou 'Nenhuma prova para exibir' para página fora dos limites. Obteve: %s", outputOOB)
	}
}

// TestProvaListCombinedFlags tests a combination of filtering, sorting, and pagination.
func TestProvaListCombinedFlags(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	args := []string{
		"--subject", "Matemática",
		"--sort-by", "title",
		"--order", "asc",
		"--limit", "1",
		"--page", "1",
	}
	output, err := executeProvaListCommand(args...)
	if err != nil {
		t.Fatalf("Erro com flags combinadas: %v\nSaída: %s", err, output)
	}

	if !strings.Contains(output, "prova101") {
		t.Errorf("Esperado 'prova101' com flags combinadas. Obteve: %s", output)
	}
	if strings.Contains(output, "prova123") || strings.Contains(output, "prova202") {
		t.Errorf("Não esperadas outras provas de Matemática com limite 1. Obteve: %s", output)
	}
	if !strings.Contains(output, "Página 1 de 3") { // Total Matemática = 3
		t.Errorf("Informação de paginação incorreta para flags combinadas. Esperado 'Página 1 de 3'. Obteve: %s", output)
	}
}

// Helper to count rows in the output table (excluding header)
func countTableRows(output string) int {
	lines := strings.Split(output, "\n")
	count := 0
	rowRegex := regexp.MustCompile(`^[a-zA-Z0-9]+ *\|`)
	for _, line := range lines {
		if rowRegex.MatchString(strings.TrimSpace(line)) {
			count++
		}
	}
	return count
}

// Import pflag for resetting flags
// Moved to top import block
