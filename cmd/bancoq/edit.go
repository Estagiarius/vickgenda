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
)

var bancoqEditCmd = &cobra.Command{
	Use:   "edit <ID_DA_QUESTAO>",
	Short: "Edita uma questão existente no banco de dados",
	Long: `Permite editar interativamente os campos de uma questão existente,
identificada pelo seu ID. Os valores atuais são apresentados e podem ser alterados.
Para campos de texto simples, deixar o campo vazio e pressionar Enter manterá o valor atual.
Para listas (opções, respostas, tags), a edição é mais interativa.`,
	Args: cobra.ExactArgs(1),
	Run:  runEditQuestion,
}

func init() {
	BancoqCmd.AddCommand(bancoqEditCmd)
}

// editCollectSliceItemsInteractive provides a more user-friendly way to edit a slice of strings.
// It starts with currentItems and allows adding, removing, or clearing.
func editCollectSliceItemsInteractive(promptMessageSingular, fieldNameForMessages string, currentItems []string, required bool) []string {
	items := make([]string, len(currentItems))
	copy(items, currentItems)

	fmt.Printf("\n--- Editando '%s' ---\n", fieldNameForMessages)
	if len(items) > 0 {
		fmt.Println("Itens atuais:")
		for i, item := range items {
			fmt.Printf("  %d: %s\n", i+1, item)
		}
	} else {
		fmt.Println("Nenhum item atualmente.")
	}
	fmt.Println("Opções: [a]dicionar, [r]emover item pelo número, [c]lear todos, [f]inalizar edição deste campo.")

	for {
		var action string
		actionPrompt := &survey.Input{Message: fmt.Sprintf("Ação para '%s' (a, r, c, f):", fieldNameForMessages)}
		err := survey.AskOne(actionPrompt, &action)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro no prompt de ação: %v\n", err)
			continue // Re-ask
		}
		action = strings.ToLower(strings.TrimSpace(action))

		switch action {
		case "a", "adicionar":
			var newItem string
			newItemPrompt := &survey.Input{Message: promptMessageSingular + ":"}
			err := survey.AskOne(newItemPrompt, &newItem)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao adicionar novo item: %v\n", err)
				continue
			}
			trimmedNewItem := strings.TrimSpace(newItem)
			if trimmedNewItem != "" {
				items = append(items, trimmedNewItem)
				fmt.Printf("Item '%s' adicionado.\n", trimmedNewItem)
			} else {
				fmt.Println("Nenhum item adicionado (entrada vazia).")
			}
		case "r", "remover":
			if len(items) == 0 {
				fmt.Println("Nenhum item para remover.")
				continue
			}
			var itemNumberStr string
			itemNumPrompt := &survey.Input{Message: fmt.Sprintf("Número do item a remover (1-%d):", len(items))}
			err := survey.AskOne(itemNumPrompt, &itemNumberStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao obter número do item: %v\n", err)
				continue
			}
			itemNumber, convErr := survey.StdAsk.Atoi(itemNumberStr)
			if convErr != nil || itemNumber < 1 || itemNumber > len(items) {
				fmt.Println("Número inválido.")
				continue
			}
			removedItem := items[itemNumber-1]
			items = append(items[:itemNumber-1], items[itemNumber:]...)
			fmt.Printf("Item '%s' removido.\n", removedItem)
		case "c", "clear":
			var confirmClear bool
			confirmPrompt := &survey.Confirm{Message: fmt.Sprintf("Tem certeza que deseja remover TODOS os itens de '%s'?", fieldNameForMessages), Default: false}
			survey.AskOne(confirmPrompt, &confirmClear)
			if confirmClear {
				items = []string{}
				fmt.Println("Todos os itens foram removidos.")
			}
		case "f", "finalizar":
			if required && len(items) == 0 {
				fmt.Printf("'%s' é um campo obrigatório e não pode estar vazio. Por favor, adicione pelo menos um item ou cancele a edição (Ctrl+C).\n", fieldNameForMessages)
				continue // Don't allow finalizing if required and empty
			}
			fmt.Printf("--- Finalizada a edição de '%s' ---\n", fieldNameForMessages)
			return items
		default:
			fmt.Println("Ação inválida. Use 'a', 'r', 'c' ou 'f'.")
		}

		// Display current state after action
		if len(items) > 0 {
			fmt.Println("Itens atuais:")
			for i, item := range items {
				fmt.Printf("  %d: %s\n", i+1, item)
			}
		} else {
			fmt.Println("Nenhum item atualmente.")
		}
	}
}

