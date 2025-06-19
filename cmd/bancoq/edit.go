package bancoq

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	// "time" // Not directly needed for edit logic, CreatedAt is preserved, LastUsedAt not edited here

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	// "github.com/google/uuid" // Not needed, ID is not changed
)

var bancoqEditCmd = &cobra.Command{
	Use:   "edit <ID_DA_QUESTAO>",
	Short: "Edita uma questão existente no banco de dados",
	Long:  `Permite editar interativamente os campos de uma questão existente, dado o seu ID.`,
	Args:  cobra.ExactArgs(1),
	Run:   runEditQuestion,
}

func init() {
	BancoqCmd.AddCommand(bancoqEditCmd)
}

// Adapted from add.go - simplified for editing contexts
func editCollectSliceItems(promptMessage string, fieldName string, currentItems []string) []string {
	var items []string
	fmt.Printf("Editando '%s'. Itens atuais: %s\n", fieldName, strings.Join(currentItems, ", "))

	for {
		var item string
		itemPrompt := &survey.Input{Message: fmt.Sprintf("%s (deixe vazio para terminar de adicionar, ou digite '-' para remover o último, ou '--clear' para limpar todos):", promptMessage)}

		if err := survey.AskOne(itemPrompt, &item); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao coletar item: %v\n", err)
			os.Exit(1)
		}

		trimmedItem := strings.TrimSpace(item)

		if trimmedItem == "" {
			break // Finish adding items
		}
		if trimmedItem == "--clear" {
			items = []string{} // Clear all items
			fmt.Println("Todos os itens foram removidos.")
			// We can either break or continue to add new items after clearing
			// For now, let's break and if they want to add, they can re-run edit or we can add another loop
			break
		}
		if trimmedItem == "-" {
			if len(items) > 0 {
				removed := items[len(items)-1]
				items = items[:len(items)-1]
				fmt.Printf("Item '%s' removido. Itens restantes: %s\n", removed, strings.Join(items, ", "))
			} else {
				fmt.Println("Nenhum item para remover.")
			}
			continue
		}

		items = append(items, trimmedItem)
		fmt.Printf("Item '%s' adicionado. Itens atuais: %s\n", trimmedItem, strings.Join(items, ", "))
	}
	if len(items) == 0 && (fieldName == "Respostas Corretas" || fieldName == "Opções de Resposta") {
		// If essential slices become empty, prompt again or handle based on requirements
		// For CorrectAnswers, it must not be empty. For AnswerOptions, it depends on QuestionType.
		if fieldName == "Respostas Corretas" {
			fmt.Println("Atenção: 'Respostas Corretas' não pode ser vazio. Por favor, adicione pelo menos uma resposta.")
			// Re-call or loop until at least one is added. For simplicity, let's prompt once more.
			return editCollectSliceItems(promptMessage, fieldName, items) // Recursive call to ensure it's not empty
		}
	}
	return items
}


