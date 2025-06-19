package squad4

import (
	"fmt"
	"log"
	"strings"
	"time" // Required for date formatting if not already present

	"vickgenda-cli/internal/commands/tarefa"
	"vickgenda-cli/internal/models"

	"github.com/spf13/cobra"
)

var RelembrarCmd = &cobra.Command{
	Use:   "relembrar",
	Short: "Gerencia lembretes (tarefas com a tag 'lembrete').",
	Long:  `Permite adicionar e listar lembretes. Lembretes são tarefas marcadas com a tag 'lembrete'.`,
}

var relembrarAdicionarCmd = &cobra.Command{
	Use:   "adicionar \"<lembrete>\" <data_YYYY-MM-DD> [hora_HH:MM]",
	Short: "Adiciona um novo lembrete.",
	Long: `Adiciona um novo lembrete (tarefa) com uma descrição, data e, opcionalmente, uma hora.
A descrição do lembrete deve estar entre aspas.
Formato da data: YYYY-MM-DD (ex: 2024-07-23) ou deixe em branco para nenhum prazo.
Formato da hora (opcional): HH:MM. Se fornecida, será anexada à descrição.`,
	Example: `relembrar adicionar "Comprar canetas vermelhas" 2024-07-21 10:00
relembrar adicionar "Buscar provas na gráfica" 2024-07-23`,
	Args: cobra.MinimumNArgs(1), // lembrete é obrigatório, data é opcional mas precisa de "" se hora for usada
	Run: func(cmd *cobra.Command, args []string) {
		descricaoLembrete := args[0]
		data := ""
		if len(args) > 1 {
			data = args[1] // Expected "YYYY-MM-DD" or empty
		}
		hora := ""
		if len(args) > 2 {
			hora = args[2]
		}

		finalDescription := descricaoLembrete
		if hora != "" {
			finalDescription = fmt.Sprintf("%s (Hora: %s)", descricaoLembrete, hora)
		}

		// Default priority (e.g., 2 for Medium, adjust as needed)
		priority := 2
		tags := []string{"lembrete"}

		// Call tarefa.CriarTarefa
		// CriarTarefa(description string, dueDateStr string, priority int, tags []string) (string, error)
		id, err := tarefa.CriarTarefa(finalDescription, data, priority, tags)
		if err != nil {
			log.Printf("Erro ao criar lembrete (tarefa): %v", err)
			fmt.Println("Falha ao adicionar o lembrete. Verifique os logs para mais detalhes.")
			return
		}
		fmt.Printf("Lembrete (ID: %s) adicionado com sucesso: '%s'\n", id, finalDescription)
		if data != "" {
			fmt.Printf("Data: %s\n", data)
		}
	},
}

var relembrarListarCmd = &cobra.Command{
	Use:   "listar",
	Short: "Lista todos os lembretes pendentes.",
	Long:  `Exibe uma lista de todos os lembretes (tarefas com a tag 'lembrete') que estão pendentes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ListarTarefas(status string, priority int, dueDate string, tag string, sortBy string, sortOrder string) ([]models.Task, error)
		tasks, err := tarefa.ListarTarefas(string(models.TaskStatusPending), 0, "", "lembrete", "DueDate", "asc")
		if err != nil {
			log.Printf("Erro ao listar lembretes (tarefas): %v", err)
			fmt.Println("Falha ao carregar lembretes. Verifique os logs para mais detalhes.")
			return
		}

		fmt.Println("==================================================")
		fmt.Println("                  LEMBRETES PENDENTES")
		fmt.Println("==================================================")
		if len(tasks) == 0 {
			fmt.Println("Nenhum lembrete pendente encontrado.")
			return
		}

		fmt.Printf("%-36s %-30s %-12s %-8s\n", "ID", "LEMBRETE (DESCRIÇÃO)", "DATA", "HORA")
		fmt.Printf("%-36s %-30s %-12s %-8s\n", strings.Repeat("-", 36), strings.Repeat("-", 30), strings.Repeat("-", 12), strings.Repeat("-", 8))

		for _, task := range tasks {
			dueDateStr := ""
			if !task.DueDate.IsZero() {
				dueDateStr = task.DueDate.Format("02/01/2006")
			}
			// Hora will be part of description if added, otherwise blank.
			// For simplicity, not trying to parse it out here from task.Description
			horaStr := ""

			// Truncate description if too long for display
			displayDescription := task.Description
			if len(displayDescription) > 28 {
				displayDescription = displayDescription[:27] + "..."
			}

			fmt.Printf("%-36s %-30s %-12s %-8s\n", task.ID, displayDescription, dueDateStr, horaStr)
		}
	},
}

func init() {
	RelembrarCmd.AddCommand(relembrarAdicionarCmd)
	RelembrarCmd.AddCommand(relembrarListarCmd)
}
