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

const createLessonsTableSQL = `
CREATE TABLE IF NOT EXISTS lessons (
    id TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    topic TEXT NOT NULL,
    date DATETIME NOT NULL,
    class_id TEXT,
    plan TEXT,
    observations TEXT
);`

// InitLessonsTable ensures the lessons table exists.
func InitLessonsTable() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("InitLessonsTable (GetDB): %w", err)
	}
	_, err = database.Exec(createLessonsTableSQL)
	if err != nil {
		return fmt.Errorf("InitLessonsTable (Exec): %w", err)
	}
	return nil
}

// CreateLesson saves a new lesson to the database.
func CreateLesson(lesson models.Lesson) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("CreateLesson (GetDB): %w", err)
	}

	stmt, err := database.Prepare("INSERT INTO lessons (id, subject, topic, date, class_id, plan, observations) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("CreateLesson (Prepare): %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(lesson.ID, lesson.Subject, lesson.Topic, lesson.Date, lesson.ClassID, lesson.Plan, lesson.Observations)
	if err != nil {
		return fmt.Errorf("CreateLesson (Exec): %w", err)
	}
	return nil
}

// GetLessonByID retrieves a lesson by its ID.
func GetLessonByID(id string) (models.Lesson, error) {
	database, err := db.GetDB()
	if err != nil {
		return models.Lesson{}, fmt.Errorf("GetLessonByID (GetDB): %w", err)
	}
	row := database.QueryRow("SELECT id, subject, topic, date, class_id, plan, observations FROM lessons WHERE id = ?", id)
	var lesson models.Lesson
	err = row.Scan(&lesson.ID, &lesson.Subject, &lesson.Topic, &lesson.Date, &lesson.ClassID, &lesson.Plan, &lesson.Observations)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Lesson{}, fmt.Errorf("lesson with ID '%s' not found", id)
		}
		return models.Lesson{}, fmt.Errorf("GetLessonByID (Scan): %w", err)
	}
	return lesson, nil
}

// UpdateLessonPlanAndObservations updates only the plan and observations of a lesson.
func UpdateLessonPlanAndObservations(id, plan, observations string) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("UpdateLessonPlanAndObservations (GetDB): %w", err)
	}

	stmt, err := database.Prepare("UPDATE lessons SET plan = ?, observations = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("UpdateLessonPlanAndObservations (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(plan, observations, id)
	if err != nil {
		return fmt.Errorf("UpdateLessonPlanAndObservations (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateLessonPlanAndObservations: no lesson found with ID '%s'", id)
	}
	return nil
}

// UpdateLesson updates all fields of an existing lesson (more general than UpdateLessonPlanAndObservations).
func UpdateLesson(lesson models.Lesson) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("UpdateLesson (GetDB): %w", err)
	}
	stmt, err := database.Prepare("UPDATE lessons SET subject = ?, topic = ?, date = ?, class_id = ?, plan = ?, observations = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("UpdateLesson (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(lesson.Subject, lesson.Topic, lesson.Date, lesson.ClassID, lesson.Plan, lesson.Observations, lesson.ID)
	if err != nil {
		return fmt.Errorf("UpdateLesson (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateLesson: no lesson found with ID '%s'", lesson.ID)
	}
	return nil
}

// DeleteLesson removes a lesson from the database by its ID.
func DeleteLesson(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("DeleteLesson (GetDB): %w", err)
	}
	stmt, err := database.Prepare("DELETE FROM lessons WHERE id = ?")
	if err != nil {
		return fmt.Errorf("DeleteLesson (Prepare): %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("DeleteLesson (Exec): %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("DeleteLesson: no lesson found with ID '%s'", id)
	}
	return nil
}

// ListLessons performs querying of lessons based on provided filters.
// Filters map can contain: "subject", "class_id", "period_start", "period_end", "month_year" (MM-YYYY), "year" (YYYY)
// Results are ordered by date.
func ListLessons(filters map[string]string) ([]models.Lesson, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, fmt.Errorf("ListLessons (GetDB): %w", err)
	}

	var queryArgs []interface{}
	var conditions []string

	query := "SELECT id, subject, topic, date, class_id, plan, observations FROM lessons"

	if val, ok := filters["subject"]; ok && val != "" {
		conditions = append(conditions, "subject = ?")
		queryArgs = append(queryArgs, val)
	}
	if val, ok := filters["class_id"]; ok && val != "" {
		conditions = append(conditions, "class_id = ?")
		queryArgs = append(queryArgs, val)
	}
	if val, ok := filters["year"]; ok && val != "" {
		conditions = append(conditions, "strftime('%Y', date) = ?")
		queryArgs = append(queryArgs, val)
	}
	if val, ok := filters["month_year"]; ok && val != "" { // Format: MM-YYYY
		parts := strings.Split(val, "-")
		if len(parts) == 2 {
			conditions = append(conditions, "strftime('%m-%Y', date) = ?")
			queryArgs = append(queryArgs, val)
		} else {
			log.Printf("Warning: Invalid month_year filter format: %s", val)
		}
	}
	if startStr, startOk := filters["period_start"]; startOk && startStr != "" {
		if endStr, endOk := filters["period_end"]; endOk && endStr != "" {
			// Assuming dates are YYYY-MM-DD HH:MM:SS or just YYYY-MM-DD.
			// For period, ensure the end_date includes the whole day if time is not specified.
			conditions = append(conditions, "date >= ? AND date <= ?")
			queryArgs = append(queryArgs, startStr, endStr)
		} else {
			log.Printf("Warning: period_start specified without period_end")
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY date ASC"

	rows, err := database.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("ListLessons (Query - '%s', Args: %v): %w", query, queryArgs, err)
	}
	defer rows.Close()

	var lessons []models.Lesson
	for rows.Next() {
		var lesson models.Lesson
		err := rows.Scan(&lesson.ID, &lesson.Subject, &lesson.Topic, &lesson.Date, &lesson.ClassID, &lesson.Plan, &lesson.Observations)
		if err != nil {
			log.Printf("ERROR: ListLessons (Scan): %v. Query: %s, Args: %v", err, query, queryArgs)
			continue
		}
		lessons = append(lessons, lesson)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ListLessons (RowsErr): %w", err)
	}
	return lessons, nil
}

// ClearLessonsTableForTesting is a helper for tests to clear all lessons.
func ClearLessonsTableForTesting() error {
	database, err := db.GetDB()
	if err != nil {
		return fmt.Errorf("ClearLessonsTableForTesting (GetDB): %w", err)
	}
	_, err = database.Exec("DELETE FROM lessons")
	if err != nil {
		return fmt.Errorf("ClearLessonsTableForTesting (Exec): %w", err)
	}
	return nil
}
