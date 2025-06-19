package squad4

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var RelembrarCmd = &cobra.Command{
	Use:   "relembrar",
	Short: "Gerencia lembretes.",
	Long:  `Permite adicionar e listar lembretes para ajudar o professor a se organizar.`,
}

var relembrarAdicionarCmd = &cobra.Command{
	Use:   "adicionar \"<lembrete>\" <data> [hora]",
	Short: "Adiciona um novo lembrete.",
	Long: `Adiciona um novo lembrete com uma descrição, data e, opcionalmente, uma hora.
A descrição do lembrete deve estar entre aspas.
Formato da data: DD/MM/YYYY ou palavras-chave como "hoje", "amanha".
Formato da hora (opcional): HH:MM.`,
	Example: `relembrar adicionar "Comprar canetas vermelhas" amanha 10:00
relembrar adicionar "Buscar provas na gráfica" 23/07/2025`,
	Args: cobra.MinimumNArgs(2), // lembrete e data são obrigatórios
	Run: func(cmd *cobra.Command, args []string) {
		lembrete := args[0]
		data := args[1]
		hora := ""
		if len(args) > 2 {
			hora = args[2]
		}

		// Mocked interaction: Just print a confirmation
		if hora != "" {
			fmt.Printf("Lembrete '%s' adicionado para %s %s.\n", lembrete, data, hora)
		} else {
			fmt.Printf("Lembrete '%s' adicionado para %s.\n", lembrete, data)
		}
		// In a real application, this would be saved to a database or other storage.
	},
}

var relembrarListarCmd = &cobra.Command{
	Use:   "listar",
	Short: "Lista todos os lembretes.",
	Long:  `Exibe uma lista de todos os lembretes pendentes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Mocked data based on mockups.txt
		type LembreteMock struct {
			ID        int
			Descricao string
			Data      string
			Hora      string
		}
		lembretes := []LembreteMock{
			{ID: 1, Descricao: "Comprar canetas", Data: "21/06/2025", Hora: "10:00"},
			{ID: 2, Descricao: "Ligar para gráfica", Data: "24/06/2025", Hora: "15:30"},
			{ID: 3, Descricao: "Confirmar reunião", Data: "HOJE", Hora: "17:00"},
			// Adding one more from Phase 2 requirements example
			{ID: 4, Descricao: "Preparar aula de Biologia", Data: "Amanhã", Hora: "09:00"},
		}

		fmt.Println("==================================================")
		fmt.Println("                            LEMBRETES")
		fmt.Println("==================================================")
		fmt.Printf("%-3s %-30s %-12s %-8s\n", "ID", "LEMBRETE", "DATA", "HORA")
		fmt.Printf("--- %-30s %-12s %-8s\n", strings.Repeat("-", 30), strings.Repeat("-", 12), strings.Repeat("-", 8))

		for _, l := range lembretes {
			fmt.Printf("%-3d %-30s %-12s %-8s\n", l.ID, l.Descricao, l.Data, l.Hora)
		}
		// In a real application, this data would be fetched from storage.
	},
}

func init() {
	RelembrarCmd.AddCommand(relembrarAdicionarCmd)
	RelembrarCmd.AddCommand(relembrarListarCmd)
}
