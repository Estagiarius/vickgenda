package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"vickgenda/internal/db" // Assuming db.GetDB() provides *sql.DB
	"vickgenda/internal/models"
)

const createTermsTableSQL = `
CREATE TABLE IF NOT EXISTS terms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    UNIQUE(name, strftime('%Y', start_date))
);`

// InitTermsTable ensures the terms table exists.
func InitTermsTable() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("InitTermsTable (GetDB): %w", err)
	}
	_, err = database.Exec(createTermsTableSQL)
	if err != nil {
		return fmt.Errorf("InitTermsTable (Exec): %w", err)
	}
	return nil
}

// CreateTerm saves a new term.
func CreateTerm(term models.Term) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("CreateTerm (GetDB): %w", err)
	}

	stmt, err := database.Prepare("INSERT INTO terms (id, name, start_date, end_date) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("CreateTerm (Prepare): %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(term.ID, term.Name, term.StartDate, term.EndDate)
	if err != nil {
		return fmt.Errorf("CreateTerm (Exec): %w", err)
	}
	return nil
}

// UpdateTerm updates an existing term.
func UpdateTerm(term models.Term) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("UpdateTerm (GetDB): %w", err)
	}

	stmt, err := database.Prepare("UPDATE terms SET name = ?, start_date = ?, end_date = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("UpdateTerm (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(term.Name, term.StartDate, term.EndDate, term.ID)
	if err != nil {
		return fmt.Errorf("UpdateTerm (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateTerm: no term found with ID '%s'", term.ID)
	}
	return nil
}


// GetTermByID retrieves a term by its ID.
func GetTermByID(id string) (models.Term, error) {
	database, err := db.GetDB()
	if err != nil {
		return models.Term{}, fmt.Errorf("GetTermByID (GetDB): %w", err)
	}
	row := database.QueryRow("SELECT id, name, start_date, end_date FROM terms WHERE id = ?", id)
	var term models.Term
	err = row.Scan(&term.ID, &term.Name, &term.StartDate, &term.EndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Term{}, fmt.Errorf("term with ID '%s' not found", id)
		}
		return models.Term{}, fmt.Errorf("GetTermByID (Scan): %w", err)
	}
	return term, nil
}

// ListTermsByYear retrieves all terms that are part of a given year.
func ListTermsByYear(year int) ([]models.Term, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, fmt.Errorf("ListTermsByYear (GetDB): %w", err)
	}
	rows, err := database.Query("SELECT id, name, start_date, end_date FROM terms WHERE strftime('%Y', start_date) = ? ORDER BY start_date", fmt.Sprintf("%04d", year))
	if err != nil {
		return nil, fmt.Errorf("ListTermsByYear (Query): %w", err)
	}
	defer rows.Close()

	var terms []models.Term
	for rows.Next() {
		var term models.Term
		err := rows.Scan(&term.ID, &term.Name, &term.StartDate, &term.EndDate)
		if err != nil {
			log.Printf("ERROR: ListTermsByYear (Scan) for year %d: %v", year, err)
			continue
		}
		terms = append(terms, term)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ListTermsByYear (RowsErr): %w", err)
	}
	return terms, nil
}

// DeleteTerm removes a term from the database by its ID.
func DeleteTerm(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("DeleteTerm (GetDB): %w", err)
	}
	stmt, err := database.Prepare("DELETE FROM terms WHERE id = ?")
	if err != nil {
		return fmt.Errorf("DeleteTerm (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("DeleteTerm (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("DeleteTerm: no term found with ID '%s'", id)
	}
	return nil
}

// ClearTermsTableForTesting is a helper for tests to clear all terms.
func ClearTermsTableForTesting() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("ClearTermsTableForTesting (GetDB): %w", err)
	}
	_, err = database.Exec("DELETE FROM terms")
	if err != nil {
		return fmt.Errorf("ClearTermsTableForTesting (Exec): %w", err)
	}
	return nil
}
