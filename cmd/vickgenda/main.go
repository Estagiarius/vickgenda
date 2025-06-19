package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"vickgenda-cli/internal/ids"
	"vickgenda-cli/internal/squad4" // Import for DashboardCmd
	"vickgenda-cli/internal/tui"
)

// model represents the main state of the Bubble Tea TUI application.
// It encapsulates all UI elements and data needed for rendering and interaction.
type model struct {
	// textInput is a placeholder for a simple text input field in the TUI.
	textInput string
	// cursor is the cursor position for the textInput field.
	cursor int
	// statusBar holds the model for the persistent status bar component.
	statusBar tui.StatusBarModel
	// modeOutput is used to store output from CLI commands if vickgenda
	// is invoked with a subcommand before potentially entering TUI mode,
	// or if a command needs to print output without starting the TUI.
	modeOutput string
	// isTuiMode indicates whether the application should run in interactive TUI mode.
	// This is true by default if no subcommand is executed.
	isTuiMode bool
}

// initialModel creates and returns the initial state of the main TUI model.
// It sets up the status bar and defaults to TUI mode.
func initialModel() model {
	return model{
		statusBar: tui.NewStatusBarModel(), // Initialize the status bar component
		isTuiMode: true,                    // Default to TUI mode
	}
}

// Init is the first command that will be executed when the Bubble Tea program starts.
// It initializes any components that require it, like the status bar.
func (m model) Init() tea.Cmd {
	if m.isTuiMode {
		// Initialize the status bar component.
		// This could return commands for the status bar to run on startup (e.g., fetching initial data).
		return m.statusBar.Init()
	}
	// If not in TUI mode, no initial commands are necessary from the main model.
	return nil
}

// Update handles incoming messages (events) and updates the model's state accordingly.
// Messages can be key presses, window size changes, or custom commands.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If not in TUI mode, this model should generally not be receiving updates.
	// This check prevents unexpected behavior if Update is somehow called.
	if !m.isTuiMode {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		default:
			if msg.Type == tea.KeyRunes { // Handle printable characters
				m.textInput = m.textInput[:m.cursor] + string(msg.Runes) + m.textInput[m.cursor:]
				m.cursor += len(msg.Runes)
			}
		}
	}

	// Pass messages to status bar model for potential updates
	newStatusBar, statusBarCmd := m.statusBar.Update(msg)
	m.statusBar = newStatusBar // Ensure the model's status bar is updated
	if statusBarCmd != nil {
		cmds = append(cmds, statusBarCmd)
	}

	return m, tea.Batch(cmds...)
}

// View generates the string representation of the TUI display.
// It is called by Bubble Tea when the display needs to be updated.
func (m model) View() string {
	// If the application is not in TUI mode (e.g., a subcommand was run that printed output),
	// this returns any output captured in modeOutput.
	// This scenario is less common if subcommands print directly and exit.
	if !m.isTuiMode {
		return m.modeOutput
	}

	// Main content of the TUI. This is a placeholder.
	// In a real application, this would be composed of various UI components.
	mainContent := fmt.Sprintf(
		"Hello, Bubble Tea! (Press Ctrl+C or Esc to quit)\n\n%s",
		m.textInput, // Display the current text input
	)

	// If modeOutput has any content (e.g., from a previous command before TUI fully started,
	// or some initial status message), prepend it to the main content.
	// This is a simple way to show pre-TUI information.
	if m.modeOutput != "" {
		mainContent = m.modeOutput + "\n\n---\n" + mainContent
	}

	// Get the view from the status bar component.
	statusBarView := m.statusBar.View()

	// TODO: Adjust layout dynamically based on terminal height.
	// A more robust solution would involve handling tea.WindowSizeMsg to get terminal dimensions
	// and then use lipgloss.Height() and other properties to arrange content dynamically,
	// ensuring the status bar is always at the bottom and main content uses available space.
	// For now, lipgloss.JoinVertical places the status bar directly below the main content.
	return lipgloss.JoinVertical(lipgloss.Left, mainContent, statusBarView)
}

