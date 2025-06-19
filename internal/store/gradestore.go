package store

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"vickgenda/internal/db"
	"vickgenda/internal/models"
)

const createGradesTableSQL = `
CREATE TABLE IF NOT EXISTS grades (
    id TEXT PRIMARY KEY,
    student_id TEXT NOT NULL,
    term_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    description TEXT NOT NULL,
    value REAL NOT NULL,
    weight REAL NOT NULL,
    date DATETIME NOT NULL,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (term_id) REFERENCES terms(id) ON DELETE CASCADE
);`

// InitGradesTable ensures the grades table exists.
func InitGradesTable() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("InitGradesTable (GetDB): %w", err)
	}
	_, err = database.Exec(createGradesTableSQL)
	if err != nil {
		return fmt.Errorf("InitGradesTable (Exec): %w", err)
	}
	return nil
}

// CreateGrade saves a new grade to the database.
func CreateGrade(grade models.Grade) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("CreateGrade (GetDB): %w", err)
	}

	stmt, err := database.Prepare("INSERT INTO grades (id, student_id, term_id, subject, description, value, weight, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("CreateGrade (Prepare): %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(grade.ID, grade.StudentID, grade.TermID, grade.Subject, grade.Description, grade.Value, grade.Weight, grade.Date)
	if err != nil {
		return fmt.Errorf("CreateGrade (Exec): %w", err)
	}
	return nil
}

// GetGradeByID retrieves a grade by its ID.
func GetGradeByID(id string) (models.Grade, error) {
	database, err := db.GetDB()
	if err != nil {
		return models.Grade{}, fmt.Errorf("GetGradeByID (GetDB): %w", err)
	}
	row := database.QueryRow("SELECT id, student_id, term_id, subject, description, value, weight, date FROM grades WHERE id = ?", id)
	var grade models.Grade
	err = row.Scan(&grade.ID, &grade.StudentID, &grade.TermID, &grade.Subject, &grade.Description, &grade.Value, &grade.Weight, &grade.Date)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Grade{}, fmt.Errorf("grade with ID '%s' not found", id)
		}
		return models.Grade{}, fmt.Errorf("GetGradeByID (Scan): %w", err)
	}
	return grade, nil
}

// UpdateGrade updates an existing grade in the database.
func UpdateGrade(grade models.Grade) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("UpdateGrade (GetDB): %w", err)
	}

	stmt, err := database.Prepare("UPDATE grades SET student_id = ?, term_id = ?, subject = ?, description = ?, value = ?, weight = ?, date = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("UpdateGrade (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(grade.StudentID, grade.TermID, grade.Subject, grade.Description, grade.Value, grade.Weight, grade.Date, grade.ID)
	if err != nil {
		return fmt.Errorf("UpdateGrade (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateGrade: no grade found with ID '%s'", grade.ID)
	}
	return nil
}

// DeleteGrade removes a grade from the database by its ID.
func DeleteGrade(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("DeleteGrade (GetDB): %w", err)
	}
	stmt, err := database.Prepare("DELETE FROM grades WHERE id = ?")
	if err != nil {
		return fmt.Errorf("DeleteGrade (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("DeleteGrade (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("DeleteGrade: no grade found with ID '%s'", id)
	}
	return nil
}

// ListGrades performs querying of grades based on provided filters.
// Filters: studentID (required), termID (optional), subject (optional).
// Results are ordered by date.
func ListGrades(studentID, termID, subject string) ([]models.Grade, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, fmt.Errorf("ListGrades (GetDB): %w", err)
	}

	var queryArgs []interface{}
	var conditions []string

	if studentID == "" {
		return nil, fmt.Errorf("studentID is required to list grades")
	}
	conditions = append(conditions, "student_id = ?")
	queryArgs = append(queryArgs, studentID)

	if termID != "" {
		conditions = append(conditions, "term_id = ?")
		queryArgs = append(queryArgs, termID)
	}
	if subject != "" {
		conditions = append(conditions, "subject = ?") // Case-sensitive, use COLLATE NOCASE for case-insensitivity if needed
		queryArgs = append(queryArgs, subject)
	}

	query := "SELECT id, student_id, term_id, subject, description, value, weight, date FROM grades"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY date ASC"

	rows, err := database.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("ListGrades (Query - '%s'): %w", query, err)
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var grade models.Grade
		err := rows.Scan(&grade.ID, &grade.StudentID, &grade.TermID, &grade.Subject, &grade.Description, &grade.Value, &grade.Weight, &grade.Date)
		if err != nil {
			log.Printf("ERROR: ListGrades (Scan): %v. Query: %s, Args: %v", err, query, queryArgs)
			// Depending on strictness, might return error or skip row
			continue
		}
		grades = append(grades, grade)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ListGrades (RowsErr): %w", err)
	}
	return grades, nil
}

// ClearGradesTableForTesting is a helper for tests to clear all grades.
func ClearGradesTableForTesting() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("ClearGradesTableForTesting (GetDB): %w", err)
	}
	_, err = database.Exec("DELETE FROM grades")
	if err != nil {
		return fmt.Errorf("ClearGradesTableForTesting (Exec): %w", err)
	}
	return nil
}
