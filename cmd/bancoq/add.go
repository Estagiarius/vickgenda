package bancoq

import (
	"fmt"
	"os"
	"strings"
	"time"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var bancoqAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adiciona uma nova questão ao banco",
	Long: `Permite adicionar uma nova questão ao banco de questões.
Pode ser usado de forma interativa, solicitando cada campo, ou de forma não-interativa através de flags.
Exemplo não-interativo:
  vickgenda bancoq add --subject "Matemática" --topic "Álgebra" --difficulty "medium" --type "multiple_choice" --question "Qual o valor de x em 2x = 4?" --option "x = 1" --option "x = 2" --answer "x = 2" --source "Livro Y, p. 10" --tag "Básico" --tag "Equação"
`,
	Run: runAddQuestion,
}

// Struct to hold flag values for the add command
var addQuestionFlags struct {
	Subject        string
	Topic          string
	Difficulty     string
	QuestionType   string
	QuestionText   string
	AnswerOptions  []string
	CorrectAnswers []string
	Source         string
	Tags           []string
	Author         string
}

func init() {
	BancoqCmd.AddCommand(bancoqAddCmd)

	// Flags for non-interactive mode
	bancoqAddCmd.Flags().StringVarP(&addQuestionFlags.Subject, "subject", "s", "", "Disciplina da questão (obrigatório)")
	bancoqAddCmd.Flags().StringVarP(&addQuestionFlags.Topic, "topic", "t", "", "Tópico da questão (obrigatório)")
	bancoqAddCmd.Flags().StringVarP(&addQuestionFlags.Difficulty, "difficulty", "d", "", "Nível de dificuldade (easy, medium, hard) (obrigatório)")
	bancoqAddCmd.Flags().StringVarP(&addQuestionFlags.QuestionType, "type", "q", "", "Tipo da questão (multiple_choice, true_false, essay, short_answer) (obrigatório)")
	bancoqAddCmd.Flags().StringVarP(&addQuestionFlags.QuestionText, "question", "x", "", "Texto da questão (obrigatório)")
	bancoqAddCmd.Flags().StringSliceVarP(&addQuestionFlags.AnswerOptions, "option", "o", []string{}, "Opção de resposta (para multiple_choice, true_false). Use múltiplas vezes para várias opções.")
	bancoqAddCmd.Flags().StringSliceVarP(&addQuestionFlags.CorrectAnswers, "answer", "a", []string{}, "Resposta(s) correta(s). Use múltiplas vezes para várias respostas corretas (obrigatório).")
	bancoqAddCmd.Flags().StringVar(&addQuestionFlags.Source, "source", "", "Fonte da questão (opcional)")
	bancoqAddCmd.Flags().StringSliceVar(&addQuestionFlags.Tags, "tag", []string{}, "Tag para a questão (opcional). Use múltiplas vezes para várias tags.")
	bancoqAddCmd.Flags().StringVar(&addQuestionFlags.Author, "author", "", "Autor da questão (opcional)")
}

// isValidDifficulty checks if the provided difficulty is valid.
func isValidDifficulty(difficulty string) bool {
	switch difficulty {
	case models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard:
		return true
	default:
		return false
	}
}

// isValidQuestionType checks if the provided question type is valid.
func isValidQuestionType(qType string) bool {
	switch qType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer:
		return true
	default:
		return false
	}
}

