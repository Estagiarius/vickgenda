package squad4

import (
	"fmt"
	"log"
	"time"

	"vickgenda-cli/internal/commands/agenda"
	"vickgenda-cli/internal/commands/tarefa"
	"vickgenda-cli/internal/models"

	"github.com/spf13/cobra"
)

var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Exibe o painel principal com informações resumidas.",
	Long:  `Exibe o painel principal contendo um resumo de eventos do dia, tarefas pendentes e outras informações úteis para o professor.`,
	Run: func(cmd *cobra.Command, args []string) {
		// It's good practice to ensure DB is initialized before running commands that need it.
		// This might be handled globally in main.go or cmd/root.go's init.
		// If not, db.InitDB("") should be called here or before this command runs.
		// For now, assuming DB is initialized by the time this command runs.
		displayDashboard()
	},
}

func displayDashboard() {
	userName := "Prof. Exemplo" // Static for now
	today := time.Now().Format("02/01/2006")
	focusQuote := "\"Concentre-se em uma tarefa de cada vez.\"" // Static for now

	// --- Fetch Real Events ---
	var eventStrings []string
	realEvents, err := agenda.ListarEventos("dia", "", "", "inicio", "asc")
	if err != nil {
		log.Printf("Error fetching events for dashboard: %v", err)
		eventStrings = append(eventStrings, "Erro ao carregar eventos do dia.")
	} else {
		if len(realEvents) == 0 {
			eventStrings = append(eventStrings, "Nenhum evento para hoje.")
		} else {
			for _, event := range realEvents {
				// Format: [HH:MM] Titulo - Descricao (Local)
				// Assuming event.Inicio is a time.Time object
				// Assuming event.Local is a field in models.Event
				local := ""
				if event.Local != "" {
					local = fmt.Sprintf(" (%s)", event.Local)
				}
				eventStrings = append(eventStrings, fmt.Sprintf("[%s] %s - %s%s",
					event.Inicio.Format("15:04"), event.Titulo, event.Descricao, local))
			}
		}
	}

	// --- Fetch Real Pending Tasks ---
	var taskStrings []string
	// ListarTarefas(status string, priority int, dueDate string, tag string, sortBy string, sortOrder string)
	realTasks, err := tarefa.ListarTarefas("Pendente", 0, "", "", "DueDate", "asc")
	if err != nil {
		log.Printf("Error fetching tasks for dashboard: %v", err)
		taskStrings = append(taskStrings, "Erro ao carregar tarefas pendentes.")
	} else {
		if len(realTasks) == 0 {
			taskStrings = append(taskStrings, "Nenhuma tarefa pendente. Bom trabalho!")
		} else {
			for _, task := range realTasks {
				// Format: [ ] Descricao (Prazo: DD/MM/YYYY)
				dueDateStr := ""
				if !task.DueDate.IsZero() { // Check if DueDate is set
					dueDateStr = fmt.Sprintf(" (Prazo: %s)", task.DueDate.Format("02/01/2006"))
				}
				// Assuming task.Status gives "Pendente", "Em Andamento", "Concluída"
				// For pending, we use "[ ]"
				statusMarker := "[ ]" // Default for pending
				if task.Status == models.TaskStatusCompleted {
					statusMarker = "[x]"
				} else if task.Status == models.TaskStatusInProgress {
					statusMarker = "[/]"
				}

				taskStrings = append(taskStrings, fmt.Sprintf("%s %s%s",
					statusMarker, task.Description, dueDateStr))
			}
		}
	}

	// --- Display Logic ---
	fmt.Println("==================================================")
	fmt.Println("                PAINEL PRINCIPAL")
	fmt.Println("==================================================")
	fmt.Printf("Bom dia, Professor(a) %s!\n\n", userName)

	fmt.Printf("HOJE (%s):\n", today)
	fmt.Println("--------------------------------------------------")
	if len(eventStrings) == 0 {
		fmt.Println("Nenhum evento para exibir.")
	} else {
		for _, eventStr := range eventStrings {
			fmt.Println(eventStr)
		}
	}
	fmt.Println()

	fmt.Println("TAREFAS PENDENTES:")
	fmt.Println("--------------------------------------------------")
	if len(taskStrings) == 0 {
		fmt.Println("Nenhuma tarefa para exibir.")
	} else {
		for _, taskStr := range taskStrings {
			fmt.Println(taskStr)
		}
	}
	fmt.Println()

	fmt.Println("FOCO DO DIA:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(focusQuote)
	fmt.Println()

	fmt.Println("--------------------------------------------------")
	fmt.Println("Use 'vickgenda ajuda' para ver todos os comandos.") // Updated help command
}

func init() {
	// This function is run when the package is initialized.
}
