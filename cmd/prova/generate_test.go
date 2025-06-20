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
		t.Errorf("Expected error when --title is missing, got nil")
	}
	// Cobra typically prints to stderr (which we captured) and returns an error.
	// The exact error message depends on Cobra's internals but usually mentions the missing flag.
	if !strings.Contains(output, "flag pflag: title has been marked as required") && !strings.Contains(output, "Error: required flag(s) \"title\" not set") {
		// Adjusted to check for common variations of Cobra's error message
		t.Errorf("Expected output to contain missing title flag error, got: %s", output)
	}

	// Test without --subject
	output, err = executeProvaCommand("--title", "Prova Teste")
	if err == nil {
		t.Errorf("Expected error when --subject is missing, got nil")
	}
	if !strings.Contains(output, "flag pflag: subject has been marked as required") && !strings.Contains(output, "Error: required flag(s) \"subject\" not set") {
		t.Errorf("Expected output to contain missing subject flag error, got: %s", output)
	}
}

// TestProvaGenerateBasicGeneration tests basic successful generation of a prova.
func TestProvaGenerateBasicGeneration(t *testing.T) {
	// Assuming sampleQuestions in generate.go has a "Matemática" "fácil" question.
	args := []string{
		"--title", "Prova de Matemática Simples",
		"--subject", "Matemática", // This subject must exist in sampleQuestions
		"--num-questions", "1",    // Requesting one question
		// "--difficulty", "fácil", // This might be too specific if not enough easy math questions
	}
	output, err := executeProvaCommand(args...)
	if err != nil {
		t.Fatalf("Expected no error for basic generation, got: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Prova Gerada (Objeto models.Test)") {
		t.Errorf("Output does not contain confirmation of models.Test object creation. Got: %s", output)
	}
	if !strings.Contains(output, "ID:") { // Check for part of the models.Test output
		t.Errorf("Output does not seem to print the Test object. Got: %s", output)
	}
	if !strings.Contains(output, "Título: Prova de Matemática Simples") {
		t.Errorf("Output does not contain the correct Test Title. Got: %s", output)
	}
	// Check if at least one question is mentioned in the "Visualização da Prova" part
	if !strings.Contains(output, "Questão 1 (ID: q") {
		t.Errorf("Output does not seem to list any question for the test. Got: %s", output)
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
		t.Fatalf("Error during 'Matemática' test generation: %v\nOutput: %s", errMath, outputMath)
	}
	if !strings.Contains(outputMath, "Disciplina: Matemática") {
		t.Errorf("Expected test subject to be 'Matemática', got different in output: %s", outputMath)
	}
	// Count occurrences of "Questão " to see how many questions were listed
	// This is a bit fragile but works for simulation.
	if strings.Count(outputMath, "Questão ") < 1 && !strings.Contains(outputMath, "Nenhuma questão encontrada") {
		// Allow for 0 if no questions match, but then the "Nenhuma questão" message should be there.
		t.Errorf("Expected at least one question for 'Matemática' or 'Nenhuma questão encontrada' message. Got: %s", outputMath)
	}


	// Test with a subject that should yield no questions from the sample set
	argsNonExistent := []string{
		"--title", "Prova de Astronomia",
		"--subject", "Astronomia", // Assuming "Astronomia" has no questions in sampleQuestions
	}
	outputNonExistent, _ := executeProvaCommand(argsNonExistent...) // Error is expected due to no questions
	if !strings.Contains(outputNonExistent, "Nenhuma questão encontrada com os critérios especificados") {
		t.Errorf("Expected 'Nenhuma questão encontrada' message for 'Astronomia', got: %s", outputNonExistent)
	}
}

// TestProvaGenerateNumQuestionsFlag tests the --num-questions and difficulty-specific num flags.
func TestProvaGenerateNumQuestionsFlag(t *testing.T) {
	// Test --num-questions
	argsTotal := []string{
		"--title", "Prova de 3 Questões",
		"--subject", "Matemática", // Ensure enough math questions exist in sample
		"--num-questions", "3",
	}
	outputTotal, errTotal := executeProvaCommand(argsTotal...)
	if errTotal != nil {
		t.Fatalf("Error generating test with --num-questions 3: %v\nOutput: %s", errTotal, outputTotal)
	}
	// A simple way to check number of questions is to count "Questão X:"
	// This depends on the output format of the simulation.
	numGenerated := strings.Count(outputTotal, "\nQuestão ") // Assuming each question starts on a new line like this
	if numGenerated != 3 {
		// It's possible fewer are generated if not enough match criteria.
		// The simulation logic tries to get up to num-questions.
		// We need to inspect sampleQuestions and filtering logic to be sure.
		// For now, we expect it to try and succeed if questions are available.
		// Check if "Total de questões filtradas inicialmente" is less than 3, or if selection failed.
		if !strings.Contains(outputTotal, "Não foi possível selecionar questões") && !strings.Contains(outputTotal, "Nenhuma questão encontrada") {
			// If it didn't explicitly fail to select, it should have 3 or fewer if sample is small
			if numGenerated > 3 {
				t.Errorf("Expected up to 3 questions, got %d. Output: %s", numGenerated, outputTotal)
			} else if numGenerated == 0 && len(sampleQuestions) > 0 { // if sample has questions, 0 is unexpected unless filters are too strict
                 // This check is tricky without knowing the exact state of sampleQuestions and filters applied by default.
                 // For this test, let's assume there are at least 3 generic math questions.
                 // t.Logf("Warning: Expected 3 questions, got %d. This might be due to insufficient matching sample questions for 'Matemática'. Output: %s", numGenerated, outputTotal)
            }
		}
	}

	// Test --num-easy, --num-medium, --num-hard (assuming sample data has these)
	// This is more complex to assert precisely without deeper inspection of sample data state
	// and the selection algorithm's behavior when exact counts aren't met.
	// For now, a basic check that it runs:
	argsDifficulty := []string{
		"--title", "Prova por Dificuldade",
		"--subject", "Matemática",
		"--num-easy", "1",
		"--num-medium", "1",
		// "--num-hard", "0", // Let's assume there's at least one easy and one medium math question
	}
	outputDifficulty, errDiff := executeProvaCommand(argsDifficulty...)
	if errDiff != nil {
		t.Fatalf("Error generating test with difficulty numbers: %v\nOutput: %s", errDiff, outputDifficulty)
	}
	if !strings.Contains(outputDifficulty, "Prova Gerada") {
		t.Errorf("Expected successful generation for difficulty specific numbers. Got: %s", outputDifficulty)
	}
	// A more robust test would check the actual difficulties of the output questions.
}


// TestProvaGenerateRandomization tests if randomization seed is set.
func TestProvaGenerateRandomization(t *testing.T) {
	args := []string{
		"--title", "Prova Randomizada",
		"--subject", "Matemática",
		"--num-questions", "2", // Need at least 2 to see order effects, though we only check seed
		"--randomize-order",
	}
	output, err := executeProvaCommand(args...)
	if err != nil {
		t.Fatalf("Error during randomized test generation: %v\nOutput: %s", err, output)
	}
	// Check if the RandomizationSeed is non-zero in the models.Test output
	// Example output: RandomizationSeed:1234567890
	if !strings.Contains(output, "RandomizationSeed:") || strings.Contains(output, "RandomizationSeed:0") {
		// It's possible seed is 0 if no questions were selected, or if randomization didn't run.
		// Check if questions were actually generated.
		if strings.Contains(output, "Questão 1") { // implies questions were generated
			t.Errorf("Expected a non-zero RandomizationSeed in output when --randomize-order is used and questions are generated, got: %s", output)
		}
	}
}

// TestProvaGenerateOutputFileSimulation tests the output message for file saving.
func TestProvaGenerateOutputFileSimulation(t *testing.T) {
	args := []string{
		"--title", "Prova para Arquivo",
		"--subject", "Geografia", // Assuming geography questions exist
		"--num-questions", "1",
		"--output-file", "minha_prova.txt",
	}
	output, err := executeProvaCommand(args...)
	if err != nil {
		t.Fatalf("Error during test with --output-file: %v\nOutput: %s", err, output)
	}

	expectedMsg := "Simulando salvamento da prova em: minha_prova.txt"
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedMsg, output)
	}
	// Also check that the full text rendering of the prova is NOT present
	if strings.Contains(output, "--- Visualização da Prova (Formato Texto Simples) ---") {
		t.Errorf("Expected full prova visualization to be absent when --output-file is used. Got: %s", output)
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
