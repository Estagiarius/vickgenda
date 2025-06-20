package store_test

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3
	"vickgenda/internal/models"
	"vickgenda/internal/store"
)

// setupAulaDB initializes an in-memory SQLite database and an AulaStore for testing.
func setupAulaDB(t *testing.T) (*sql.DB, store.AulaStore, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	aulaStore := store.NewSQLiteAulaStore(db)
	if err := aulaStore.Init(); err != nil {
		db.Close()
		t.Fatalf("Failed to initialize aula store: %v", err)
	}

	teardown := func() {
		db.Close()
	}
	return db, aulaStore, teardown
}

// parseAulaTestDate is a helper for parsing dates in tests.
// Note: This is similar to helpers in other _test.go files.
// In a real project, consider a shared testutil package.
func parseAulaTestDate(t *testing.T, dateStr string, includeTime ...string) time.Time {
	t.Helper()
	layout := "2006-01-02"
	if len(includeTime) > 0 && includeTime[0] != "" {
		layout = "2006-01-02 15:04"
		dateStr = dateStr + " " + includeTime[0]
	}

	tm, err := time.Parse(layout, dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date string '%s' with layout '%s': %v", dateStr, layout, err)
	}
	return tm
}

func TestAulaStore_SaveAndGetLesson(t *testing.T) {
	_, aulaStore, teardown := setupAulaDB(t)
	defer teardown()

	lesson := models.Lesson{
		Subject:      "Math",
		Topic:        "Algebra Basics",
		Date:         parseAulaTestDate(t, "2024-08-01", "10:00"),
		ClassID:      "TurmaA",
		Plan:         "Introduction to variables.",
		Observations: "Students engaged well.",
	}

	savedLesson, err := aulaStore.SaveLesson(lesson)
	if err != nil {
		t.Fatalf("SaveLesson failed: %v", err)
	}
	if savedLesson.ID == "" {
		t.Errorf("Expected saved lesson to have an ID, got empty string")
	}
	// Check if other fields are preserved
	if savedLesson.Subject != lesson.Subject || savedLesson.Topic != lesson.Topic || !savedLesson.Date.Equal(lesson.Date) {
		t.Errorf("Saved lesson data does not match input. Got %+v, expected (similar to) %+v", savedLesson, lesson)
	}


	retrievedLesson, err := aulaStore.GetLessonByID(savedLesson.ID)
	if err != nil {
		t.Fatalf("GetLessonByID failed: %v", err)
	}
	if retrievedLesson.ID != savedLesson.ID || retrievedLesson.Subject != lesson.Subject ||
		retrievedLesson.Topic != lesson.Topic || !retrievedLesson.Date.Equal(lesson.Date) ||
		retrievedLesson.ClassID != lesson.ClassID || retrievedLesson.Plan != lesson.Plan ||
		retrievedLesson.Observations != lesson.Observations {
		t.Errorf("Retrieved lesson %+v does not match saved lesson %+v", retrievedLesson, savedLesson)
	}

	_, err = aulaStore.GetLessonByID("non-existent-lesson-id")
	if err == nil {
		t.Errorf("Expected error when getting non-existent lesson, got nil")
	} else {
		if !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected sql.ErrNoRows or 'not found' error, got: %v", err)
		}
	}
}