func runEditQuestion(cmd *cobra.Command, args []string) {
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	questionID := args[0]
	q, err := db.GetQuestion(questionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			fmt.Fprintf(os.Stderr, "Erro: Questão com ID '%s' não encontrada.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao buscar questão ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return
	}

	fmt.Printf("Editando questão ID: %s (Deixe o campo em branco e pressione Enter para manter o valor atual, exceto para listas).\n\n", q.ID)

	// String fields
	survey.AskOne(&survey.Input{Message: "Matéria:", Default: q.Subject}, &q.Subject, survey.WithValidator(survey.Required))
	survey.AskOne(&survey.Input{Message: "Tópico:", Default: q.Topic}, &q.Topic, survey.WithValidator(survey.Required))

	// Difficulty
	difficultyOptions := []string{models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard}
	difficultyDisplayMap := map[string]string{
		models.DifficultyEasy:   models.FormatDifficultyToPtBR(models.DifficultyEasy),
		models.DifficultyMedium: models.FormatDifficultyToPtBR(models.DifficultyMedium),
		models.DifficultyHard:   models.FormatDifficultyToPtBR(models.DifficultyHard),
	}
	displayDifficultyOptions := make([]string, len(difficultyOptions))
	for i, code := range difficultyOptions {
		displayDifficultyOptions[i] = difficultyDisplayMap[code]
	}
	currentDifficultyDisplay := difficultyDisplayMap[q.Difficulty]
	var selectedDifficultyDisplay string
	survey.AskOne(&survey.Select{Message: "Nível de dificuldade:", Options: displayDifficultyOptions, Default: currentDifficultyDisplay}, &selectedDifficultyDisplay)
	for code, display := range difficultyDisplayMap {
		if display == selectedDifficultyDisplay {
			q.Difficulty = code
			break
		}
	}

	// QuestionType
	typeOptions := []string{models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer}
	typeDisplayMap := map[string]string{
		models.QuestionTypeMultipleChoice: models.FormatQuestionTypeToPtBR(models.QuestionTypeMultipleChoice),
		models.QuestionTypeTrueFalse:      models.FormatQuestionTypeToPtBR(models.QuestionTypeTrueFalse),
		models.QuestionTypeEssay:          models.FormatQuestionTypeToPtBR(models.QuestionTypeEssay),
		models.QuestionTypeShortAnswer:    models.FormatQuestionTypeToPtBR(models.QuestionTypeShortAnswer),
	}
	displayTypeOptions := make([]string, len(typeOptions))
	for i, code := range typeOptions {
		displayTypeOptions[i] = typeDisplayMap[code]
	}
	currentTypeDisplay := typeDisplayMap[q.QuestionType]
	var selectedTypeDisplay string
	survey.AskOne(&survey.Select{Message: "Tipo da questão:", Options: displayTypeOptions, Default: currentTypeDisplay}, &selectedTypeDisplay)
	for code, display := range typeDisplayMap {
		if display == selectedTypeDisplay {
			q.QuestionType = code
			break
		}
	}

	// QuestionText
	survey.AskOne(&survey.Editor{Message: "Texto da questão:", Default: q.QuestionText, HideDefault: true, AppendDefault: true}, &q.QuestionText, survey.WithValidator(survey.Required))

	// Slice fields - using a confirm-then-re-enter approach
	if confirmReEnter("Opções de Resposta", q.AnswerOptions, q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse) {
		if q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse {
			q.AnswerOptions = editCollectSliceItems("Opção de resposta", "Opções de Resposta", q.AnswerOptions)
		} else {
			fmt.Println("Opções de resposta não são aplicáveis para este tipo de questão. Serão limpas.")
			q.AnswerOptions = []string{}
		}
	}

	// CorrectAnswers are always applicable and required
	if confirmReEnter("Respostas Corretas", q.CorrectAnswers, true) { // true because it's always applicable/editable
		q.CorrectAnswers = editCollectSliceItems("Resposta correta", "Respostas Corretas", q.CorrectAnswers)
		if len(q.CorrectAnswers) == 0 {
			fmt.Fprintln(os.Stderr, "Erro: Pelo menos uma resposta correta deve ser fornecida.")
			os.Exit(1) // or re-prompt, but for now exit
		}
	}


	if confirmReEnter("Tags", q.Tags, true) {
		q.Tags = editCollectSliceItems("Tag", "Tags", q.Tags)
	}

	// Optional string fields
	survey.AskOne(&survey.Input{Message: "Fonte:", Default: q.Source}, &q.Source)
	survey.AskOne(&survey.Input{Message: "Autor:", Default: q.Author}, &q.Author)

	// Note: CreatedAt and LastUsedAt are not typically edited directly by the user.
	// LastUsedAt would be updated when a question is part of a generated test.

	err = db.UpdateQuestion(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao atualizar questão: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Questão '%s' atualizada com sucesso.\n", questionID)
}

// Helper function to confirm if user wants to edit a slice field
func confirmReEnter(fieldName string, currentItems []string, applicable bool) bool {
	if !applicable { // If the field is not applicable (e.g. AnswerOptions for Essay)
		return false // Don't even ask to edit
	}

	currentValueDisplay := "(nenhum)"
	if len(currentItems) > 0 {
		currentValueDisplay = strings.Join(currentItems, "; ")
	}

	message := fmt.Sprintf("%s atuais: %s. Deseja editar?", fieldName, currentValueDisplay)

	confirm := false
	prompt := &survey.Confirm{
		Message: message,
		Default: false, // Default to not editing
	}
	err := survey.AskOne(prompt, &confirm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro no prompt de confirmação: %v\n", err)
		return false // Default to not editing on error
	}
	return confirm
}
