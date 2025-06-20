package prova

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	// No direct cobra import needed here if ProvaCmd is globally accessible
	// "vickgenda-cli/internal/models" // Not directly used in test logic, but command uses it
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
	// This is important because Cobra commands retain flag values across executions.
	// listCmd is a child of ProvaCmd.
	// We need to find listCmd and reset its flags.
	// This assumes listCmd is already added to ProvaCmd.
	// A more robust way might be to re-initialize ProvaCmd and its subcommands for each test.
	var listCmdInstance *cobra.Command
	for _, cmd := range ProvaCmd.Commands() {
		if cmd.Name() == "list" {
			listCmdInstance = cmd
			break
		}
	}
	if listCmdInstance != nil {
		listCmdInstance.Flags().Visit(func(f *pflag.Flag) {
			f.Value.Set(f.DefValue) // Reset to default value
		})
	}


	err := cmdToTest.Execute()
	return b.String(), err
}

// TestProvaListDefault tests the default output of 'prova list'.
func TestProvaListDefault(t *testing.T) {
	// Reset sample data to its original state before this test block
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})


	output, err := executeProvaListCommand()
	if err != nil {
		t.Fatalf("Expected no error for default list, got: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "ID         | Título                              | Disciplina      | Data de Criação      | Nº Questões") {
		t.Errorf("Output does not contain the expected table header. Got: %s", output)
	}

	// Check for a known entry from sampleGeneratedProvas (e.g., "prova123")
	// This depends on the sample data in list.go
	if !strings.Contains(output, "prova123") {
		t.Errorf("Output does not contain expected sample test ID 'prova123'. Got: %s", output)
	}
	if !strings.Contains(output, "prova456") {
		t.Errorf("Output does not contain expected sample test ID 'prova456'. Got: %s", output)
	}

	// Default sort is created_at desc.
	// sampleGeneratedProvas:
	// {ID: "prova789", Title: "Teste Surpresa de Geografia", ..., CreatedAt: time.Now()}
	// {ID: "prova202", Title: "Revisão de Tópicos Matemáticos", ..., CreatedAt: time.Now().Add(-12 * time.Hour)}
	// {ID: "prova123", Title: "Prova de Matemática Básica", ..., CreatedAt: time.Now().Add(-24 * time.Hour)}
	// {ID: "prova456", Title: "Avaliação de História do Brasil", ..., CreatedAt: time.Now().Add(-48 * time.Hour)}
	// {ID: "prova101", Title: "Prova Avançada de Cálculo", ..., CreatedAt: time.Now().Add(-72 * time.Hour)}

	// So, "prova789" should appear before "prova123"
	idx789 := strings.Index(output, "prova789")
	idx123 := strings.Index(output, "prova123")

	if idx789 == -1 || idx123 == -1 || idx789 > idx123 {
		t.Errorf("Default sort order (created_at desc) seems incorrect or items not found. 'prova789' index: %d, 'prova123' index: %d. Output:\n%s", idx789, idx123, output)
	}
}

