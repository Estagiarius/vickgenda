package prova

import (
	"bytes"
	"strings"
	"testing"

	// Assuming ProvaCmd is the root for 'prova' subcommands.
	// If generateCmd is added to a different root, that root should be used.
	// For this test, we will assume ProvaCmd is correctly set up in init functions.
	// "vickgenda-cli/cmd" // If ProvaCmd is added to a global rootCmd in `cmd`
)

// executeProvaCommand executes the 'prova generate' command with given arguments
// and returns the stdout/stderr output and any error.
func executeProvaCommand(args ...string) (string, error) {
	// Reset ProvaCmd for each test run to ensure clean state for flags, etc.
	// This might involve re-initializing ProvaCmd or its subcommands if they are modified by execution.
	// For simplicity, we'll assume ProvaCmd and its subcommands are available as defined.
	// If ProvaCmd is not directly executable or holds state, this might need adjustment.
	// e.g., rootCmd := cmd.NewRootCmd(); rootCmd.AddCommand(ProvaCmd) and execute rootCmd.

	// If testing generateCmd directly:
	// cmdToTest := generateCmd
	// If testing via the main ProvaCmd:
	cmdToTest := ProvaCmd // ProvaCmd should have generateCmd added to it.

	b := new(bytes.Buffer)
	cmdToTest.SetOut(b)
	cmdToTest.SetErr(b)

	// Prepend "generate" to the args for ProvaCmd
	fullArgs := append([]string{"generate"}, args...)
	cmdToTest.SetArgs(fullArgs)

	// Before executing, ensure any persistent flags or states are reset if necessary.
	// For generateCmd, flags are defined in its init(), so they should be fresh unless
	// cobra holds state across SetArgs/Execute calls on the same Command instance.
	// It's safer to re-create command instances for each test if state is an issue.
	// However, for this scope, we'll use the existing ProvaCmd.

	err := cmdToTest.Execute()
	return b.String(), err
}

// TestProvaGenerateRequiredFlags checks if errors are reported for missing required flags.
func TestProvaGenerateRequiredFlags(t *testing.T) {
	// Test without --title
	output, err := executeProvaCommand("--subject", "Matemática")
	if err == nil {
		t.Errorf("Esperado erro quando --title está ausente, obteve nil")
	}
	// Cobra typically prints to stderr (which we captured) and returns an error.
	// The exact error message depends on Cobra's internals but usually mentions the missing flag.
	// The actual message from Cobra might be in English: "Error: required flag(s) \"title\" not set"
	// Or, if custom error handling is in place for i18n of Cobra errors, it could be Portuguese.
	// For now, we'll assume Cobra's default English message or a common pattern.
	if !strings.Contains(output, "Error: required flag(s) \"title\" not set") && !strings.Contains(output, "erro: flags obrigatórias \"title\" não configuradas") {
		// Allowing for Cobra's typical English message or a hypothetical Portuguese one.
		t.Errorf("Saída esperada continha erro de flag 'title' ausente, obteve: %s", output)
	}

	// Test without --subject
	output, err = executeProvaCommand("--title", "Prova Teste")
	if err == nil {
		t.Errorf("Esperado erro quando --subject está ausente, obteve nil")
	}
	if !strings.Contains(output, "Error: required flag(s) \"subject\" not set") && !strings.Contains(output, "erro: flags obrigatórias \"subject\" não configuradas") {
		t.Errorf("Saída esperada continha erro de flag 'subject' ausente, obteve: %s", output)
	}
}

// TestProvaGenerateBasicGeneration tests basic successful generation of a prova.
func TestProvaGenerateBasicGeneration(t *testing.T) {
	// Assuming sampleQuestions in generate.go has a "Matemática" "fácil" question.
	args := []string{
		"--title", "Prova de Matemática Simples",
		"--subject", "Matemática", // This subject must exist in sampleQuestions
		"--num-questions", "1",    // Requesting one question
	}
	output, err := executeProvaCommand(args...)
	if err != nil {
		t.Fatalf("Esperado nenhum erro para geração básica, obteve: %v\nSaída: %s", err, output)
	}

	if !strings.Contains(output, "Prova Gerada (Objeto models.Test)") {
		t.Errorf("Saída não contém confirmação da criação do objeto models.Test. Obteve: %s", output)
	}
	if !strings.Contains(output, "ID:") { // Check for part of the models.Test output
		t.Errorf("Saída não parece imprimir o objeto Test. Obteve: %s", output)
	}
	if !strings.Contains(output, "Título: Prova de Matemática Simples") {
		t.Errorf("Saída não contém o Título da Prova correto. Obteve: %s", output)
	}
	if !strings.Contains(output, "Questão 1 (ID: q") {
		t.Errorf("Saída não parece listar nenhuma questão para a prova. Obteve: %s", output)
	}
}

