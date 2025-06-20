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
	Long:  `Permite editar interativamente os campos de uma questão existente, identificada pelo seu ID. Os valores atuais são apresentados e podem ser alterados.`,
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
	fmt.Println("Instruções: Digite um novo item e pressione Enter. Para remover o último item adicionado/existente, digite '-'. Para limpar todos os itens, digite '--clear'. Deixe vazio e pressione Enter para finalizar a adição/edição deste campo.")


	for {
		var item string
		// Ajustar a mensagem do prompt para ser mais clara no contexto de edição
		itemPromptMessage := fmt.Sprintf("%s (atual: %d itens. '-' para remover, '--clear' para limpar, Enter para finalizar):", promptMessage, len(items))
		if len(items) == 0 && len(currentItems) > 0 && items == nil { // Se 'items' ainda não foi populado, mas 'currentItems' existe
			items = append(items, currentItems...) // Começar com os itens atuais para edição
			fmt.Printf("Itens carregados para edição: %s. Prossiga para adicionar, remover ou finalizar.\n", strings.Join(items, ", "))
			itemPromptMessage = fmt.Sprintf("%s (carregado: %d itens. '-' para remover, '--clear' para limpar, Enter para finalizar):", promptMessage, len(items))
		}


		itemPrompt := &survey.Input{Message: itemPromptMessage}

		if err := survey.AskOne(itemPrompt, &item); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao coletar item para '%s': %v\n", fieldName, err)
			os.Exit(1)
		}

		trimmedItem := strings.TrimSpace(item)

		if trimmedItem == "" {
			break // Finaliza a adição/edição para este campo
		}
		if trimmedItem == "--clear" {
			items = []string{} // Limpa todos os itens
			fmt.Println("Todos os itens foram removidos.")
			// Considerar se deve continuar no loop para adicionar novos itens ou finalizar.
			// Para consistência, finalizar aqui. O usuário pode re-editar se necessário.
			break
		}
		if trimmedItem == "-" {
			if len(items) > 0 {
				removed := items[len(items)-1]
				items = items[:len(items)-1]
				fmt.Printf("Item '%s' removido. Itens restantes para '%s': %s\n", removed, fieldName, strings.Join(items, ", "))
			} else {
				fmt.Println("Nenhum item para remover.")
			}
			continue
		}

		items = append(items, trimmedItem)
		fmt.Printf("Item '%s' adicionado/mantido. Itens atuais para '%s': %s\n", trimmedItem, fieldName, strings.Join(items, ", "))
	}

	// Validação específica para campos obrigatórios como 'Respostas Corretas'
	if fieldName == "Respostas Corretas" && len(items) == 0 {
		fmt.Println("Atenção: O campo 'Respostas Corretas' não pode ficar vazio. Por favor, adicione pelo menos uma resposta.")
		// Chama recursivamente para garantir que o usuário adicione pelo menos uma resposta.
		// Passa um slice vazio como `currentItems` para evitar repopular com os antigos se o usuário limpou e tentou sair.
		return editCollectSliceItems(promptMessage, fieldName, []string{})
	}
	return items
}


func runEditQuestion(cmd *cobra.Command, args []string) {
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	questionID := args[0]
	q, err := db.GetQuestion(questionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			fmt.Fprintf(os.Stderr, "Erro: A questão com ID '%s' não foi encontrada.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao buscar a questão com ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return
	}

	fmt.Printf("Editando questão com ID: %s (Deixe o campo em branco e pressione Enter para manter o valor atual, exceto para listas que têm manejo especial).\n\n", q.ID)

	// String fields
	survey.AskOne(&survey.Input{Message: "Disciplina:", Default: q.Subject}, &q.Subject, survey.WithValidator(survey.Required))
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
	survey.AskOne(&survey.Editor{Message: "Texto da questão (pressione Enter para abrir o editor, Ctrl+D para salvar, Esc para cancelar):", Default: q.QuestionText, HideDefault: true, AppendDefault: true, Help: "Use Markdown para formatação se desejar. O texto atual será carregado no editor."}, &q.QuestionText, survey.WithValidator(survey.Required))


	// Slice fields - using a confirm-then-re-enter approach
	isMultipleChoiceType := q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse
	if confirmReEnter("Opções de Resposta", q.AnswerOptions, isMultipleChoiceType) {
		if isMultipleChoiceType {
			q.AnswerOptions = editCollectSliceItems("Nova opção de resposta", "Opções de Resposta", q.AnswerOptions)
		} else {
			fmt.Println("Opções de resposta não são aplicáveis para este tipo de questão e serão limpas.")
			q.AnswerOptions = []string{}
		}
	}

	// CorrectAnswers are always applicable and required
	if confirmReEnter("Respostas Corretas", q.CorrectAnswers, true) { // true because it's always applicable/editable
		q.CorrectAnswers = editCollectSliceItems("Nova resposta correta", "Respostas Corretas", q.CorrectAnswers)
		if len(q.CorrectAnswers) == 0 {
			fmt.Fprintln(os.Stderr, "Erro crítico: Pelo menos uma resposta correta deve ser fornecida. A edição não pode prosseguir sem isso.")
			os.Exit(1)
		}
	}


	if confirmReEnter("Tags", q.Tags, true) {
		q.Tags = editCollectSliceItems("Nova tag", "Tags", q.Tags)
	}

	// Optional string fields
	survey.AskOne(&survey.Input{Message: "Fonte (opcional):", Default: q.Source}, &q.Source)
	survey.AskOne(&survey.Input{Message: "Autor (opcional):", Default: q.Author}, &q.Author)


	err = db.UpdateQuestion(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao atualizar a questão: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Questão com ID '%s' atualizada com sucesso.\n", questionID)
}

// Helper function to confirm if user wants to edit a slice field
func confirmReEnter(fieldName string, currentItems []string, applicable bool) bool {
	if !applicable { // If the field is not applicable (e.g. AnswerOptions for Essay)
		// fmt.Printf("Campo '%s' não aplicável para o tipo de questão atual.\n", fieldName) // Optional: user feedback
		return false // Don't even ask to edit
	}

	currentValueDisplay := "(vazio)"
	if len(currentItems) > 0 {
		currentValueDisplay = strings.Join(currentItems, "; ")
	}

	message := fmt.Sprintf("O campo '%s' atualmente contém: [%s]. Deseja editar este campo?", fieldName, currentValueDisplay)

	confirm := false
	prompt := &survey.Confirm{
		Message: message,
		Default: false, // Default to not editing
		Help:    fmt.Sprintf("Se escolher 'sim', você poderá adicionar, remover ou limpar os itens de '%s'. Se 'não', os itens atuais serão mantidos.", fieldName),
	}
	err := survey.AskOne(prompt, &confirm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro no prompt de confirmação para '%s': %v. O campo não será editado.\n", fieldName, err)
		return false // Default to not editing on error
	}
	return confirm
}
