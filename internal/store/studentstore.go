package store

import (
	"database/sql"
	"fmt"

	"vickgenda/internal/db"
	"vickgenda/internal/models"
)

const createStudentsTableSQL = `
CREATE TABLE IF NOT EXISTS students (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);`

// InitStudentsTable ensures the students table exists.
func InitStudentsTable() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("InitStudentsTable (GetDB): %w", err)
	}
	_, err = database.Exec(createStudentsTableSQL)
	if err != nil {
		return fmt.Errorf("InitStudentsTable (Exec): %w", err)
	}
	return nil
}

// CreateStudent saves a new student.
func CreateStudent(student models.Student) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("CreateStudent (GetDB): %w", err)
	}

	stmt, err := database.Prepare("INSERT INTO students (id, name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("CreateStudent (Prepare): %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(student.ID, student.Name)
	if err != nil {
		// Consider checking for sqlite3.ErrConstraintUnique if ID should be globally unique and re-insertion is an error
		return fmt.Errorf("CreateStudent (Exec): %w", err)
	}
	return nil
}

// UpdateStudent updates an existing student's name.
func UpdateStudent(student models.Student) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("UpdateStudent (GetDB): %w", err)
	}

	stmt, err := database.Prepare("UPDATE students SET name = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("UpdateStudent (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(student.Name, student.ID)
	if err != nil {
		return fmt.Errorf("UpdateStudent (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateStudent: no student found with ID '%s'", student.ID)
	}
	return nil
}

// GetStudentByID retrieves a student by their ID.
func GetStudentByID(id string) (models.Student, error) {
	database, err := db.GetDB()
	if err != nil {
		return models.Student{}, fmt.Errorf("GetStudentByID (GetDB): %w", err)
	}
	row := database.QueryRow("SELECT id, name FROM students WHERE id = ?", id)
	var student models.Student
	err = row.Scan(&student.ID, &student.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, fmt.Errorf("student with ID '%s' not found", id)
		}
		return models.Student{}, fmt.Errorf("GetStudentByID (Scan): %w", err)
	}
	return student, nil
}

// ListStudents retrieves all students.
func ListStudents() ([]models.Student, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, fmt.Errorf("ListStudents (GetDB): %w", err)
	}
	rows, err := database.Query("SELECT id, name FROM students ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("ListStudents (Query): %w", err)
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.Name)
		if err != nil {
			// Log or handle individual row scan errors
			return nil, fmt.Errorf("ListStudents (Scan): %w", err)
		}
		students = append(students, student)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ListStudents (RowsErr): %w", err)
	}
	return students, nil
}

// DeleteStudent removes a student from the database by their ID.
// Note: If ON DELETE CASCADE is set for grades.student_id, deleting a student will also delete their grades.
func DeleteStudent(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("DeleteStudent (GetDB): %w", err)
	}
	stmt, err := database.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		return fmt.Errorf("DeleteStudent (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("DeleteStudent (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("DeleteStudent: no student found with ID '%s'", id)
	}
	return nil
}

// ClearStudentsTableForTesting is a helper for tests to clear all students.
func ClearStudentsTableForTesting() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("ClearStudentsTableForTesting (GetDB): %w", err)
	}
	_, err = database.Exec("DELETE FROM students")
	if err != nil {
		return fmt.Errorf("ClearStudentsTableForTesting (Exec): %w", err)
	}
	return nil
}