func runAddQuestion(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	q := models.Question{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
	}

	// Determine mode: if any core flag is explicitly set by the user, assume non-interactive.
	// Core flags: subject, topic, difficulty, type, question, answer.
	nonInteractive := cmd.Flags().Changed("subject") ||
		cmd.Flags().Changed("topic") ||
		cmd.Flags().Changed("difficulty") ||
		cmd.Flags().Changed("type") ||
		cmd.Flags().Changed("question") ||
		cmd.Flags().Changed("answer")

	if nonInteractive {
		// --- Non-interactive mode: Populate from flags and validate ---
		errorMessages := []string{}

		if addQuestionFlags.Subject == "" {
			errorMessages = append(errorMessages, "--subject é obrigatório.")
		}
		if addQuestionFlags.Topic == "" {
			errorMessages = append(errorMessages, "--topic é obrigatório.")
		}
		if addQuestionFlags.Difficulty == "" {
			errorMessages = append(errorMessages, "--difficulty é obrigatório.")
		} else if !isValidDifficulty(addQuestionFlags.Difficulty) {
			errorMessages = append(errorMessages, fmt.Sprintf("--difficulty inválido. Use um de: %s, %s, %s.", models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard))
		}
		if addQuestionFlags.QuestionType == "" {
			errorMessages = append(errorMessages, "--type é obrigatório.")
		} else if !isValidQuestionType(addQuestionFlags.QuestionType) {
			errorMessages = append(errorMessages, fmt.Sprintf("--type inválido. Use um de: %s, %s, %s, %s.", models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer))
		}
		if addQuestionFlags.QuestionText == "" {
			errorMessages = append(errorMessages, "--question é obrigatório.")
		}
		if len(addQuestionFlags.CorrectAnswers) == 0 {
			errorMessages = append(errorMessages, "Pelo menos uma --answer é obrigatória.")
		}

		// Validate AnswerOptions based on QuestionType
		if addQuestionFlags.QuestionType == models.QuestionTypeMultipleChoice || addQuestionFlags.QuestionType == models.QuestionTypeTrueFalse {
			if len(addQuestionFlags.AnswerOptions) == 0 {
				errorMessages = append(errorMessages, fmt.Sprintf("Para o tipo '%s', pelo menos uma --option é obrigatória.", addQuestionFlags.QuestionType))
			}
			// Ensure correct answers are among options for multiple choice
			if addQuestionFlags.QuestionType == models.QuestionTypeMultipleChoice {
				for _, ans := range addQuestionFlags.CorrectAnswers {
					found := false
					for _, opt := range addQuestionFlags.AnswerOptions {
						if ans == opt {
							found = true
							break
						}
					}
					if !found {
						errorMessages = append(errorMessages, fmt.Sprintf("Resposta correta '%s' não está entre as opções fornecidas.", ans))
					}
				}
			}
		}


		if len(errorMessages) > 0 {
			fmt.Fprintln(os.Stderr, "Erro(s) ao adicionar questão no modo não-interativo:")
			for _, msg := range errorMessages {
				fmt.Fprintf(os.Stderr, "- %s\n", msg)
			}
			cmd.Help() // Show help
			os.Exit(1)
		}

		q.Subject = addQuestionFlags.Subject
		q.Topic = addQuestionFlags.Topic
		q.Difficulty = addQuestionFlags.Difficulty
		q.QuestionType = addQuestionFlags.QuestionType
		q.QuestionText = addQuestionFlags.QuestionText
		q.AnswerOptions = addQuestionFlags.AnswerOptions
		q.CorrectAnswers = addQuestionFlags.CorrectAnswers
		q.Source = addQuestionFlags.Source
		q.Tags = addQuestionFlags.Tags
		q.Author = addQuestionFlags.Author

	} else {
		// --- Interactive mode ---
		fmt.Println("Adicionando nova questão (modo interativo)...")

		prompts := []*survey.Question{
			{Name: "Subject", Prompt: &survey.Input{Message: "Disciplina:"}, Validate: survey.Required},
			{Name: "Topic", Prompt: &survey.Input{Message: "Tópico:"}, Validate: survey.Required},
			{
				Name: "Difficulty",
				Prompt: &survey.Select{
					Message: "Nível de dificuldade:",
					Options: []string{models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard},
					Description: func(value string, index int) string { return models.FormatDifficultyToPtBR(value) },
				},
				Validate: survey.Required,
			},
			{
				Name: "QuestionType",
				Prompt: &survey.Select{
					Message: "Tipo da questão:",
					Options: []string{models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer},
					Description: func(value string, index int) string { return models.FormatQuestionTypeToPtBR(value) },
				},
				Validate: survey.Required,
			},
			{
				Name:     "QuestionText",
				Prompt:   &survey.Editor{Message: "Texto da questão (pressione Enter para abrir o editor):", FileName: "vickgenda_questao_*.md", HideDefault: true, AppendDefault: true},
				Validate: survey.Required,
			},
		}
		answers := struct { // Temporary struct for survey answers
			Subject      string
			Topic        string
			Difficulty   string
			QuestionType string
			QuestionText string
		}{}
		if err := survey.Ask(prompts, &answers); err != nil {
			fmt.Fprintf(os.Stderr, "Erro no formulário interativo: %v\n", err)
			os.Exit(1)
		}
		q.Subject = answers.Subject
		q.Topic = answers.Topic
		q.Difficulty = answers.Difficulty
		q.QuestionType = answers.QuestionType
		q.QuestionText = strings.TrimSpace(answers.QuestionText)


		if q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse {
			q.AnswerOptions = collectSliceItemsInteractively("Opção de resposta (deixe vazio e pressione Enter para terminar de adicionar opções):", "Adicionar outra opção de resposta?", false)
			if len(q.AnswerOptions) == 0 {
				fmt.Fprintf(os.Stderr, "Para o tipo '%s', pelo menos uma opção de resposta deve ser fornecida.\n", models.FormatQuestionTypeToPtBR(q.QuestionType))
				os.Exit(1)
			}
		}

		q.CorrectAnswers = collectSliceItemsInteractively("Resposta correta (deixe vazio e pressione Enter para terminar de adicionar respostas):", "Adicionar outra resposta correta?", true)
		if len(q.CorrectAnswers) == 0 {
			fmt.Fprintln(os.Stderr, "Pelo menos uma resposta correta é obrigatória.")
			os.Exit(1)
		}

        // Validate correct answers against options for multiple choice
        if q.QuestionType == models.QuestionTypeMultipleChoice {
            validAnswers := []string{}
            for _, ans := range q.CorrectAnswers {
                found := false
                for _, opt := range q.AnswerOptions {
                    if ans == opt {
                        found = true
                        break
                    }
                }
                if !found {
                    fmt.Fprintf(os.Stderr, "Atenção: A resposta correta '%s' não está entre as opções fornecidas. Por favor, adicione-a como uma opção ou corrija a resposta.\n", ans)
                    // Re-collect options or answers, or simply error out
                    // For simplicity, we'll error out here. More complex UX could allow correction.
                    os.Exit(1)
                }
                validAnswers = append(validAnswers, ans)
            }
            q.CorrectAnswers = validAnswers
        }


		optionalPrompts := []*survey.Question{
			{Name: "Source", Prompt: &survey.Input{Message: "Fonte da questão (opcional):"}},
			{Name: "Author", Prompt: &survey.Input{Message: "Autor da questão (opcional):"}},
		}
		optionalAnswers := struct{ Source, Author string }{}
		if err := survey.Ask(optionalPrompts, &optionalAnswers); err != nil {
			fmt.Fprintf(os.Stderr, "Erro no formulário interativo (opcionais): %v\n", err)
			os.Exit(1)
		}
		q.Source = optionalAnswers.Source
		q.Author = optionalAnswers.Author

		q.Tags = collectSliceItemsInteractively("Tag (opcional; deixe vazio e pressione Enter para terminar de adicionar tags):", "Adicionar outra tag?", false)
	}

	// Finalize and save
	newID, err := db.CreateQuestion(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao adicionar questão ao banco de dados: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Questão adicionada com ID: %s\n", newID)
}

// collectSliceItemsInteractively collects multiple string items for a slice field using survey.
// `isRequired` means at least one item must be provided.
func collectSliceItemsInteractively(initialMessage, confirmMessage string, isRequired bool) []string {
	var items []string
	for {
		var item string
		itemPrompt := &survey.Input{Message: initialMessage}
		if err := survey.AskOne(itemPrompt, &item); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao coletar item: %v\n", err)
			os.Exit(1)
		}

		item = strings.TrimSpace(item)
		if item == "" {
			if isRequired && len(items) == 0 {
				fmt.Println("Este campo é obrigatório. Pelo menos um item deve ser adicionado.")
				continue // Re-prompt for the first item
			}
			break // Empty input signals end of list for optional or if already has items
		}

		items = append(items, item)

		var addMore bool
		confirmPrompt := &survey.Confirm{Message: confirmMessage, Default: true}
		if err := survey.AskOne(confirmPrompt, &addMore); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na confirmação: %v\n", err)
			os.Exit(1)
		}
		if !addMore {
			break
		}
	}
	return items
}
