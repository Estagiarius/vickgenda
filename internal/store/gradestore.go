package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"vickgenda/internal/models"
)

// GradeStore defines the interface for grade persistence.
type GradeStore interface {
	Init() error
	SaveGrade(grade models.Grade) (models.Grade, error)
	GetGradeByID(id string) (models.Grade, error)
	ListGradesByStudent(studentID, termID, subject string) ([]models.Grade, error)
	UpdateGrade(grade models.Grade) (models.Grade, error) // Or specific update fields
	DeleteGrade(id string) error
}

// SQLiteGradeStore (Placeholder)
type SQLiteGradeStore struct {
	DB *sql.DB
}

func NewSQLiteGradeStore(db *sql.DB) GradeStore {
	return &SQLiteGradeStore{DB: db}
}

func (s *SQLiteGradeStore) Init() error {
	// Enable foreign key support if not enabled by default.
	// For SQLite, this is often done per connection, but some drivers might handle it.
	// _, err := s.DB.Exec("PRAGMA foreign_keys = ON;")
	// if err != nil {
	// 	return fmt.Errorf("failed to enable foreign keys: %w", err)
	// }

	stmt, err := s.DB.Prepare(`
		CREATE TABLE IF NOT EXISTS grades (
			id TEXT PRIMARY KEY,
			student_id TEXT,
			term_id TEXT,
			subject TEXT,
			description TEXT,
			value REAL,
			weight REAL,
			date DATETIME,
			FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
			FOREIGN KEY (term_id) REFERENCES terms(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare create grades table statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute create grades table statement: %w", err)
	}
	return nil
}

func (s *SQLiteGradeStore) SaveGrade(grade models.Grade) (models.Grade, error) {
	if grade.ID == "" {
		grade.ID = uuid.NewString()
	}

	stmt, err := s.DB.Prepare(`
		INSERT OR REPLACE INTO grades
		(id, student_id, term_id, subject, description, value, weight, date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return models.Grade{}, fmt.Errorf("failed to prepare save grade statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(grade.ID, grade.StudentID, grade.TermID, grade.Subject, grade.Description, grade.Value, grade.Weight, grade.Date)
	if err != nil {
		return models.Grade{}, fmt.Errorf("failed to execute save grade statement for grade ID %s: %w", grade.ID, err)
	}
	return grade, nil
}

func (s *SQLiteGradeStore) GetGradeByID(id string) (models.Grade, error) {
	var grade models.Grade
	err := s.DB.QueryRow(
		"SELECT id, student_id, term_id, subject, description, value, weight, date FROM grades WHERE id = ?",
		id,
	).Scan(
		&grade.ID, &grade.StudentID, &grade.TermID, &grade.Subject,
		&grade.Description, &grade.Value, &grade.Weight, &grade.Date,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Grade{}, fmt.Errorf("grade with ID '%s' not found: %w", id, err)
		}
		return models.Grade{}, fmt.Errorf("failed to get grade by ID '%s': %w", id, err)
	}
	return grade, nil
}

func (s *SQLiteGradeStore) ListGradesByStudent(studentID, termID, subject string) ([]models.Grade, error) {
	query := "SELECT id, student_id, term_id, subject, description, value, weight, date FROM grades WHERE student_id = ?"
	args := []interface{}{studentID}

	if termID != "" {
		query += " AND term_id = ?"
		args = append(args, termID)
	}
	if subject != "" {
		query += " AND subject = ?"
		args = append(args, subject)
	}
	query += " ORDER BY date ASC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query grades for student ID '%s': %w", studentID, err)
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var grade models.Grade
		if err := rows.Scan(
			&grade.ID, &grade.StudentID, &grade.TermID, &grade.Subject,
			&grade.Description, &grade.Value, &grade.Weight, &grade.Date,
		); err != nil {
			return nil, fmt.Errorf("failed to scan grade for student ID '%s': %w", studentID, err)
		}
		grades = append(grades, grade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of grades for student ID '%s': %w", studentID, err)
	}

	return grades, nil
}

func (s *SQLiteGradeStore) UpdateGrade(grade models.Grade) (models.Grade, error) {
	// SaveGrade uses "INSERT OR REPLACE", so it handles both creation and update.
	// A check for grade.ID == "" is not strictly necessary here if we assume
	// an update operation will always have an ID. However, SaveGrade handles it.
	if grade.ID == "" {
		return models.Grade{}, fmt.Errorf("cannot update grade without an ID")
	}
	return s.SaveGrade(grade)
}

func (s *SQLiteGradeStore) DeleteGrade(id string) error {
	if id == "" {
		return fmt.Errorf("cannot delete grade without an ID")
	}
	stmt, err := s.DB.Prepare("DELETE FROM grades WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete grade statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to execute delete grade statement for ID %s: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting grade ID %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no grade found with ID '%s' to delete", id) // Or sql.ErrNoRows style
	}

	return nil
}
