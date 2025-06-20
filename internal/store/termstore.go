package store

import (
	"database/sql"
	"fmt"
	"time"

	"vickgenda/internal/models"
)

// TermStore defines the interface for term (bimestre) persistence.
type TermStore interface {
	Init() error
	SaveTerm(term models.Term) (models.Term, error)
	GetTermByID(id string) (models.Term, error)
	ListTermsByYear(year int) ([]models.Term, error)
    // Potentially: GetTermByNameAndYear(name string, year int) (models.Term, error)
    // Potentially: CheckOverlap(term models.Term) (bool, error)
}

// SQLiteTermStore (Placeholder)
type SQLiteTermStore struct {
	DB *sql.DB
}

func NewSQLiteTermStore(db *sql.DB) TermStore {
	return &SQLiteTermStore{DB: db}
}

func (s *SQLiteTermStore) Init() error {
	// Placeholder: Create 'terms' table
    // CREATE TABLE IF NOT EXISTS terms (id TEXT PRIMARY KEY, name TEXT, start_date DATETIME, end_date DATETIME, year INTEGER);
	fmt.Println("Placeholder: SQLiteTermStore.Init() called")
	return nil
}

func (s *SQLiteTermStore) SaveTerm(term models.Term) (models.Term, error) {
	fmt.Printf("Placeholder: SQLiteTermStore.SaveTerm() for term ID %s\n", term.ID)
	return term, nil
}

func (s *SQLiteTermStore) GetTermByID(id string) (models.Term, error) {
	fmt.Printf("Placeholder: SQLiteTermStore.GetTermByID() for ID %s\n", id)
	return models.Term{ID: id, Name: "Placeholder Term"}, nil
}

func (s *SQLiteTermStore) ListTermsByYear(year int) ([]models.Term, error) {
	fmt.Printf("Placeholder: SQLiteTermStore.ListTermsByYear() for year %d\n", year)
	return []models.Term{}, nil
}
