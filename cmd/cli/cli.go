package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"vickgenda-cli/internal/db"     // For InitDB in root.go
	"vickgenda-cli/internal/ids"    // For resolveCmd
	"vickgenda-cli/internal/squad4" // For DashboardCmd
	// "vickgenda-cli/internal/tui" // Will be needed if TUI logic is separate
)

// rootCmd will be defined here
var rootCmd = &cobra.Command{
	Use:   "vickgenda",
	Short: "Vickgenda CLI - Sua agenda inteligente de linha de comando",
	Long: `Vickgenda CLI é uma ferramenta de linha de comando e TUI (Interface de Usuário de Texto)
projetada para auxiliar no gerenciamento de sua agenda, tarefas, notas e mais, diretamente do terminal.
Por padrão, se nenhum subcomando for fornecido, a interface TUI interativa será iniciada.
Também suporta a geração de scripts de autocompletar para o shell. Por exemplo:
  vickgenda completion bash > /etc/bash_completion.d/vickgenda`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// TUI launching logic will go here.
			// For now, a placeholder:
			fmt.Println("TUI would start here.")
			// tuiModel := initialModel()
			// p := tea.NewProgram(tuiModel)
			// if _, err := p.Run(); err != nil {
			// 	return fmt.Errorf("erro ao executar a TUI: %w", err)
			// }
			return nil
		}
		return nil
	},
}

// resolveCmd definition (copied from cmd/vickgenda/main.go)
var resolveCmd = &cobra.Command{
	Use:   "resolve [tipo_contexto] [token]",
	Short: "Resolve um ID contextual para seu ID de banco de dados (placeholder).",
	Long:  "Exemplo de uso: vickgenda resolve tarefa t1\n\n" + "Exemplos de placeholders conhecidos:\n" + fmt.Sprintf("  %s\n", ids.GetPlaceholderContextualIDExamples()),
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		contextType := args[0]
		token := args[1]

		resolvedID, err := ids.Resolve(contextType, token)
		if err != nil {
			return fmt.Errorf("erro na resolução: %w", err)
		}
		cmd.Printf("ID Resolvido:\n  Token Original: %s\n  ID do Banco de Dados: %s\n  Tipo de Contexto:   %s\n",
			resolvedID.OriginalToken, resolvedID.DatabaseID, resolvedID.ContextType)
		return nil
	},
}


func init() {
	// This InitDB was in cmd/root.go's init. It should be called early.
	// Consider calling it from main() in cmd/vickgenda/main.go before Execute.
	// For now, keeping it here to ensure it's part of the CLI setup.
	// However, init() functions are generally discouraged for explicit setup.
	// if err := db.InitDB(""); err != nil {
	// 	 fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados em cmd/cli/cli.go: %v\n", err)
	// 	 os.Exit(1)
	// }

	// Add commands that were previously in cmd/vickgenda/main.go's main()
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(squad4.DashboardCmd)
}

// GetRootCmd returns the main root command for the application.
// Renamed from GetMainRootCmd for brevity within this package context.
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	// The InitDB call from cmd/root.go's init() needs to happen before Execute.
	// It's better to do this explicitly in main.go's main().
	// For now, I'll add it here, but this might need revisiting.
	if err := db.InitDB(""); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados antes de Execute(): %v\n", err)
		os.Exit(1) // Or return the error to be handled by main
	}
	return rootCmd.Execute()
}

// TUI related structures and functions (to be moved from cmd/vickgenda/main.go)
// This is a simplified version for now. The full TUI logic is more complex.

// model represents the main state of the TUI application.
type model struct {
	textInput string
	cursor    int
	// statusBar tui.StatusBarModel // Assuming tui.StatusBarModel will be defined or imported
	modeOutput string
	isTuiMode  bool
}

// initialModel creates and returns the initial state of the main TUI model.
func initialModel() model {
	return model{
		// statusBar: tui.NewStatusBarModel(),
		isTuiMode: true,
	}
}

func (m model) Init() tea.Cmd {
	if m.isTuiMode {
		// return m.statusBar.Init()
		return nil // Placeholder
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.isTuiMode {
		return m, nil
	}
	// Simplified update logic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.isTuiMode {
		return m.modeOutput
	}
	return fmt.Sprintf("Olá, TUI (Ctrl+C para sair)\n%s", m.textInput) // Simplified view
}

// Update RunE in rootCmd to use the local TUI model and functions
func NewRootCmdRunner() func(cmd *cobra.Command, args []string) error {
    return func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            tuiModel := initialModel()
            p := tea.NewProgram(tuiModel)
            if _, err := p.Run(); err != nil {
                return fmt.Errorf("erro ao executar a TUI: %w", err)
            }
            return nil
        }
        return nil
    }
}

// Re-assign RunE for rootCmd using the new runner func
// This needs to be done after rootCmd is declared.
// A good place is in an init function or just before returning GetRootCmd or executing.
// For now, I'll create a separate function to setup rootCmd completely.

func SetupRootCmd() {
	rootCmd.RunE = NewRootCmdRunner()
	// Add other commands that should be part of the main CLI structure
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(squad4.DashboardCmd)
}
