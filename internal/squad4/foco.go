package squad4

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings" // Added import for strings package
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var FocoCmd = &cobra.Command{
	Use:   "foco",
	Short: "Gerencia o modo de foco para trabalho concentrado.",
	Long:  `Ajuda o professor a manter o foco em uma tarefa específica por um período determinado.`,
}

var focoIniciarCmd = &cobra.Command{
	Use:   "iniciar <duração_minutos> \"<tarefa>\"",
	Short: "Inicia uma nova sessão de foco.",
	Long: `Inicia um temporizador para uma sessão de foco com uma duração especificada em minutos e uma descrição da tarefa.
A descrição da tarefa deve estar entre aspas.`,
	Example: `foco iniciar 25 "Corrigir provas Turma A"
foco iniciar 45 "Planejar aula de História Moderna"`,
	Args: cobra.ExactArgs(2), // duração e tarefa
	Run: func(cmd *cobra.Command, args []string) {
		duracaoMinutosStr := args[0]
		tarefa := args[1]

		duracaoMinutos, err := strconv.Atoi(duracaoMinutosStr)
		if err != nil {
			fmt.Println("Erro: A duração deve ser um número inteiro de minutos.")
			return
		}

		if duracaoMinutos <= 0 {
			fmt.Println("Erro: A duração deve ser maior que zero minutos.")
			return
		}

		iniciarSessaoFoco(duracaoMinutos, tarefa)
	},
}

func iniciarSessaoFoco(duracaoMinutos int, tarefa string) {
	fmt.Println("==================================================")
	fmt.Println("                            MODO FOCO")
	fmt.Println("==================================================")
	// This initial display matches the mockup for "Screen during focus mode"
	// The timer will then update the "TEMPO RESTANTE" line.

	duracaoTotal := time.Duration(duracaoMinutos) * time.Minute
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	endTime := startTime.Add(duracaoTotal)

	// Setup signal handling for CTRL+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initial display before ticker starts updating
	fmt.Printf("\n                TEMPO RESTANTE: %02d:00\n", duracaoMinutos) // Show full minutes initially
	fmt.Printf("\n                TAREFA: %s\n", tarefa)
	fmt.Println("\n--------------------------------------------------")
	fmt.Println("Mantenha o foco! Pressione CTRL+C para sair.")


	for {
		select {
		case <-ticker.C:
			tempoRestante := time.Until(endTime)
			if tempoRestante <= 0 {
				// Clear the "TEMPO RESTANTE" line before printing completion message
				fmt.Printf("\r%s\r", strings.Repeat(" ", 60)) // Clear line with spaces
				fmt.Println("==================================================")
				fmt.Println("                            MODO FOCO") // Repeated from mockup for consistency
				fmt.Println("==================================================")
				fmt.Println("\n                SESSÃO CONCLUÍDA!")
				fmt.Printf("\n                TAREFA: %s", tarefa)
				fmt.Printf("\n                DURAÇÃO: %d minutos\n", duracaoMinutos)
				fmt.Println("\n--------------------------------------------------")
				fmt.Print("Bom trabalho! Deseja iniciar outra sessão? (s/N) ")
				// For now, we just print and exit. A real app might handle input.
				fmt.Println()
				return
			}
			minutos := int(tempoRestante.Minutes())
			segundos := int(tempoRestante.Seconds()) % 60
			// Use carriage return to overwrite the "TEMPO RESTANTE" line
			fmt.Printf("\r                TEMPO RESTANTE: %02d:%02d", minutos, segundos)
		case <-sigChan:
			// Clear the "TEMPO RESTANTE" line before printing interruption message
			fmt.Printf("\r%s\r", strings.Repeat(" ", 60))
			fmt.Println("\n\nSessão de foco interrompida pelo usuário.")
			fmt.Println("--------------------------------------------------")
			return
		}
	}
}

func init() {
	FocoCmd.AddCommand(focoIniciarCmd)
}