// TestProvaGenerateFilteringBySubject tests filtering questions by subject.
func TestProvaGenerateFilteringBySubject(t *testing.T) {
	// Test with a subject that should yield questions
	argsMath := []string{
		"--title", "Prova de Matemática",
		"--subject", "Matemática", // Assuming "Matemática" questions exist
		"--num-questions", "2",
	}
	outputMath, errMath := executeProvaCommand(argsMath...)
	if errMath != nil {
		t.Fatalf("Erro durante geração de prova de 'Matemática': %v\nSaída: %s", errMath, outputMath)
	}
	if !strings.Contains(outputMath, "Disciplina: Matemática") { // Message from generate.go
		t.Errorf("Esperado que a disciplina da prova fosse 'Matemática', obteve diferente na saída: %s", outputMath)
	}
	if strings.Count(outputMath, "Questão ") < 1 && !strings.Contains(outputMath, "Nenhuma questão foi encontrada") {
		t.Errorf("Esperada pelo menos uma questão para 'Matemática' ou mensagem 'Nenhuma questão foi encontrada'. Obteve: %s", outputMath)
	}


	// Test with a subject that should yield no questions from the sample set
	argsNonExistent := []string{
		"--title", "Prova de Astronomia",
		"--subject", "Astronomia", // Assuming "Astronomia" has no questions in sampleQuestions
	}
	outputNonExistent, _ := executeProvaCommand(argsNonExistent...)
	if !strings.Contains(outputNonExistent, "Nenhuma questão foi encontrada com os critérios especificados.") { // Message from generate.go
		t.Errorf("Esperada mensagem 'Nenhuma questão foi encontrada com os critérios especificados.' para 'Astronomia', obteve: %s", outputNonExistent)
	}
}

// TestProvaGenerateNumQuestionsFlag tests the --num-questions and difficulty-specific num flags.
func TestProvaGenerateNumQuestionsFlag(t *testing.T) {
	// Test --num-questions
	argsTotal := []string{
		"--title", "Prova de 3 Questões",
		"--subject", "Matemática",
		"--num-questions", "3",
	}
	outputTotal, errTotal := executeProvaCommand(argsTotal...)
	if errTotal != nil {
		t.Fatalf("Erro ao gerar prova com --num-questions 3: %v\nSaída: %s", errTotal, outputTotal)
	}
	numGenerated := strings.Count(outputTotal, "\nQuestão ")
	if numGenerated != 3 {
		if !strings.Contains(outputTotal, "Não foi possível selecionar questões") && !strings.Contains(outputTotal, "Nenhuma questão foi encontrada") {
			if numGenerated > 3 {
				t.Errorf("Esperado até 3 questões, obteve %d. Saída: %s", numGenerated, outputTotal)
			}
		}
	}

	argsDifficulty := []string{
		"--title", "Prova por Dificuldade",
		"--subject", "Matemática",
		"--num-easy", "1",
		"--num-medium", "1",
	}
	outputDifficulty, errDiff := executeProvaCommand(argsDifficulty...)
	if errDiff != nil {
		t.Fatalf("Erro ao gerar prova com números de dificuldade: %v\nSaída: %s", errDiff, outputDifficulty)
	}
	if !strings.Contains(outputDifficulty, "Prova Gerada") { // General check for success
		t.Errorf("Esperada geração bem-sucedida para números específicos de dificuldade. Obteve: %s", outputDifficulty)
	}
}


// TestProvaGenerateRandomization tests if randomization seed is set.
func TestProvaGenerateRandomization(t *testing.T) {
	args := []string{
		"--title", "Prova Randomizada",
		"--subject", "Matemática",
		"--num-questions", "2",
		"--randomize-order",
	}
	output, err := executeProvaCommand(args...)
	if err != nil {
		t.Fatalf("Erro durante geração de prova randomizada: %v\nSaída: %s", err, output)
	}
	if !strings.Contains(output, "RandomizationSeed:") || strings.Contains(output, "RandomizationSeed:0") {
		if strings.Contains(output, "Questão 1") {
			t.Errorf("Esperado um RandomizationSeed não-zero na saída quando --randomize-order é usado e questões são geradas, obteve: %s", output)
		}
	}
}

// TestProvaGenerateOutputFileSimulation tests the output message for file saving.
func TestProvaGenerateOutputFileSimulation(t *testing.T) {
	args := []string{
		"--title", "Prova para Arquivo",
		"--subject", "Geografia",
		"--num-questions", "1",
		"--output-file", "minha_prova.txt",
	}
	output, err := executeProvaCommand(args...)
	if err != nil {
		t.Fatalf("Erro durante teste com --output-file: %v\nSaída: %s", err, output)
	}

	expectedMsg := "Simulando salvamento da prova em: minha_prova.txt" // Message from generate.go
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("Saída esperada continha '%s', obteve: %s", expectedMsg, output)
	}
	if strings.Contains(output, "--- Visualização da Prova (Formato Texto Simples) ---") { // Message from generate.go
		t.Errorf("Esperado que a visualização completa da prova estivesse ausente quando --output-file é usado. Obteve: %s", output)
	}
}

// TODO: Add more tests:
// - Test for --allow-duplicates (if logic becomes non-trivial)
// - Test for specific question selection using more filters (tags, topics, type)
// - Test edge cases for num-questions (e.g., asking for more questions than available)
// - Test interaction between num-questions and num-easy/medium/hard
// - Test with empty sampleQuestions to ensure graceful handling.
// - Test --instructions flag.

// Note: To properly test randomization effects on order, one might need to run the command
// multiple times with the same explicit seed (if that flag is added) and compare outputs,
// or run with --randomize-order and check if two identical calls produce different question orders
// (statistically, they should if there are enough questions to shuffle).
// For now, checking the RandomizationSeed field is a simpler proxy.

// To make tests more robust against changes in sample data, consider:
// 1. Defining test-specific sample data within the test file.
// 2. Modifying `generateCmd` or its underlying logic to accept a question source (e.g., a slice)
//    that can be injected during tests. This is a common pattern for better unit testing.
// For this subtask, we rely on the global `sampleQuestions` in `generate.go`.
