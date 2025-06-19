package ids

import "fmt"

// ResolvedID represents an ID that has been resolved from a contextual token.
type ResolvedID struct {
	OriginalToken string
	DatabaseID    string // Could be int or string depending on DB
	ContextType   string
}

// Resolve attempts to find a database ID for a given token within a specific context.
// This is placeholder logic.
func Resolve(contextType string, token string) (ResolvedID, error) {
	switch token {
	case "t1":
		return ResolvedID{OriginalToken: token, DatabaseID: "task-db-id-1", ContextType: "task"}, nil
	case "p2":
		return ResolvedID{OriginalToken: token, DatabaseID: "project-db-id-2", ContextType: "project"}, nil
	case "n3":
		return ResolvedID{OriginalToken: token, DatabaseID: "note-db-id-3", ContextType: "note"}, nil
	default:
		return ResolvedID{}, fmt.Errorf("token '%s' not found in context '%s'", token, contextType)
	}
}

// GetPlaceholderContextualIDExamples returns a list of example contextual IDs.
func GetPlaceholderContextualIDExamples() []string {
	return []string{
		"t1 (resolves to a task)",
		"p2 (resolves to a project)",
		"n3 (resolves to a note)",
		"e.g., <command> t1",
	}
}
