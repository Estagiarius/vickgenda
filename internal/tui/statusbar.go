package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarAppNameStyle = lipgloss.NewStyle().Bold(true)
	statusBarVersionStyle = lipgloss.NewStyle().Faint(true)
	statusBarContextStyle = lipgloss.NewStyle().Italic(true)
	statusBarStyle        = lipgloss.NewStyle().
				Background(lipgloss.Color("235")). // Dark gray background
				Foreground(lipgloss.Color("250")). // Light gray foreground
				PaddingLeft(1).
				PaddingRight(1)
)

// StatusBarModel represents the state and behavior of the application's status bar.
// It's a Bubble Tea model component designed to be embedded in parent models.
type StatusBarModel struct {
	// AppName is the name of the application, displayed in the status bar.
	AppName string
	// Version is the current version of the application.
	Version string
	// CurrentContext provides information about the current view or mode of the application.
	// E.g., "Navegação", "Editando Tarefa".
	CurrentContext string
	// TODO: Add a field for dynamic messages or short-lived alerts,
	// e.g., "Salvando...", "Erro: Falha ao conectar". This would require
	// handling messages or having methods to set/clear this message.
}

// NewStatusBarModel creates and returns a new StatusBarModel initialized with default values.
// These defaults can be overridden by the parent model if needed.
func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{
		AppName:        "Vickgenda",    // Default application name.
		Version:        "0.1.0",        // Default version string.
		CurrentContext: "Navegação", // Default context; "Navegação" implies general browsing/idle state.
	}
}

// Init is called when the StatusBarModel is initialized as part of a Bubble Tea program.
// It can return a command to be executed by the Bubble Tea runtime.
// For the status bar, no initial commands are typically needed.
func (m StatusBarModel) Init() tea.Cmd {
	return nil
}

// Update handles messages sent to the StatusBarModel. This allows the status bar
// to react to application events or changes in state.
//
// For now, the status bar's content is relatively static or updated by direct field
// modification in the parent model. This Update method could be expanded to handle:
//   - tea.WindowSizeMsg: If the status bar needs to adjust its layout dynamically.
//   - Custom messages: To update AppName, Version, CurrentContext, or show temporary alerts.
//     For example, a `SetContextMsg` could update `m.CurrentContext`.
//
// Returns the updated model and any command to be executed.
func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
	// Example of how a custom message could be handled:
	// switch msg := msg.(type) {
	// case tui.StatusMessage: // Assuming tui.StatusMessage is a custom message type
	//   m.CurrentContext = string(msg) // Update context from message
	//   return m, nil                  // No command to return
	// }
	return m, nil // No messages are currently processed, return model and no command.
}

// View renders the StatusBarModel as a string, suitable for display in the TUI.
// It uses lipgloss styles defined at the package level to format the different parts
// of the status bar, ensuring a consistent appearance.
func (m StatusBarModel) View() string {
	// Render each part of the status bar using its defined lipgloss style.
	appName := statusBarAppNameStyle.Render(m.AppName)
	version := statusBarVersionStyle.Render("v" + m.Version)
	// "Contexto" is used for pt-BR UI consistency.
	context := statusBarContextStyle.Render("Contexto: " + m.CurrentContext)

	// Combine the styled parts into a single string.
	statusText := fmt.Sprintf("%s %s | %s", appName, version, context)
	// Apply the overall status bar style (background, foreground, padding) to the combined text.
	return statusBarStyle.Render(statusText)
}
