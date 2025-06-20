package prova

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vickgenda-cli/internal/models"
)

// Sample tests for simulation - this list will be modified by the delete operation.
// It needs to be a package-level variable to persist changes across calls in a real scenario,
// but for a single command execution simulation, it's fine.
// For more robust state between CLI calls, an actual DB or file store would be needed.
var sampleGeneratedProvasForDelete = []models.Test{
	{ID: "del123", Title: "Prova de Exatas Antiga", Subject: "Matemática", CreatedAt: time.Now().Add(-72 * time.Hour), QuestionIDs: []string{"q1", "q4"}},
	{ID: "del456", Title: "Teste de Humanas Passado", Subject: "História", CreatedAt: time.Now().Add(-96 * time.Hour), QuestionIDs: []string{"q3"}},
	{ID: "del789", Title: "Avaliação Rápida de Geografia", Subject: "Geografia", CreatedAt: time.Now().Add(-120 * time.Hour), QuestionIDs: []string{"q5"}},
}

// deleteCmd representa o comando para remover uma prova.
var deleteCmd = &cobra.Command{
	Use:   "delete <id_prova>",
	Short: "Remove uma prova",
	Long:  `Remove uma prova do sistema com base no seu ID.`,
	Args:  cobra.ExactArgs(1), // Espera exatamente um argumento: o ID da prova.
	Run: func(cmd *cobra.Command, args []string) {
		provaID := args[0] // Already validated by cobra.ExactArgs(1)
		force, _ := cmd.Flags().GetBool("force")

		fmt.Printf("Executando o comando 'prova delete' para a Prova ID: %s\n", provaID)

		// 1. Encontrar a prova
		foundIndex := -1
		var provaTitle string
		for i, p := range sampleGeneratedProvasForDelete {
			if p.ID == provaID {
				foundIndex = i
				provaTitle = p.Title // Guardar o título para a mensagem de confirmação/sucesso
				break
			}
		}

		if foundIndex == -1 {
			fmt.Printf("Erro: Prova com ID '%s' não encontrada na lista de simulação.\n", provaID)
			return
		}

		// 2. Confirmação (se --force não for usado)
		proceedToDelete := false
		if force {
			fmt.Println("Opção --force utilizada. Removendo a prova diretamente.")
			proceedToDelete = true
		} else {
			fmt.Printf("\nTem certeza que deseja remover a prova '%s' (ID: %s)?\n", provaTitle, provaID)
			fmt.Print("Esta ação não pode ser desfeita. Digite 'sim' para confirmar: ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			if input == "sim" {
				proceedToDelete = true
			} else {
				fmt.Println("Remoção cancelada pelo usuário.")
			}
		}

		// 3. Remover a prova (se confirmado ou forçado)
		if proceedToDelete {
			// Remover o elemento da slice
			sampleGeneratedProvasForDelete = append(sampleGeneratedProvasForDelete[:foundIndex], sampleGeneratedProvasForDelete[foundIndex+1:]...)
			fmt.Printf("\nProva '%s' (ID: %s) removida com sucesso.\n", provaTitle, provaID)

			// Opcional: Mostrar a lista restante para verificar (para fins de depuração/simulação)
			fmt.Println("\nLista de provas restantes (simulação):")
			if len(sampleGeneratedProvasForDelete) == 0 {
				fmt.Println("Nenhuma prova restante.")
			} else {
				for _, p := range sampleGeneratedProvasForDelete {
					fmt.Printf("- ID: %s, Título: %s\n", p.ID, p.Title)
				}
			}
		}
		fmt.Println("\nComando 'prova delete' concluído.")
	},
}

func init() {
	ProvaCmd.AddCommand(deleteCmd)
	// Flags para o comando delete (baseado em docs/specifications/prova_command_spec.md):
	deleteCmd.Flags().BoolP("force", "f", false, "Forçar a remoção da prova sem confirmação (opcional, padrão: false)")
}
