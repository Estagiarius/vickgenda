package bancoq

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"vickgenda-cli/internal/db"
	"vickgenda-cli/internal/models"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var bancoqViewCmd = &cobra.Command{
	Use:   "view <ID_DA_QUESTAO>",
	Short: "Visualiza todos os detalhes de uma questão específica",
	Long: `Exibe todos os detalhes de uma questão específica do banco de dados,
identificada pelo seu ID. As informações são apresentadas de forma clara e legível.
Exemplo:
  vickgenda bancoq view 123e4567-e89b-12d3-a456-426614174000`,
	Args: cobra.ExactArgs(1), // Garante que exatamente um argumento (o ID) seja fornecido
	Run:  runViewQuestion,
}

func init() {
	BancoqCmd.AddCommand(bancoqViewCmd)
	// Nenhuma flag específica para este comando por enquanto.
}

func runViewQuestion(cmd *cobra.Command, args []string) {
	// A inicialização do DB agora é feita no PersistentPreRunE do BancoqCmd

	questionID := args[0]
	if strings.TrimSpace(questionID) == "" {
		fmt.Fprintln(os.Stderr, "Erro: O ID da questão não pode ser vazio.")
		os.Exit(1)
	}

	question, err := db.GetQuestion(questionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Checagem mais idiomática para sql.ErrNoRows
			fmt.Fprintf(os.Stderr, "Erro: A questão com ID '%s' não foi encontrada.\n", questionID)
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao buscar a questão com ID '%s': %v\n", questionID, err)
		}
		os.Exit(1)
		return // Redundante devido ao os.Exit(1), mas bom para clareza
	}

	fmt.Printf("Detalhes da Questão ID: %s\n", question.ID)
	fmt.Println(strings.Repeat("-", 40)) // Linha separadora

	// Usando tablewriter para uma exibição formatada como lista de definições (chave: valor)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(true) // Habilita quebra de linha automática para textos longos
	table.SetBorder(false)      // Sem bordas externas
	table.SetColumnSeparator(":") // Separador entre chave e valor
	table.SetHeaderLine(false)  // Sem linha de cabeçalho
	table.SetCenterSeparator("")  // Sem separador central (útil para outros modos de tabela)
	table.SetTablePadding("\t")   // Padding com tabulação para alinhar melhor
	table.SetAlignment(tablewriter.ALIGN_LEFT) // Alinhar todo o texto à esquerda

	// Adicionando dados à tabela
	data := [][]string{
		{"ID", question.ID},
		{"Disciplina", question.Subject},
		{"Tópico", question.Topic},
		{"Dificuldade", models.FormatDifficultyToPtBR(question.Difficulty)},
		{"Tipo", models.FormatQuestionTypeToPtBR(question.QuestionType)},
	}
	table.AppendBulk(data) // Adiciona os primeiros campos

	// Campo de Texto da Questão (pode ser longo)
	// Para o texto da questão, vamos adicioná-lo separadamente para melhor controle da quebra de linha,
	// ou garantir que SetAutoWrapText(true) funcione bem com a biblioteca.
	// Tablewriter com SetAutoWrapText(true) deve lidar bem.
	table.Append([]string{"Texto da Questão", question.QuestionText})

	// Opções de Resposta (se houver)
	if len(question.AnswerOptions) > 0 {
		optionsStr := new(strings.Builder)
		for i, opt := range question.AnswerOptions {
			fmt.Fprintf(optionsStr, "%c) %s", 'A'+i, opt)
			if i < len(question.AnswerOptions)-1 {
				optionsStr.WriteString("\n") // Nova linha para cada opção
			}
		}
		table.Append([]string{"Opções de Resposta", optionsStr.String()})
	}

	// Respostas Corretas
	if len(question.CorrectAnswers) > 0 {
		answersStr := new(strings.Builder)
		for i, ans := range question.CorrectAnswers {
			fmt.Fprintf(answersStr, "- %s", ans)
			if i < len(question.CorrectAnswers)-1 {
				answersStr.WriteString("\n") // Nova linha para cada resposta
			}
		}
		table.Append([]string{"Respostas Corretas", answersStr.String()})
	} else {
		table.Append([]string{"Respostas Corretas", "(Não especificadas)"})
	}

	// Campos opcionais e metadados
	optionalData := [][]string{}
	if question.Source != "" {
		optionalData = append(optionalData, []string{"Fonte", question.Source})
	}
	if len(question.Tags) > 0 {
		optionalData = append(optionalData, []string{"Tags", strings.Join(question.Tags, ", ")})
	}
	if question.Author != "" {
		optionalData = append(optionalData, []string{"Autor", question.Author})
	}
	optionalData = append(optionalData, []string{"Criada em", question.CreatedAt.Format("02/01/2006 às 15:04:05 MST")})
	optionalData = append(optionalData, []string{"Usada pela Última Vez", models.FormatLastUsedAt(question.LastUsedAt)})

	table.AppendBulk(optionalData)

	table.Render() // Renderiza a tabela
	fmt.Println(strings.Repeat("-", 40)) // Linha separadora no final
}