func TestAulaStore_ListLessons_Filtering(t *testing.T) {
	_, aulaStore, teardown := setupAulaDB(t)
	defer teardown()

	l1 := models.Lesson{Subject: "Math", Topic: "Algebra", Date: parseAulaTestDate(t, "2024-08-01", "09:00"), ClassID: "TurmaA", Plan: "P1"}
	l2 := models.Lesson{Subject: "math", Topic: "Geometry", Date: parseAulaTestDate(t, "2024-08-02", "10:00"), ClassID: "turmaa", Plan: "P2"} // different case for subject/class
	l3 := models.Lesson{Subject: "Science", Topic: "Biology", Date: parseAulaTestDate(t, "2024-08-01", "11:00"), ClassID: "TurmaB", Plan: "P3"}
	l4 := models.Lesson{Subject: "History", Topic: "WW2", Date: parseAulaTestDate(t, "2024-09-05", "14:00"), ClassID: "TurmaA", Plan: "P4"}
	l5 := models.Lesson{Subject: "Math", Topic: "Calculus", Date: parseAulaTestDate(t, "2023-08-05", "09:00"), ClassID: "TurmaC", Plan: "P5"}


	lessonsToSave := []models.Lesson{l1, l2, l3, l4, l5}
	savedLessons := make(map[string]models.Lesson) // Store by original topic for easy reference
	for i, l := range lessonsToSave {
		saved, err := aulaStore.SaveLesson(l)
		if err != nil {
			t.Fatalf("Failed to save lesson %d for filtering test: %v", i+1, err)
		}
		savedLessons[l.Topic] = saved
	}

	t.Run("NoFilters", func(t *testing.T) {
		lessons, err := aulaStore.ListLessons("", "", "", "", "")
		if err != nil {
			t.Fatalf("ListLessons with no filters failed: %v", err)
		}
		if len(lessons) != 5 {
			t.Errorf("Expected 5 lessons with no filters, got %d", len(lessons))
		}
		// Check sorting by date (oldest first)
		for i := 0; i < len(lessons)-1; i++ {
			if lessons[i].Date.After(lessons[i+1].Date) {
				t.Errorf("Lessons not sorted by date: lesson %s (%v) is after %s (%v)", lessons[i].Topic, lessons[i].Date, lessons[i+1].Topic, lessons[i+1].Date)
			}
		}
	})

	t.Run("ByDisciplina", func(t *testing.T) {
		lessons, err := aulaStore.ListLessons("Math", "", "", "", "") // Should be case-insensitive
		if err != nil {
			t.Fatalf("ListLessons by disciplina 'Math' failed: %v", err)
		}
		if len(lessons) != 3 { // l1, l2, l5
			t.Errorf("Expected 3 'Math' lessons, got %d. Lessons: %+v", len(lessons), lessons)
		}
	})

	t.Run("ByTurma", func(t *testing.T) {
		lessons, err := aulaStore.ListLessons("", "TurmaA", "", "", "") // Should be case-insensitive
		if err != nil {
			t.Fatalf("ListLessons by turma 'TurmaA' failed: %v", err)
		}
		if len(lessons) != 3 { // l1, l2, l4
			t.Errorf("Expected 3 'TurmaA' lessons, got %d. Lessons: %+v", len(lessons), lessons)
		}
	})

	t.Run("ByPeriodo", func(t *testing.T) {
		// Periodo: 2024-08-01 to 2024-08-02
		lessons, err := aulaStore.ListLessons("", "", "01-08-2024:02-08-2024", "", "")
		if err != nil {
			t.Fatalf("ListLessons by periodo failed: %v", err)
		}
		if len(lessons) != 3 { // l1, l2, l3
			t.Errorf("Expected 3 lessons in periodo 01-08 to 02-08-2024, got %d. Lessons: %+v", len(lessons), lessons)
		}
	})

	t.Run("ByMes", func(t *testing.T) {
		lessons, err := aulaStore.ListLessons("", "", "", "08-2024", "") // August 2024
		if err != nil {
			t.Fatalf("ListLessons by mes '08-2024' failed: %v", err)
		}
		if len(lessons) != 3 { // l1, l2, l3
			t.Errorf("Expected 3 lessons in mes '08-2024', got %d. Lessons: %+v", len(lessons), lessons)
		}
	})

	t.Run("ByAno", func(t *testing.T) {
		lessons, err := aulaStore.ListLessons("", "", "", "", "2024")
		if err != nil {
			t.Fatalf("ListLessons by ano '2024' failed: %v", err)
		}
		if len(lessons) != 4 { // l1, l2, l3, l4
			t.Errorf("Expected 4 lessons in ano '2024', got %d. Lessons: %+v", len(lessons), lessons)
		}
	})

	t.Run("CombinationDisciplinaAndAno", func(t *testing.T) {
		lessons, err := aulaStore.ListLessons("Math", "", "", "", "2024")
		if err != nil {
			t.Fatalf("ListLessons by Math and 2024 failed: %v", err)
		}
		if len(lessons) != 2 { // l1, l2
			t.Errorf("Expected 2 Math lessons in 2024, got %d. Lessons: %+v", len(lessons), lessons)
		}
	})
}