// TestProvaListFilterBySubject tests filtering by subject.
func TestProvaListFilterBySubject(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	// Filter by "História"
	output, err := executeProvaListCommand("--subject", "História")
	if err != nil {
		t.Fatalf("Error filtering by subject 'História': %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "prova456") { // Belongs to História
		t.Errorf("Expected 'prova456' (História) in output, got: %s", output)
	}
	if strings.Contains(output, "prova123") { // Belongs to Matemática
		t.Errorf("Did not expect 'prova123' (Matemática) in 'História' filtered output. Got: %s", output)
	}

	// Filter by a subject with no tests
	outputNoMatch, errNoMatch := executeProvaListCommand("--subject", "QuímicaAvançada")
	if errNoMatch != nil {
		// The command itself doesn't error, it just prints a message.
		// Depending on implementation, errNoMatch might be nil.
		// t.Fatalf("Error filtering by subject 'QuímicaAvançada': %v\nOutput: %s", errNoMatch, outputNoMatch)
	}
	if !strings.Contains(outputNoMatch, "Nenhuma prova encontrada com o filtro de disciplina: 'QuímicaAvançada'") &&
	   !strings.Contains(outputNoMatch, "Nenhuma prova encontrada para os critérios especificados.") { // Fallback message if pagination results in none
		t.Errorf("Expected 'Nenhuma prova encontrada' message for 'QuímicaAvançada', got: %s", outputNoMatch)
	}
}

// TestProvaListSort tests sorting functionality.
func TestProvaListSort(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	// Sort by title asc
	// Titles: "Avaliação de História do Brasil" (prova456), "Prova Avançada de Cálculo" (prova101),
	// "Prova de Matemática Básica" (prova123), "Revisão de Tópicos Matemáticos" (prova202), "Teste Surpresa de Geografia" (prova789)
	output, err := executeProvaListCommand("--sort-by", "title", "--order", "asc")
	if err != nil {
		t.Fatalf("Error sorting by title asc: %v\nOutput: %s", err, output)
	}

	idxAvaliacao := strings.Index(output, "prova456") // Avaliação
	idxProvaAvancada := strings.Index(output, "prova101") // Prova Avançada
	idxProvaBasica := strings.Index(output, "prova123") // Prova de Matemática

	if !(idxAvaliacao < idxProvaAvancada && idxProvaAvancada < idxProvaBasica) {
		t.Errorf("Expected titles in ascending order. Got indices: Avaliação (%d), Prova Avançada (%d), Prova Básica (%d). Output:\n%s", idxAvaliacao, idxProvaAvancada, idxProvaBasica, output)
	}

	// Test invalid sort-by value (should default to created_at and print a message)
	outputInvalidSort, _ := executeProvaListCommand("--sort-by", "invalidField")
	if !strings.Contains(outputInvalidSort, "Critério de ordenação inválido: 'invalidField'. Usando 'created_at'.") {
		t.Errorf("Expected warning for invalid sort-by value. Got: %s", outputInvalidSort)
	}
}

// TestProvaListPagination tests pagination functionality.
func TestProvaListPagination(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	// Assuming at least 3 sample provas. Let's use the default sort (created_at desc)
	// Prova IDs by created_at desc: prova789, prova202, prova123, prova456, prova101

	// Page 1, Limit 1
	outputP1L1, err := executeProvaListCommand("--limit", "1", "--page", "1")
	if err != nil {
		t.Fatalf("Error with limit 1 page 1: %v\nOutput: %s", err, outputP1L1)
	}
	if !strings.Contains(outputP1L1, "prova789") { // First item by created_at desc
		t.Errorf("Expected 'prova789' on page 1 limit 1. Got: %s", outputP1L1)
	}
	if strings.Contains(outputP1L1, "prova202") {
		t.Errorf("Did not expect 'prova202' on page 1 limit 1. Got: %s", outputP1L1)
	}
	if !strings.Contains(outputP1L1, "Página 1 de 5") { // Assuming 5 total sample items
		t.Errorf("Page info incorrect for page 1 limit 1. Got: %s", outputP1L1)
	}


	// Page 2, Limit 1
	outputP2L1, err := executeProvaListCommand("--limit", "1", "--page", "2")
	if err != nil {
		t.Fatalf("Error with limit 1 page 2: %v\nOutput: %s", err, outputP2L1)
	}
	if !strings.Contains(outputP2L1, "prova202") { // Second item
		t.Errorf("Expected 'prova202' on page 2 limit 1. Got: %s", outputP2L1)
	}
	if strings.Contains(outputP2L1, "prova789") {
		t.Errorf("Did not expect 'prova789' on page 2 limit 1. Got: %s", outputP2L1)
	}
	if !strings.Contains(outputP2L1, "Página 2 de 5") {
		t.Errorf("Page info incorrect for page 2 limit 1. Got: %s", outputP2L1)
	}

	// Page out of bounds
	outputOOB, _ := executeProvaListCommand("--limit", "1", "--page", "10") // Assuming only 5 sample items
	if !strings.Contains(outputOOB, "Página 10 fora do alcance") && !strings.Contains(outputOOB, "Nenhuma prova para exibir nesta página.") {
		t.Errorf("Expected 'Página fora do alcance' or 'Nenhuma prova para exibir' message for out-of-bounds page. Got: %s", outputOOB)
	}
}

// TestProvaListCombinedFlags tests a combination of filtering, sorting, and pagination.
func TestProvaListCombinedFlags(t *testing.T) {
	originalSampleProvas := make([]models.Test, len(sampleGeneratedProvas))
	copy(originalSampleProvas, sampleGeneratedProvas)
	t.Cleanup(func() {
		sampleGeneratedProvas = originalSampleProvas
	})

	// Filter by Matemática, sort by title asc, limit 1, page 1
	// Matemática titles: "Prova Avançada de Cálculo" (prova101), "Prova de Matemática Básica" (prova123), "Revisão de Tópicos Matemáticos" (prova202)
	// Sorted by title asc: "Prova Avançada de Cálculo" (prova101) should be first.
	args := []string{
		"--subject", "Matemática",
		"--sort-by", "title",
		"--order", "asc",
		"--limit", "1",
		"--page", "1",
	}
	output, err := executeProvaListCommand(args...)
	if err != nil {
		t.Fatalf("Error with combined flags: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "prova101") { // "Prova Avançada de Cálculo"
		t.Errorf("Expected 'prova101' with combined flags. Got: %s", output)
	}
	if strings.Contains(output, "prova123") || strings.Contains(output, "prova202") {
		t.Errorf("Did not expect other Matemática provas with limit 1. Got: %s", output)
	}
	// Total de provas (Matemática) = 3. TotalPages = (3+1-1)/1 = 3
	if !strings.Contains(output, "Página 1 de 3") {
		t.Errorf("Page info incorrect for combined flags. Expected 'Página 1 de 3'. Got: %s", output)
	}
}

// Helper to count rows in the output table (excluding header)
func countTableRows(output string) int {
	lines := strings.Split(output, "\n")
	count := 0
	// Regex to match a line that seems like a data row (starts with an ID, has pipes)
	// This is fragile and depends on the exact format.
	// Example ID: prova123 (not just digits)
	rowRegex := regexp.MustCompile(`^[a-zA-Z0-9]+ *\|`)
	for _, line := range lines {
		if rowRegex.MatchString(strings.TrimSpace(line)) {
			count++
		}
	}
	return count
}

// Note: The `sampleGeneratedProvas` is modified by some operations in `list.go` (e.g. sorting is in-place).
// For robust tests, we should ensure this sample data is reset before each test or test group,
// or the functions in list.go should operate on a copy.
// Added t.Cleanup to reset sampleGeneratedProvas.
// Also, `pflag` state can persist. Added reset for listCmd flags.

// Import pflag for resetting flags
import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"vickgenda-cli/internal/models"
)
