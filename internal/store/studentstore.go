package store

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"vickgenda-cli/internal/models"
)

// StudentStore defines the interface for student persistence.
type StudentStore interface {
	Init() error
	SaveStudent(student models.Student) (models.Student, error)
	GetStudentByID(id string) (models.Student, error)
	ListStudents() ([]models.Student, error)
}

// SQLiteStudentStore implements the StudentStore interface using SQLite.
type SQLiteStudentStore struct {
	DB *sql.DB
}

// NewSQLiteStudentStore creates a new SQLiteStudentStore.
func NewSQLiteStudentStore(db *sql.DB) StudentStore {
	return &SQLiteStudentStore{DB: db}
}

// Init creates the 'students' table if it doesn't exist.
func (s *SQLiteStudentStore) Init() error {
	stmt, err := s.DB.Prepare(`
		CREATE TABLE IF NOT EXISTS students (
			id TEXT PRIMARY KEY,
			name TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare create students table statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute create students table statement: %w", err)
	}
	return nil
}

// SaveStudent saves a student to the database. If the student's ID is empty, a new UUID is generated.
func (s *SQLiteStudentStore) SaveStudent(student models.Student) (models.Student, error) {
	if student.ID == "" {
		student.ID = uuid.NewString()
	}

	stmt, err := s.DB.Prepare("INSERT OR REPLACE INTO students (id, name) VALUES (?, ?)")
	if err != nil {
		return models.Student{}, fmt.Errorf("failed to prepare save student statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(student.ID, student.Name)
	if err != nil {
		return models.Student{}, fmt.Errorf("failed to execute save student statement for student ID %s: %w", student.ID, err)
	}
	return student, nil
}

// GetStudentByID retrieves a student from the database by their ID.
func (s *SQLiteStudentStore) GetStudentByID(id string) (models.Student, error) {
	var student models.Student
	err := s.DB.QueryRow("SELECT id, name FROM students WHERE id = ?", id).Scan(&student.ID, &student.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, fmt.Errorf("student with ID '%s' not found: %w", id, err)
		}
		return models.Student{}, fmt.Errorf("failed to get student by ID '%s': %w", id, err)
	}
	return student, nil
}

// ListStudents retrieves all students from the database, ordered by name.
func (s *SQLiteStudentStore) ListStudents() ([]models.Student, error) {
	rows, err := s.DB.Query("SELECT id, name FROM students ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query students: %w", err)
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.Name); err != nil {
			return nil, fmt.Errorf("failed to scan student during ListStudents: %w", err)
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of students: %w", err)
	}

	return students, nil
}