func TestAulaStore_UpdateLessonPlan(t *testing.T) {
	_, aulaStore, teardown := setupAulaDB(t)
	defer teardown()

	initialLesson := models.Lesson{Subject: "Physics", Topic: "Newton's Laws", Date: time.Now(), ClassID: "Sci101", Plan: "Old plan", Observations: "Old obs"}
	savedLesson, err := aulaStore.SaveLesson(initialLesson)
	if err != nil {
		t.Fatalf("Failed to save initial lesson for update test: %v", err)
	}

	newPlan := "Detailed new plan."
	newObs := "Students need more examples."
	updatedLesson, err := aulaStore.UpdateLessonPlan(savedLesson.ID, newPlan, newObs)
	if err != nil {
		t.Fatalf("UpdateLessonPlan failed: %v", err)
	}

	if updatedLesson.Plan != newPlan || updatedLesson.Observations != newObs {
		t.Errorf("Lesson plan/observations not updated. Got Plan: '%s', Obs: '%s'", updatedLesson.Plan, updatedLesson.Observations)
	}
	// Check other fields remain unchanged
	if updatedLesson.ID != savedLesson.ID || updatedLesson.Subject != initialLesson.Subject || updatedLesson.Topic != initialLesson.Topic {
		t.Errorf("Other fields changed during update. Original: %+v, Updated: %+v", savedLesson, updatedLesson)
	}

	// Test updating a non-existent lesson
	_, err = aulaStore.UpdateLessonPlan("non-existent-id", "plan", "obs")
	if err == nil {
		t.Errorf("Expected error when updating non-existent lesson, got nil")
	} else if !strings.Contains(err.Error(), "not found") { // UpdateLessonPlan itself calls GetLessonByID
		t.Errorf("Expected 'not found' error for non-existent update, got: %v", err)
	}
}

func TestAulaStore_DeleteLesson(t *testing.T) {
	_, aulaStore, teardown := setupAulaDB(t)
	defer teardown()

	lesson := models.Lesson{Subject: "Geography", Topic: "Capitals", Date: time.Now(), ClassID: "Geo202"}
	savedLesson, err := aulaStore.SaveLesson(lesson)
	if err != nil {
		t.Fatalf("Failed to save lesson for delete test: %v", err)
	}

	err = aulaStore.DeleteLesson(savedLesson.ID)
	if err != nil {
		t.Fatalf("DeleteLesson failed: %v", err)
	}

	_, err = aulaStore.GetLessonByID(savedLesson.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted lesson, got nil")
	} else if !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected sql.ErrNoRows or 'not found' error for deleted lesson, got: %v", err)
	}

	// Test deleting a non-existent lesson
	err = aulaStore.DeleteLesson("non-existent-lesson-to-delete")
	if err == nil {
		t.Errorf("Expected error when deleting non-existent lesson, got nil")
	} else if !strings.Contains(err.Error(), "no lesson found with ID") { // Based on store's error message
		t.Errorf("Expected 'no lesson found' error for non-existent delete, got: %v", err)
	}
}

// TestMain for global setup/teardown, if any.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// Dummy usage for linters if imports are not directly used in some test variations.
var _ = fmt.Errorf
var _ = uuid.NewString
var _ = errors.Is
var _ = time.Now
var _ = strings.Contains
var _ = models.Lesson{}
var _ = store.AulaStore(nil)
var _ = sql.Open
var _ = os.Exit
var _ = testing.T{}
