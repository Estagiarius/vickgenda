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
	Long:  `Permite adicionar uma nova questão ao banco de questões. Pode ser usado de forma interativa, solicitando cada campo, ou de forma não-interativa através de flags.`,
	Run:   runAddQuestion,
}

// Struct to hold flag values
var questionFlags struct {
	Subject       string
	Topic         string
	Difficulty    string
	QuestionType  string
	QuestionText  string
	AnswerOptions []string
	CorrectAnswers []string
	Source        string
	Tags          []string
	Author        string
}

func init() {
	// fmt.Println("DEBUG: init() in cmd/bancoq/add.go called") // DEBUG line removed
	BancoqCmd.AddCommand(bancoqAddCmd)

	// Flags for non-interactive mode
	bancoqAddCmd.Flags().StringVarP(&questionFlags.Subject, "subject", "s", "", "Disciplina da questão (obrigatório em modo não-interativo)")
	bancoqAddCmd.Flags().StringVarP(&questionFlags.Topic, "topic", "t", "", "Tópico da questão (obrigatório em modo não-interativo)")
	bancoqAddCmd.Flags().StringVarP(&questionFlags.Difficulty, "difficulty", "d", "", "Nível de dificuldade (easy, medium, hard) (obrigatório em modo não-interativo)")
	bancoqAddCmd.Flags().StringVarP(&questionFlags.QuestionType, "type", "q", "", "Tipo da questão (multiple_choice, true_false, essay, short_answer) (obrigatório em modo não-interativo)")
	bancoqAddCmd.Flags().StringVarP(&questionFlags.QuestionText, "question", "x", "", "Texto da questão (obrigatório em modo não-interativo)")
	bancoqAddCmd.Flags().StringSliceVarP(&questionFlags.AnswerOptions, "option", "o", []string{}, "Opção de resposta (para multiple_choice, true_false). Pode ser usado múltiplas vezes")
	bancoqAddCmd.Flags().StringSliceVarP(&questionFlags.CorrectAnswers, "answer", "a", []string{}, "Resposta(s) correta(s). Pode ser usado múltiplas vezes (obrigatório em modo não-interativo)")
	bancoqAddCmd.Flags().StringVar(&questionFlags.Source, "source", "", "Fonte da questão (opcional)")
	bancoqAddCmd.Flags().StringSliceVar(&questionFlags.Tags, "tag", []string{}, "Tag para a questão (opcional). Pode ser usado múltiplas vezes")
	bancoqAddCmd.Flags().StringVar(&questionFlags.Author, "author", "", "Autor da questão (opcional)")
}