// Cobra Command definitions

// rootCmd is the base command when vickgenda is called without any subcommands.
// It's configured to launch the TUI by default if no subcommand is specified.
var rootCmd = &cobra.Command{
	Use:   "vickgenda",
	Short: "Vickgenda is a CLI tool and TUI for managing your agenda.",
	Long: `Vickgenda can be run as a TUI application (default if no subcommand is given)
or with subcommands for specific actions.
It also supports shell completion generation. For example:
  vickgenda completion bash > /etc/bash_completion.d/vickgenda`,
	// CompletionOptions.DisableDefaultCmd = false enables Cobra's built-in 'completion' command.
	// This allows users to generate shell completion scripts (e.g., for bash, zsh).
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
	// RunE is executed if no subcommand is specified or if the subcommand
	// doesn't have its own RunE/Run.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default action: if no arguments (subcommands) are provided, run the TUI.
		if len(args) == 0 {
			tuiModel := initialModel() // Create the initial TUI model.
			p := tea.NewProgram(tuiModel) // Create a new Bubble Tea program.

			// p.Run() starts the Bubble Tea event loop and renders the TUI.
			// It blocks until the TUI exits (e.g., via tea.Quit).
			if _, err := p.Run(); err != nil {
				// If p.Run() returns an error, report it.
				return fmt.Errorf("error running TUI: %w", err)
			}
			return nil // TUI exited cleanly.
		}
		// If arguments are present but not matched by a known subcommand,
		// Cobra will typically show a "command not found" error or help.
		return nil
	},
}

// resolveCmd is a Cobra command for resolving contextual IDs using the ids package.
// This is a placeholder/example command demonstrating the ID resolution functionality.
var resolveCmd = &cobra.Command{
	Use:   "resolve [context_type] [token]",
	Short: "Resolves a contextual ID to its database ID (placeholder).",
	Long:  "Example usage: vickgenda resolve task t1\n\n" + "Known placeholder examples:\n" + fmt.Sprintf("  %s\n", ids.GetPlaceholderContextualIDExamples()),
	Args:  cobra.ExactArgs(2), // Expects exactly two arguments: context_type and token.
	RunE: func(cmd *cobra.Command, args []string) error {
		contextType := args[0] // First argument is context_type.
		token := args[1]       // Second argument is the token to resolve.

		// Call the Resolve function from the ids package.
		resolvedID, err := ids.Resolve(contextType, token)
		if err != nil {
			// Error will be printed to Stderr by Cobra's default error handling.
			return fmt.Errorf("resolution error: %w", err)
		}

		// Print to Stdout for successful resolution.
		// Using cmd.OutOrStdout() is good practice for Cobra commands.
		cmd.Printf("Resolved ID:\n  Original Token: %s\n  Database ID:    %s\n  Context Type:   %s\n",
			resolvedID.OriginalToken, resolvedID.DatabaseID, resolvedID.ContextType)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Execute runs the root command, which in turn parses arguments and flags
// and runs the appropriate subcommand or the root command's RunE function.
func Execute() error {
	return rootCmd.Execute()
}

// GetMainRootCmd returns the main root command for the application.
// This allows other packages (like cmd) to access and add commands to it.
func GetMainRootCmd() *cobra.Command {
	return rootCmd
}

func main() {
	// Add subcommands to the root command.
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(squad4.DashboardCmd) // Add DashboardCmd

	// Cobra's Execute() handles command parsing, execution, and error printing.
	//  - If a registered subcommand (like "resolve") is called, its RunE/Run function is executed.
	//  - If no subcommand is called (or "vickgenda" is called by itself), rootCmd's RunE is executed,
	//    which in this application, launches the Bubble Tea TUI.
	//  - If "vickgenda completion [shell]" is called, the built-in completion command runs.
	if err := Execute(); err != nil {
		// os.Exit(1) ensures that the process exits with an error code if Execute()
		// returns an error. Cobra itself usually prints the error message to stderr
		// before Execute() returns.
		os.Exit(1)
	}
}
