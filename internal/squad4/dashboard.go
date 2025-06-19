package squad4

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Exibe o painel principal com informações resumidas.",
	Long:  `Exibe o painel principal contendo um resumo de eventos do dia, tarefas pendentes e outras informações úteis para o professor.`,
	Run: func(cmd *cobra.Command, args []string) {
		displayDashboard()
	},
}

func displayDashboard() {
	// Mocked data
	userName := "Prof. Exemplo"
	today := time.Now().Format("02/01/2006") // DD/MM/YYYY format

	// Mocked events
	events := []string{
		"[08:00] Aula: Matemática - Turma 7B - Frações",
		"[10:00] Reunião: Coordenação Pedagógica",
		"[14:00] Corrigir: Provas - História - Turma 8A",
	}

	// Mocked pending tasks
	pendingTasks := []string{
		"[ ] Preparar aula de Ciências (Amanhã)",
		"[ ] Lançar notas Bimestre 1 - Turma 7B (Até 25/06)",
		"[ ] Organizar material para feira de ciências",
	}

	// Mocked focus quote
	focusQuote := "\"Concentre-se em uma tarefa de cada vez.\""

	fmt.Println("==================================================")
	fmt.Println("                PAINEL PRINCIPAL")
	fmt.Println("==================================================")
	fmt.Printf("Bom dia, Professor(a) %s!\n\n", userName)

	fmt.Printf("HOJE (%s):\n", today)
	fmt.Println("--------------------------------------------------")
	for _, event := range events {
		fmt.Println(event)
	}
	fmt.Println()

	fmt.Println("TAREFAS PENDENTES:")
	fmt.Println("--------------------------------------------------")
	for _, task := range pendingTasks {
		fmt.Println(task)
	}
	fmt.Println()

	fmt.Println("FOCO DO DIA:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(focusQuote)
	fmt.Println()

	fmt.Println("--------------------------------------------------")
	fmt.Println("Use 'ajuda' para ver todos os comandos.")
}

func init() {
	// This function is run when the package is initialized.
	// It's a good place to add subcommands or flags if needed in the future.
}