func runEditQuestion(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	questionID := args[0]
	q, err := db.GetQuestion(questionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Fprintf(os.Stderr, "Erro: A questão com ID '%s' não foi encontrada.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao buscar a questão com ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return
	}

	originalQuestionType := q.QuestionType // Store original type for logic later

	fmt.Printf("Editando questão com ID: %s\n", q.ID)
	fmt.Println("(Deixe o campo em branco e pressione Enter para manter o valor atual para campos de texto simples)")
	fmt.Println(strings.Repeat("-", 30))

	// String fields (Subject, Topic)
	survey.AskOne(&survey.Input{Message: "Disciplina:", Default: q.Subject}, &q.Subject, survey.WithValidator(survey.Required))
	survey.AskOne(&survey.Input{Message: "Tópico:", Default: q.Topic}, &q.Topic, survey.WithValidator(survey.Required))

	// Difficulty (Select)
	q.Difficulty = askSelect("Nível de dificuldade:", q.Difficulty,
		[]string{models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard},
		func(val string) string { return models.FormatDifficultyToPtBR(val) })

	// QuestionType (Select)
	q.QuestionType = askSelect("Tipo da questão:", q.QuestionType,
		[]string{models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer},
		func(val string) string { return models.FormatQuestionTypeToPtBR(val) })

	// QuestionText (Editor)
	newQuestionText := q.QuestionText
	survey.AskOne(&survey.Editor{
		Message:       "Texto da questão (pressione Enter para abrir o editor):",
		Default:       q.QuestionText,
		HideDefault:   true,
		AppendDefault: true, // Loads default into editor
		Help:          "O texto atual será carregado no editor. Ctrl+D para salvar, Esc para cancelar.",
	}, &newQuestionText, survey.WithValidator(survey.Required))
	q.QuestionText = strings.TrimSpace(newQuestionText)


	// Slice fields: AnswerOptions, CorrectAnswers, Tags
	// Only edit AnswerOptions if applicable to the NEW or ORIGINAL question type
	isNowMcOrTf := q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse
	wasOriginallyMcOrTf := originalQuestionType == models.QuestionTypeMultipleChoice || originalQuestionType == models.QuestionTypeTrueFalse

	if isNowMcOrTf || wasOriginallyMcOrTf { // If type is or was MC/TF, allow editing options
		if confirmEditField(fmt.Sprintf("Opções de Resposta (atuais: %d)", len(q.AnswerOptions))) {
			if isNowMcOrTf {
				q.AnswerOptions = editCollectSliceItemsInteractive("Nova opção de resposta", "Opções de Resposta", q.AnswerOptions, true) // Required if type is MC/TF
				if len(q.AnswerOptions) == 0 { // Double check, editCollect should handle 'required'
					fmt.Fprintln(os.Stderr, "Erro: Opções de resposta são obrigatórias para este tipo de questão.")
					os.Exit(1)
				}
			} else { // Type changed from MC/TF to something else
				fmt.Println("Tipo da questão alterado, limpando opções de resposta.")
				q.AnswerOptions = []string{}
			}
		}
	} else { // Not MC/TF now and wasn't before
		q.AnswerOptions = []string{} // Ensure it's empty
	}


	if confirmEditField(fmt.Sprintf("Respostas Corretas (atuais: %d)", len(q.CorrectAnswers))) {
		q.CorrectAnswers = editCollectSliceItemsInteractive("Nova resposta correta", "Respostas Corretas", q.CorrectAnswers, true) // Always required
		if len(q.CorrectAnswers) == 0 { // Double check
			fmt.Fprintln(os.Stderr, "Erro: Pelo menos uma resposta correta é obrigatória.")
			os.Exit(1)
		}
	}
	// Validate CorrectAnswers against AnswerOptions if multiple choice
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
				fmt.Fprintf(os.Stderr, "Erro: A resposta correta '%s' não está entre as opções de resposta fornecidas.\n", ans)
				fmt.Println("Por favor, verifique as opções e respostas corretas.")
				os.Exit(1)
			}
		}
	}


	if confirmEditField(fmt.Sprintf("Tags (atuais: %d)", len(q.Tags))) {
		q.Tags = editCollectSliceItemsInteractive("Nova tag", "Tags", q.Tags, false) // Not required
	}

	// Optional string fields (Source, Author)
	survey.AskOne(&survey.Input{Message: "Fonte (opcional):", Default: q.Source}, &q.Source)
	survey.AskOne(&survey.Input{Message: "Autor (opcional):", Default: q.Author}, &q.Author)

	// LastUsedAt is not typically edited by the user directly. CreatedAt is preserved.

	err = db.UpdateQuestion(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao atualizar a questão: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nQuestão com ID '%s' atualizada com sucesso.\n", questionID)
}

// confirmEditField asks user if they want to edit a particular field.
func confirmEditField(fieldNameWithCurrentState string) bool {
	var confirm bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Deseja editar o campo: %s?", fieldNameWithCurrentState),
		Default: false, // Default to not editing
	}
	survey.AskOne(prompt, &confirm)
	return confirm
}

// askSelect is a helper for survey.Select, handling mapping between codes and display values.
func askSelect(message, currentCodeValue string, codeOptions []string, formatFunc func(string) string) string {
	displayOptions := make([]string, len(codeOptions))
	displayToCodeMap := make(map[string]string)
	var currentDisplayValue string

	for i, code := range codeOptions {
		display := formatFunc(code)
		displayOptions[i] = display
		displayToCodeMap[display] = code
		if code == currentCodeValue {
			currentDisplayValue = display
		}
	}

	var selectedDisplayValue string
	prompt := &survey.Select{
		Message: message,
		Options: displayOptions,
		Default: currentDisplayValue, // survey.Select expects the Default to be one of the Options
	}
	survey.AskOne(prompt, &selectedDisplayValue)

	// Map back to code, or return original if something went wrong (though survey should prevent invalid selection)
	if code, ok := displayToCodeMap[selectedDisplayValue]; ok {
		return code
	}
	return currentCodeValue // Fallback, should not happen with survey.Select
}
