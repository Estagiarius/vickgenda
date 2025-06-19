package store

import (
	"database/sql"
	"fmt"
	"time"

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
	// Placeholder: Create 'grades' table
    // CREATE TABLE IF NOT EXISTS grades (id TEXT PRIMARY KEY, student_id TEXT, term_id TEXT, subject TEXT, description TEXT, value REAL, weight REAL, date DATETIME);
	fmt.Println("Placeholder: SQLiteGradeStore.Init() called")
	return nil
}

func (s *SQLiteGradeStore) SaveGrade(grade models.Grade) (models.Grade, error) {
	fmt.Printf("Placeholder: SQLiteGradeStore.SaveGrade() for grade ID %s\n", grade.ID)
	return grade, nil
}

func (s *SQLiteGradeStore) GetGradeByID(id string) (models.Grade, error) {
	fmt.Printf("Placeholder: SQLiteGradeStore.GetGradeByID() for ID %s\n", id)
	return models.Grade{ID: id, Description: "Placeholder Grade"}, nil
}

func (s *SQLiteGradeStore) ListGradesByStudent(studentID, termID, subject string) ([]models.Grade, error) {
	fmt.Printf("Placeholder: SQLiteGradeStore.ListGradesByStudent() for student %s\n", studentID)
	return []models.Grade{}, nil
}

func (s *SQLiteGradeStore) UpdateGrade(grade models.Grade) (models.Grade, error) {
    fmt.Printf("Placeholder: SQLiteGradeStore.UpdateGrade() for grade ID %s\n", grade.ID)
    return grade, nil
}

func (s *SQLiteGradeStore) DeleteGrade(id string) error {
    fmt.Printf("Placeholder: SQLiteGradeStore.DeleteGrade() for ID %s\n", id)
    return nil
}
