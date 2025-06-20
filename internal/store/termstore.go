package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"vickgenda-cli/internal/models"
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
	stmt, err := s.DB.Prepare(`
		CREATE TABLE IF NOT EXISTS terms (
			id TEXT PRIMARY KEY,
			name TEXT,
			start_date DATETIME,
			end_date DATETIME,
			year INTEGER
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare create table statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute create table statement: %w", err)
	}
	return nil
}

func (s *SQLiteTermStore) SaveTerm(term models.Term) (models.Term, error) {
	if term.ID == "" {
		term.ID = uuid.NewString()
	}
	year := term.StartDate.Year()

	// Overlap Validation
	rows, err := s.DB.Query("SELECT id, start_date, end_date FROM terms WHERE year = ? AND id != ?", year, term.ID)
	if err != nil {
		return models.Term{}, fmt.Errorf("failed to query existing terms for overlap check: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var existingID string
		var existingStartDate, existingEndDate time.Time
		if err := rows.Scan(&existingID, &existingStartDate, &existingEndDate); err != nil {
			return models.Term{}, fmt.Errorf("failed to scan existing term for overlap check: %w", err)
		}

		// Check for overlap: (new.Start <= existing.End) AND (new.End >= existing.Start)
		if (term.StartDate.Equal(existingEndDate) || term.StartDate.Before(existingEndDate)) &&
			(term.EndDate.Equal(existingStartDate) || term.EndDate.After(existingStartDate)) {
			// Retrieving the name of the overlapping term for a more informative error message
			var existingName string
			err := s.DB.QueryRow("SELECT name FROM terms WHERE id = ?", existingID).Scan(&existingName)
			if err != nil {
				// If we can't get the name, use the ID in the error
				return models.Term{}, fmt.Errorf("term dates overlap with existing term ID '%s'", existingID)
			}
			return models.Term{}, fmt.Errorf("term '%s' overlaps with existing term '%s'", term.Name, existingName)
		}
	}
	if err = rows.Err(); err != nil {
		return models.Term{}, fmt.Errorf("error during iteration of existing terms for overlap check: %w", err)
	}

	stmt, err := s.DB.Prepare("INSERT OR REPLACE INTO terms (id, name, start_date, end_date, year) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return models.Term{}, fmt.Errorf("failed to prepare save term statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(term.ID, term.Name, term.StartDate, term.EndDate, year)
	if err != nil {
		return models.Term{}, fmt.Errorf("failed to execute save term statement: %w", err)
	}

	return term, nil
}

func (s *SQLiteTermStore) GetTermByID(id string) (models.Term, error) {
	var term models.Term
	err := s.DB.QueryRow("SELECT id, name, start_date, end_date FROM terms WHERE id = ?", id).Scan(&term.ID, &term.Name, &term.StartDate, &term.EndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Term{}, fmt.Errorf("term with ID '%s' not found: %w", id, err)
		}
		return models.Term{}, fmt.Errorf("failed to get term by ID '%s': %w", id, err)
	}
	return term, nil
}

func (s *SQLiteTermStore) ListTermsByYear(year int) ([]models.Term, error) {
	rows, err := s.DB.Query("SELECT id, name, start_date, end_date FROM terms WHERE year = ? ORDER BY start_date ASC", year)
	if err != nil {
		return nil, fmt.Errorf("failed to query terms by year %d: %w", year, err)
	}
	defer rows.Close()

	var terms []models.Term
	for rows.Next() {
		var term models.Term
		if err := rows.Scan(&term.ID, &term.Name, &term.StartDate, &term.EndDate); err != nil {
			return nil, fmt.Errorf("failed to scan term during ListTermsByYear for year %d: %w", year, err)
		}
		terms = append(terms, term)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of terms for year %d: %w", year, err)
	}

	return terms, nil
}