func runAddQuestion(cmd *cobra.Command, args []string) {
	// fmt.Println("DEBUG: runAddQuestion called") // DEBUG line removed
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados: %v\n", err)
		os.Exit(1)
	}

	q := models.Question{}

	// Determine mode: if any core flag is set, assume non-interactive
	// Core flags: subject, topic, difficulty, type, question, answer
	nonInteractive := cmd.Flags().Changed("subject") ||
		cmd.Flags().Changed("topic") ||
		cmd.Flags().Changed("difficulty") ||
		cmd.Flags().Changed("type") ||
		cmd.Flags().Changed("question") || // Using 'question' for questionText flag
		cmd.Flags().Changed("answer")

	if nonInteractive {
		// Non-interactive mode: Populate from flags
		if questionFlags.Subject == "" || questionFlags.Topic == "" || questionFlags.Difficulty == "" || questionFlags.QuestionType == "" || questionFlags.QuestionText == "" || len(questionFlags.CorrectAnswers) == 0 {
			fmt.Fprintln(os.Stderr, "Erro: No modo não-interativo, as flags --subject, --topic, --difficulty, --type, --question, e pelo menos uma --answer são obrigatórias.")
			cmd.Help() // Show help
			os.Exit(1)
		}
		q.Subject = questionFlags.Subject
		q.Topic = questionFlags.Topic
		q.Difficulty = questionFlags.Difficulty
		q.QuestionType = questionFlags.QuestionType
		q.QuestionText = questionFlags.QuestionText
		q.AnswerOptions = questionFlags.AnswerOptions
		q.CorrectAnswers = questionFlags.CorrectAnswers
		q.Source = questionFlags.Source
		q.Tags = questionFlags.Tags
		q.Author = questionFlags.Author
	} else {
		// Interactive mode
		fmt.Println("Adicionando nova questão (modo interativo)...")

		// commonStringValidator := survey.WithValidator(survey.Required) // Removed as unused
		// commonSliceValidator := survey.WithValidator(func(ans interface{}) error { // Removed as unused
		// 	if sl, ok := ans.([]string); ok {
		// 		if len(sl) == 0 {
		// 			return fmt.Errorf("pelo menos uma resposta correta é necessária")
		// 		}
		// 	} else if str, ok := ans.(string); ok { // For single string inputs that become part of a slice
		// 		if str == "" {
		// 			return fmt.Errorf("este campo não pode ser vazio se você está adicionando um item")
		// 		}
		// 	}
		// 	return nil
		// })

		prompts := []*survey.Question{
			{
				Name:   "Subject",
				Prompt: &survey.Input{Message: "Disciplina:"},
				Validate: survey.Required,
			},
			{
				Name:   "Topic",
				Prompt: &survey.Input{Message: "Tópico:"},
				Validate: survey.Required,
			},
			{
				Name: "Difficulty",
				Prompt: &survey.Select{
					Message: "Nível de dificuldade:",
					Options: []string{models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard},
					Description: func(value string, index int) string {
						switch value {
						case models.DifficultyEasy: return "Fácil"
						case models.DifficultyMedium: return "Médio"
						case models.DifficultyHard: return "Difícil"
						}
						return value
					},
				},
				Validate: survey.Required,
			},
			{
				Name: "QuestionType",
				Prompt: &survey.Select{
					Message: "Tipo da questão:",
					Options: []string{models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeEssay, models.QuestionTypeShortAnswer},
					Description: func(value string, index int) string {
						switch value {
						case models.QuestionTypeMultipleChoice: return "Múltipla Escolha"
						case models.QuestionTypeTrueFalse: return "Verdadeiro/Falso"
						case models.QuestionTypeEssay: return "Dissertativa (Ensaio)"
						case models.QuestionTypeShortAnswer: return "Resposta Curta"
						}
						return value
					},
				},
				Validate: survey.Required,
			},
			{
				Name:   "QuestionText",
				Prompt: &survey.Editor{Message: "Texto da questão (use editor externo se preferir e cole aqui):", FileName: "nova_questao_*.md"},
				Validate: survey.Required,
			},
		}
		// Using a temporary struct for survey answermap
		answers := struct {
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
		q.QuestionText = answers.QuestionText

		// Handle AnswerOptions (if multiple_choice or true_false)
		if q.QuestionType == models.QuestionTypeMultipleChoice || q.QuestionType == models.QuestionTypeTrueFalse {
			q.AnswerOptions = collectSliceItems("Opção de resposta (deixe vazio e pressione Enter para terminar):", "Adicionar outra opção de resposta?")
		}

		// Handle CorrectAnswers - always required
		q.CorrectAnswers = collectSliceItems("Resposta correta (pelo menos uma é necessária; deixe vazio e pressione Enter para terminar após a primeira):", "Adicionar outra resposta correta?")
		if len(q.CorrectAnswers) == 0 {
		    // Re-ask if empty, as it's required.
		    fmt.Println("Pelo menos uma resposta correta é obrigatória.")
		    q.CorrectAnswers = collectSliceItems("Resposta correta (pelo menos uma é necessária; deixe vazio e pressione Enter para terminar após a primeira):", "Adicionar outra resposta correta?")
		    if len(q.CorrectAnswers) == 0 {
		         fmt.Fprintln(os.Stderr, "Erro: Pelo menos uma resposta correta deve ser fornecida.")
		         os.Exit(1)
		    }
		}


		// Optional fields
		optionalPrompts := []*survey.Question{
			{Name: "Source", Prompt: &survey.Input{Message: "Fonte da questão (opcional):"}},
			{Name: "Author", Prompt: &survey.Input{Message: "Autor da questão (opcional):"}},
		}
		optionalAnswers := struct {Source string; Author string}{}
		if err := survey.Ask(optionalPrompts, &optionalAnswers); err != nil {
			fmt.Fprintf(os.Stderr, "Erro no formulário interativo (opcionais): %v\n", err)
			os.Exit(1)
		}
		q.Source = optionalAnswers.Source
		q.Author = optionalAnswers.Author

		q.Tags = collectSliceItems("Tag (opcional; deixe vazio e pressione Enter para terminar):", "Adicionar outra tag?")
	}

	// Finalize and save
	if q.ID == "" { // Should always be empty unless future logic changes
		q.ID = uuid.NewString()
	}
	if q.CreatedAt.IsZero() { // Should always be zero here
		q.CreatedAt = time.Now()
	}

	// Ensure difficulty and type are stored in English, even if prompt showed Portuguese
	// This is already handled by survey.Select if `Options` are the English constants.

	newID, err := db.CreateQuestion(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao adicionar questão ao banco de dados: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Questão adicionada com ID: %s\n", newID)
}

// Helper function to collect multiple string items for a slice field
func collectSliceItems(initialMessage string, confirmMessage string) []string {
	var items []string
	for {
		var item string
		prompt := &survey.Input{Message: initialMessage}
		if err := survey.AskOne(prompt, &item); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao coletar item: %v\n", err)
			os.Exit(1) // or handle error more gracefully
		}

		if strings.TrimSpace(item) == "" && len(items) > 0 { // Allow empty only if at least one item was added (for optional, or to finish)
			break
		}
		// Allow empty first entry to skip for optional fields or answer options (which can be empty e.g. for essay)
		if strings.TrimSpace(item) == "" && len(items) == 0 && (strings.Contains(initialMessage, "opcional") || strings.Contains(initialMessage, "Opção de resposta")) {
		    break
		}
		// Special handling for CorrectAnswers if first entry is empty - it's required.
        if strings.TrimSpace(item) == "" && len(items) == 0 && strings.Contains(initialMessage, "Resposta correta") {
            fmt.Println("Pelo menos uma resposta correta é necessária. Tente novamente.")
            continue // Re-prompt for the first correct answer
        }


		if strings.TrimSpace(item) != "" {
			items = append(items, strings.TrimSpace(item))
		}

		// Only ask to add another if the current item was not empty or if it's the first item for a required field (like CorrectAnswers).
		if strings.TrimSpace(item) != "" || (len(items) == 1 && strings.Contains(initialMessage, "Resposta correta")) {
			var addMore bool
			confirmPrompt := &survey.Confirm{Message: confirmMessage, Default: true}
			if err := survey.AskOne(confirmPrompt, &addMore); err != nil {
				fmt.Fprintf(os.Stderr, "Erro na confirmação: %v\n", err)
				os.Exit(1) // or handle error
			}
			if !addMore {
				break
			}
		} else if strings.TrimSpace(item) == "" && len(items) == 0 && !strings.Contains(initialMessage, "opcional") && !strings.Contains(initialMessage, "Opção de resposta") {
            // If a required field's first entry is empty (and not handled by specific logic above), break to avoid potential infinite loop.
            // The main validation for required fields should ideally catch this before attempting to save.
            break
        }

	}
	return items
}
