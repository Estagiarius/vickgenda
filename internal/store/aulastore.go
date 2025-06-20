package store

import (
	"database/sql" // Import for eventual SQLite use
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	stmt, err := s.DB.Prepare(`
		CREATE TABLE IF NOT EXISTS lessons (
			id TEXT PRIMARY KEY,
			subject TEXT,
			topic TEXT,
			date DATETIME,
			class_id TEXT,
			plan TEXT,
			observations TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare create lessons table statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute create lessons table statement: %w", err)
	}
	return nil
}

func (s *SQLiteAulaStore) SaveLesson(lesson models.Lesson) (models.Lesson, error) {
	if lesson.ID == "" {
		lesson.ID = uuid.NewString()
	}

	stmt, err := s.DB.Prepare(`
		INSERT OR REPLACE INTO lessons
		(id, subject, topic, date, class_id, plan, observations)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("failed to prepare save lesson statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(lesson.ID, lesson.Subject, lesson.Topic, lesson.Date, lesson.ClassID, lesson.Plan, lesson.Observations)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("failed to execute save lesson statement for lesson ID %s: %w", lesson.ID, err)
	}
	return lesson, nil
}

func (s *SQLiteAulaStore) GetLessonByID(id string) (models.Lesson, error) {
	var lesson models.Lesson
	err := s.DB.QueryRow(
		"SELECT id, subject, topic, date, class_id, plan, observations FROM lessons WHERE id = ?",
		id,
	).Scan(
		&lesson.ID, &lesson.Subject, &lesson.Topic, &lesson.Date,
		&lesson.ClassID, &lesson.Plan, &lesson.Observations,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Lesson{}, fmt.Errorf("lesson with ID '%s' not found: %w", id, err)
		}
		return models.Lesson{}, fmt.Errorf("failed to get lesson by ID '%s': %w", id, err)
	}
	return lesson, nil
}

func (s *SQLiteAulaStore) ListLessons(disciplina, turma, periodo, mes, ano string) ([]models.Lesson, error) {
	var queryFilters []string
	var args []interface{}

	baseQuery := "SELECT id, subject, topic, date, class_id, plan, observations FROM lessons"

	if disciplina != "" {
		queryFilters = append(queryFilters, "LOWER(subject) = LOWER(?)")
		args = append(args, disciplina)
	}
	if turma != "" {
		queryFilters = append(queryFilters, "LOWER(class_id) = LOWER(?)")
		args = append(args, turma)
	}
	if periodo != "" {
		parts := strings.Split(periodo, ":")
		if len(parts) == 2 {
			startDateStr := parts[0]
			endDateStr := parts[1]
			// Assuming format "dd-mm-yyyy"
			layout := "02-01-2006"
			startDate, err := time.Parse(layout, startDateStr)
			if err == nil {
				endDate, err := time.Parse(layout, endDateStr)
				if err == nil {
					// Ensure endDate includes the whole day
					endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
					queryFilters = append(queryFilters, "date BETWEEN ? AND ?")
					args = append(args, startDate, endDate)
				} else {
					return nil, fmt.Errorf("invalid end date format in periodo: %s", endDateStr)
				}
			} else {
				return nil, fmt.Errorf("invalid start date format in periodo: %s", startDateStr)
			}
		} else {
			return nil, fmt.Errorf("invalid periodo format, expected 'dd-mm-yyyy:dd-mm-yyyy': %s", periodo)
		}
	}
	if mes != "" { // Format "mm-yyyy"
		// SQLite strftime: '%m-%Y' for "mm-yyyy"
		// We need to convert "mm-yyyy" to "yyyy-mm" for SQLite's strftime('%Y-%m', date)
		parts := strings.Split(mes, "-")
		if len(parts) == 2 {
			sqliteMonthYear := parts[1] + "-" + parts[0] // yyyy-mm
			queryFilters = append(queryFilters, "strftime('%Y-%m', date) = ?")
			args = append(args, sqliteMonthYear)
		} else {
			return nil, fmt.Errorf("invalid mes format, expected 'mm-yyyy': %s", mes)
		}
	}
	if ano != "" { // Format "yyyy"
		queryFilters = append(queryFilters, "strftime('%Y', date) = ?")
		args = append(args, ano)
	}

	query := baseQuery
	if len(queryFilters) > 0 {
		query += " WHERE " + strings.Join(queryFilters, " AND ")
	}
	query += " ORDER BY date ASC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons with filters: %w", err)
	}
	defer rows.Close()

	var lessons []models.Lesson
	for rows.Next() {
		var lesson models.Lesson
		if err := rows.Scan(
			&lesson.ID, &lesson.Subject, &lesson.Topic, &lesson.Date,
			&lesson.ClassID, &lesson.Plan, &lesson.Observations,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lesson during ListLessons: %w", err)
		}
		lessons = append(lessons, lesson)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of lessons: %w", err)
	}

	return lessons, nil
}

func (s *SQLiteAulaStore) UpdateLessonPlan(id, novoPlano, novasObservacoes string) (models.Lesson, error) {
	// First, check if the lesson exists
	_, err := s.GetLessonByID(id)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("cannot update lesson plan, lesson with ID '%s' not found: %w", id, err)
	}

	stmt, err := s.DB.Prepare("UPDATE lessons SET plan = ?, observations = ? WHERE id = ?")
	if err != nil {
		return models.Lesson{}, fmt.Errorf("failed to prepare update lesson plan statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(novoPlano, novasObservacoes, id)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("failed to execute update lesson plan for ID %s: %w", id, err)
	}

	// Return the updated lesson
	return s.GetLessonByID(id)
}

func (s *SQLiteAulaStore) DeleteLesson(id string) error {
	if id == "" {
		return fmt.Errorf("cannot delete lesson without an ID")
	}
	stmt, err := s.DB.Prepare("DELETE FROM lessons WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete lesson statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to execute delete lesson statement for ID %s: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting lesson ID %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no lesson found with ID '%s' to delete", id)
	}

	return nil
}
