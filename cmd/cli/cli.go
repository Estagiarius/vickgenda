package cli

import (
	"fmt"
	// "os" // Removed as unused

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	// "vickgenda-cli/internal/db"     // Removed as unused (InitDB moved to main.go)
	"vickgenda-cli/internal/ids"    // For resolveCmd
	"vickgenda-cli/internal/squad4" // For DashboardCmd
	// "vickgenda-cli/internal/tui" // Will be needed if TUI logic is separate

	"vickgenda-cli/cmd/bancoq" // Added for bancoq.BancoqCmd
	"vickgenda-cli/cmd/prova"  // Added for prova.ProvaCmd
	"vickgenda-cli/cmd"        // Added for the new commands like TarefaCmd, AgendaCmd etc.
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
	// squad4.DashboardCmd is now added via InitSquad4Commands in SetupRootCmd

	// Add Squad 2 commands (assuming they are global vars in main cmd package, like agendaCmd, tarefaCmd, rotinaCmd)
	// These will be initialized from their respective files (e.g., cmd/agenda.go)
	// Example: rootCmd.AddCommand(cmd.AgendaCmd) - This needs actual variable names

	// Add Squad 3 commands
	// Example: rootCmd.AddCommand(cmd.AulaCmd)
	// Example: rootCmd.AddCommand(cmd.NotasCmd)

	// Add Squad 4 commands (relembrar, foco, relatorio)
	// Example: rootCmd.AddCommand(cmd.RelembrarCmd)
	// Example: rootCmd.AddCommand(cmd.FocoCmd)
	// Example: rootCmd.AddCommand(cmd.RelatorioCmd)
	// Note: squad4.DashboardCmd is added in SetupRootCmd

	// Add Squad 5 commands
	rootCmd.AddCommand(bancoq.BancoqCmd) // From imported package
	rootCmd.AddCommand(prova.ProvaCmd)   // From imported package
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
	// The db.InitDB("") call was moved to main.go's main() function.
	// if err := db.InitDB(""); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Erro ao inicializar o banco de dados antes de Execute(): %v\n", err)
	// 	os.Exit(1) // Or return the error to be handled by main
	// }
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
		textInput: "Vickgenda TUI - Digite algo ou Ctrl+C para sair.", // Initial text
		// statusBar: tui.NewStatusBarModel(),
		isTuiMode: true,
		cursor:    0, // Initialize cursor
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyRunes, tea.KeySpace: // Handle printable characters and space
			m.textInput += string(msg.Runes) // Append typed character
			// A more sophisticated input would handle cursor position, backspace, etc.
		// Add other key handling as needed (e.g., tea.KeyBackspace, tea.KeyEnter)
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.isTuiMode {
		return m.modeOutput
	}
	// Display current textInput. A real TUI would have a more structured view.
	return fmt.Sprintf("%s\n\n%s\n\n(Ctrl+C ou Esc para sair)",
		"Bem-vindo ao Vickgenda TUI!",
		m.textInput)
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
	// resolveCmd is already added in init()
	// squad4.DashboardCmd is now added here along with other Squad 4 commands
	squad4.InitSquad4Commands(rootCmd) // This adds Dashboard, Relembrar, Foco, Relatorio

	// Commands created by cobra-cli add are typically in the 'cmd' package (e.g. cmd.TarefaCmd)
	// However, they are not directly accessible here without being exported or a reference passed.
	// The current structure adds them to a 'rootCmd' within their own files, which is problematic.
	// For now, we will assume these commands need to be explicitly added here.
	// This will likely require modification of the generated command files.

	// Placeholder for where new commands would be added if they were accessible:
	rootCmd.AddCommand(cmd.TarefaCmd)
	rootCmd.AddCommand(cmd.AgendaCmd)
	rootCmd.AddCommand(cmd.RotinaCmd)
	rootCmd.AddCommand(cmd.AulaCmd)
	rootCmd.AddCommand(cmd.NotasCmd)
	// DashboardCmd, RelembrarCmd, FocoCmd, RelatorioCmd are added via squad4.InitSquad4Commands

	// Squad 5 commands (bancoq.BancoqCmd, prova.ProvaCmd) are added in init()
}
