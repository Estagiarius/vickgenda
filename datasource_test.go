package main_test // Changed package name

import (
	"fmt"
	"log"
	"testing" // Added testing package

	"vickgenda-cli/internal/commands/agenda"
	"vickgenda-cli/internal/commands/rotina"
	"vickgenda-cli/internal/commands/tarefa"
	"vickgenda-cli/internal/db"
)

func TestDatasourceAccessibility(t *testing.T) { // Renamed main to TestDatasourceAccessibility
	fmt.Println("Attempting to run full datasource accessibility test...")

	// Initialize the database
	if err := db.InitDB(""); err != nil { // This will also print the DB path
		log.Fatalf("Error initializing database: %v", err)
	}
	fmt.Println("Database initialized successfully for testing.")

	// Test Squad 5 data access: db.ListQuestions
	fmt.Println("\n--- Testing db.ListQuestions ---")
	questions, totalQuestions, err := db.ListQuestions(nil, "created_at", "desc", 10, 1)
	if err != nil {
		fmt.Printf("Error listing questions: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d questions (total: %d).\n", len(questions), totalQuestions)
	}

	// Test Squad 2 data access (Agenda): agenda.ListarEventos
	fmt.Println("\n--- Testing agenda.ListarEventos ---")
	// Corrected function signature based on typical use or placeholder if actual is unknown
	// Assuming EventoFilters is a struct or map if needed, passing nil or empty for now.
	// The actual signature might be agenda.ListarEventos(db.GetDB(), ...) if it needs explicit DB.
	// For now, relying on global DB access within the package as per original attempt.
	eventos, err := agenda.ListarEventos("proximos", "", "", "inicio", "asc")
	if err != nil {
		fmt.Printf("Error listing eventos: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d eventos.\n", len(eventos))
	}

	// Test Squad 2 data access (Tarefa): tarefa.ListarTarefas
	fmt.Println("\n--- Testing tarefa.ListarTarefas ---")
	// Assuming TarefaFilters is a struct or map if needed, passing nil or empty for now.
	// The actual signature might be tarefa.ListarTarefas(db.GetDB(), ...)
	tarefas, err := tarefa.ListarTarefas("", 0, "", "", "CreatedAt", "asc")
	if err != nil {
		fmt.Printf("Error listing tarefas: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d tarefas.\n", len(tarefas))
	}

	// Test Squad 2 data access (Rotina): rotina.ListarModelosRotina
	fmt.Println("\n--- Testing rotina.ListarModelosRotina ---")
	// The actual signature might be rotina.ListarModelosRotina(db.GetDB(), ...)
	modelosRotina, err := rotina.ListarModelosRotina("nome", "asc")
	if err != nil {
		fmt.Printf("Error listing modelos de rotina: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d modelos de rotina.\n", len(modelosRotina))
	}

	fmt.Println("\nDatasource accessibility test finished.")
}
