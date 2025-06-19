package store

import (
	"database/sql" // Import for eventual SQLite use
	"fmt"
	"time"

	"vickgenda/internal/models"
)

// AulaStore defines the interface for lesson persistence.
// This would eventually be implemented by an SQLite-backed store.
type AulaStore interface {
	Init() error // Initialize table if not exists, etc.
	SaveLesson(lesson models.Lesson) (models.Lesson, error)
	GetLessonByID(id string) (models.Lesson, error)
	ListLessons(disciplina, turma, periodo, mes, ano string) ([]models.Lesson, error)
	UpdateLessonPlan(id, novoPlano, novasObservacoes string) (models.Lesson, error)
	DeleteLesson(id string) error
}

// MemAulaStore is an in-memory implementation for AulaStore (similar to current aula.go logic)
// This helps in transitioning and testing the interface.
// For this subtask, we are just defining stubs/placeholders for the SQLite version,
// so MemAulaStore would be more fleshed out later if used as a fallback or for tests.
// For now, focus on the SQLite placeholder functions.

// SQLiteAulaStore (Placeholder - actual implementation later)
type SQLiteAulaStore struct {
	DB *sql.DB
}

func NewSQLiteAulaStore(db *sql.DB) AulaStore {
	return &SQLiteAulaStore{DB: db}
}

func (s *SQLiteAulaStore) Init() error {
	// Placeholder: Create 'lessons' table SQL statement
	// CREATE TABLE IF NOT EXISTS lessons (
	//     id TEXT PRIMARY KEY,
	//     subject TEXT,
	//     topic TEXT,
	//     date DATETIME,
	//     class_id TEXT,
	//     plan TEXT,
	//     observations TEXT
	// );
	fmt.Println("Placeholder: SQLiteAulaStore.Init() called - table creation logic would go here.")
	return nil
}

func (s *SQLiteAulaStore) SaveLesson(lesson models.Lesson) (models.Lesson, error) {
	// Placeholder: Insert or Update lesson in SQLite
	fmt.Printf("Placeholder: SQLiteAulaStore.SaveLesson() called for lesson ID %s\n", lesson.ID)
	// Actual logic: generate ID if new, then db.Exec(...)
	return lesson, nil // Return the lesson, potentially with DB-generated ID or timestamps
}

func (s *SQLiteAulaStore) GetLessonByID(id string) (models.Lesson, error) {
	// Placeholder: Select lesson by ID from SQLite
	fmt.Printf("Placeholder: SQLiteAulaStore.GetLessonByID() called for ID %s\n", id)
	// Actual logic: db.QueryRow(...).Scan(...)
	return models.Lesson{ID: id, Topic: "Placeholder Lesson"}, nil // Return dummy data
}

func (s *SQLiteAulaStore) ListLessons(disciplina, turma, periodo, mes, ano string) ([]models.Lesson, error) {
	// Placeholder: Select lessons with filters from SQLite
	fmt.Println("Placeholder: SQLiteAulaStore.ListLessons() called")
	// Actual logic: build query based on filters, db.Query(...), iterate rows
	return []models.Lesson{}, nil
}

func (s *SQLiteAulaStore) UpdateLessonPlan(id, novoPlano, novasObservacoes string) (models.Lesson, error) {
    fmt.Printf("Placeholder: SQLiteAulaStore.UpdateLessonPlan() for ID %s\n", id)
    // Actual logic: db.Exec("UPDATE lessons SET plan=?, observations=? WHERE id=?", novoPlano, novasObservacoes, id)
    // Then SELECT to get the updated lesson
    return models.Lesson{ID: id, Plan: novoPlano, Observations: novasObservacoes, Topic: "Updated Placeholder"}, nil
}

func (s *SQLiteAulaStore) DeleteLesson(id string) error {
    fmt.Printf("Placeholder: SQLiteAulaStore.DeleteLesson() for ID %s\n", id)
    // Actual logic: db.Exec("DELETE FROM lessons WHERE id=?", id)
    return nil
}
