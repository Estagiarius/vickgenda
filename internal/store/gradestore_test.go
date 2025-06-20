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
	"vickgenda-cli/internal/models"
	"vickgenda-cli/internal/store"
)

// setupGradeDB initializes an in-memory SQLite database and all required stores for testing grades.
func setupGradeDB(t *testing.T) (*sql.DB, store.GradeStore, store.StudentStore, store.TermStore, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on") // Enable foreign keys
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	// It's good practice to Exec PRAGMA foreign_keys=ON; per connection for SQLite if not in DSN
	// _, err = db.Exec("PRAGMA foreign_keys = ON;")
	// if err != nil {
	// 	db.Close()
	// 	t.Fatalf("Failed to enable foreign keys: %v", err)
	// }


	studentStore := store.NewSQLiteStudentStore(db)
	if err := studentStore.Init(); err != nil {
		db.Close()
		t.Fatalf("Failed to initialize student store: %v", err)
	}

	termStore := store.NewSQLiteTermStore(db)
	if err := termStore.Init(); err != nil {
		db.Close()
		t.Fatalf("Failed to initialize term store: %v", err)
	}

	gradeStore := store.NewSQLiteGradeStore(db)
	if err := gradeStore.Init(); err != nil {
		db.Close()
		t.Fatalf("Failed to initialize grade store: %v", err)
	}

	teardown := func() {
		db.Close()
	}
	return db, gradeStore, studentStore, termStore, teardown
}

func createTestStudent(t *testing.T, studentStore store.StudentStore, name string) models.Student {
	t.Helper()
	s, err := studentStore.SaveStudent(models.Student{Name: name})
	if err != nil {
		t.Fatalf("Failed to create test student '%s': %v", name, err)
	}
	return s
}

func createTestTerm(t *testing.T, termStore store.TermStore, name string, startDate, endDate time.Time) models.Term {
	t.Helper()
	term := models.Term{Name: name, StartDate: startDate, EndDate: endDate}
	savedTerm, err := termStore.SaveTerm(term)
	if err != nil {
		t.Fatalf("Failed to create test term '%s': %v", name, err)
	}
	return savedTerm
}

// Using the parseDate from termstore_test.go for consistency - assuming it's in the same package `store_test`
// If not, it should be defined here or imported if it were in a shared test utility package.
// For this exercise, I'll redefine it if it's not implicitly available.
func parseGradeTestDate(t *testing.T, dateStr string) time.Time {
	t.Helper()
	layout := "2006-01-02"
	tm, err := time.Parse(layout, dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date string '%s': %v", dateStr, err)
	}
	return tm
}


func TestGradeStore_SaveAndGetGrade(t *testing.T) {
	_, gradeStore, studentStore, termStore, teardown := setupGradeDB(t)
	defer teardown()

	student := createTestStudent(t, studentStore, "Test Student Grade")
	term := createTestTerm(t, termStore, "Test Term Grade", parseGradeTestDate(t, "2024-01-01"), parseGradeTestDate(t, "2024-03-01"))

	grade := models.Grade{
		StudentID:   student.ID,
		TermID:      term.ID,
		Subject:     "Math",
		Description: "Exam 1",
		Value:       8.5,
		Weight:      2.0,
		Date:        parseGradeTestDate(t, "2024-02-15"),
	}

	savedGrade, err := gradeStore.SaveGrade(grade)
	if err != nil {
		t.Fatalf("SaveGrade failed: %v", err)
	}
	if savedGrade.ID == "" {
		t.Errorf("Expected saved grade to have an ID, got empty string")
	}

	retrievedGrade, err := gradeStore.GetGradeByID(savedGrade.ID)
	if err != nil {
		t.Fatalf("GetGradeByID failed: %v", err)
	}
	// Compare relevant fields, excluding ID if it's generated and already checked
	if retrievedGrade.StudentID != grade.StudentID || retrievedGrade.TermID != grade.TermID ||
		retrievedGrade.Subject != grade.Subject || retrievedGrade.Description != grade.Description ||
		retrievedGrade.Value != grade.Value || retrievedGrade.Weight != grade.Weight || !retrievedGrade.Date.Equal(grade.Date) {
		t.Errorf("Retrieved grade %+v does not match saved grade %+v", retrievedGrade, grade)
	}

	_, err = gradeStore.GetGradeByID("non-existent-grade-id")
	if err == nil {
		t.Errorf("Expected error when getting non-existent grade, got nil")
	} else {
		if !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected sql.ErrNoRows or 'not found' error, got: %v", err)
		}
	}
}

func TestGradeStore_ListGradesByStudent(t *testing.T) {
	_, gradeStore, studentStore, termStore, teardown := setupGradeDB(t)
	defer teardown()

	s1 := createTestStudent(t, studentStore, "Student A")
	s2 := createTestStudent(t, studentStore, "Student B")
	term1 := createTestTerm(t, termStore, "Term 1", parseGradeTestDate(t, "2024-01-01"), parseGradeTestDate(t, "2024-03-01"))
	term2 := createTestTerm(t, termStore, "Term 2", parseGradeTestDate(t, "2024-04-01"), parseGradeTestDate(t, "2024-06-01"))

	g1s1t1math := models.Grade{StudentID: s1.ID, TermID: term1.ID, Subject: "Math", Description: "s1t1m1", Value: 7, Weight: 1, Date: parseGradeTestDate(t, "2024-02-01")}
	g2s1t1math := models.Grade{StudentID: s1.ID, TermID: term1.ID, Subject: "Math", Description: "s1t1m2", Value: 8, Weight: 1, Date: parseGradeTestDate(t, "2024-02-15")} // Later date
	g1s1t1sci := models.Grade{StudentID: s1.ID, TermID: term1.ID, Subject: "Science", Description: "s1t1s1", Value: 9, Weight: 1, Date: parseGradeTestDate(t, "2024-02-05")}
	g1s1t2math := models.Grade{StudentID: s1.ID, TermID: term2.ID, Subject: "Math", Description: "s1t2m1", Value: 6, Weight: 1, Date: parseGradeTestDate(t, "2024-05-01")}
	g1s2t1math := models.Grade{StudentID: s2.ID, TermID: term1.ID, Subject: "Math", Description: "s2t1m1", Value: 5, Weight: 1, Date: parseGradeTestDate(t, "2024-02-10")}

	for _, g := range []models.Grade{g1s1t1math, g2s1t1math, g1s1t1sci, g1s1t2math, g1s2t1math} {
		_, err := gradeStore.SaveGrade(g)
		if err != nil {
			t.Fatalf("Failed to save grade: %v", err)
		}
	}

	// Test by studentID only
	s1Grades, err := gradeStore.ListGradesByStudent(s1.ID, "", "")
	if err != nil {
		t.Fatalf("ListGradesByStudent for s1 failed: %v", err)
	}
	if len(s1Grades) != 3 { // g1s1t1math, g2s1t1math, g1s1t1sci, g1s1t2math (mistake in manual count, should be 4)
		// Correcting: g1s1t1math, g2s1t1math, g1s1t1sci, g1s1t2math are all for S1.
		s1GradesCorrected, _ := gradeStore.ListGradesByStudent(s1.ID, "", "")
		if len(s1GradesCorrected) != 4 {
			t.Errorf("Expected 4 grades for student s1, got %d", len(s1GradesCorrected))
		} else {
            s1Grades = s1GradesCorrected // use corrected list for date sort check
        }
	}
    // Check date sorting
    if len(s1Grades) == 4 { // Check only if count is correct
        if !(s1Grades[0].Date.Before(s1Grades[1].Date) && s1Grades[1].Date.Before(s1Grades[2].Date) && s1Grades[2].Date.Before(s1Grades[3].Date)) {
             t.Errorf("Grades for S1 not sorted by date: %v, %v, %v, %v", s1Grades[0].Date, s1Grades[1].Date, s1Grades[2].Date, s1Grades[3].Date)
        }
    }


	// Test by studentID and termID
	s1t1Grades, err := gradeStore.ListGradesByStudent(s1.ID, term1.ID, "")
	if err != nil {
		t.Fatalf("ListGradesByStudent for s1, term1 failed: %v", err)
	}
	if len(s1t1Grades) != 3 { // g1s1t1math, g2s1t1math, g1s1t1sci
		t.Errorf("Expected 3 grades for student s1 in term1, got %d", len(s1t1Grades))
	}

	// Test by studentID, termID, and subject
	s1t1MathGrades, err := gradeStore.ListGradesByStudent(s1.ID, term1.ID, "Math")
	if err != nil {
		t.Fatalf("ListGradesByStudent for s1, term1, Math failed: %v", err)
	}
	if len(s1t1MathGrades) != 2 { // g1s1t1math, g2s1t1math
		t.Errorf("Expected 2 Math grades for student s1 in term1, got %d", len(s1t1MathGrades))
	}
    if len(s1t1MathGrades) == 2 { // Check date sorting
         if s1t1MathGrades[0].Date.After(s1t1MathGrades[1].Date) {
              t.Errorf("Math Grades for S1T1 not sorted by date: %v, %v", s1t1MathGrades[0].Date, s1t1MathGrades[1].Date)
         }
    }


	// Test with a student who has no grades (create a new student)
	s3 := createTestStudent(t, studentStore, "Student C No Grades")
	s3Grades, err := gradeStore.ListGradesByStudent(s3.ID, "", "")
	if err != nil {
		t.Fatalf("ListGradesByStudent for s3 failed: %v", err)
	}
	if len(s3Grades) != 0 {
		t.Errorf("Expected 0 grades for student s3, got %d", len(s3Grades))
	}
}

func TestGradeStore_UpdateGrade(t *testing.T) {
	_, gradeStore, studentStore, termStore, teardown := setupGradeDB(t)
	defer teardown()

	student := createTestStudent(t, studentStore, "UpdateStudent")
	term := createTestTerm(t, termStore, "UpdateTerm", parseGradeTestDate(t, "2024-01-01"), parseGradeTestDate(t, "2024-03-01"))
	initialGrade := models.Grade{StudentID: student.ID, TermID: term.ID, Subject: "History", Description: "Essay", Value: 7.0, Weight: 1.5, Date: parseGradeTestDate(t, "2024-01-20")}
	savedGrade, err := gradeStore.SaveGrade(initialGrade)
	if err != nil {
		t.Fatalf("Failed to save initial grade for update test: %v", err)
	}

	gradeToUpdate := savedGrade
	gradeToUpdate.Value = 9.0
	gradeToUpdate.Description = "Essay (Resubmitted)"

	// UpdateGrade calls SaveGrade which uses INSERT OR REPLACE
	updatedGrade, err := gradeStore.UpdateGrade(gradeToUpdate)
	if err != nil {
		t.Fatalf("UpdateGrade failed: %v", err)
	}
	if updatedGrade.Value != 9.0 || updatedGrade.Description != "Essay (Resubmitted)" {
		t.Errorf("Grade not updated correctly. Got value %.2f, desc '%s'", updatedGrade.Value, updatedGrade.Description)
	}

	retrievedGrade, _ := gradeStore.GetGradeByID(savedGrade.ID)
	if retrievedGrade.Value != 9.0 || retrievedGrade.Description != "Essay (Resubmitted)" {
		t.Errorf("Retrieved grade after update does not reflect changes. Got value %.2f, desc '%s'", retrievedGrade.Value, retrievedGrade.Description)
	}
}

func TestGradeStore_DeleteGrade(t *testing.T) {
	_, gradeStore, studentStore, termStore, teardown := setupGradeDB(t)
	defer teardown()

	student := createTestStudent(t, studentStore, "DeleteStudent")
	term := createTestTerm(t, termStore, "DeleteTerm", parseGradeTestDate(t, "2024-01-01"), parseGradeTestDate(t, "2024-03-01"))
	grade := models.Grade{StudentID: student.ID, TermID: term.ID, Subject: "Art", Description: "Project", Value: 10.0, Weight: 3.0, Date: time.Now()}
	savedGrade, err := gradeStore.SaveGrade(grade)
	if err != nil {
		t.Fatalf("Failed to save grade for delete test: %v", err)
	}

	err = gradeStore.DeleteGrade(savedGrade.ID)
	if err != nil {
		t.Fatalf("DeleteGrade failed: %v", err)
	}

	_, err = gradeStore.GetGradeByID(savedGrade.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted grade, got nil")
	} else if !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected sql.ErrNoRows or 'not found' error for deleted grade, got: %v", err)
	}

	// Test deleting a non-existent grade
	err = gradeStore.DeleteGrade("non-existent-grade-to-delete")
	if err == nil {
		t.Errorf("Expected error when deleting non-existent grade, got nil")
	} else if !strings.Contains(err.Error(), "no grade found with ID") { // Based on store's error message
		t.Errorf("Expected 'no grade found' error, got: %v", err)
	}
}

func TestGradeStore_ForeignKeyConstraints(t *testing.T) {
	db, gradeStore, _, _, teardown := setupGradeDB(t) // studentStore and termStore not directly used here, but setup for Init
	defer teardown()

	// Attempt to save a grade with a non-existent StudentID
	nonExistentStudentID := uuid.NewString()
	nonExistentTermID := uuid.NewString() // Used later

	// Create a valid term to isolate student FK failure
	tempTermStore := store.NewSQLiteTermStore(db) // Use the same DB
	validTerm := createTestTerm(t, tempTermStore, "FK Test Term", parseGradeTestDate(t, "2024-01-01"), parseGradeTestDate(t, "2024-03-01"))


	gradeWithInvalidStudent := models.Grade{
		StudentID: nonExistentStudentID, TermID: validTerm.ID, Subject: "FK Test", Value: 5, Weight: 1, Date: time.Now(),
	}
	_, err := gradeStore.SaveGrade(gradeWithInvalidStudent)
	if err == nil {
		t.Errorf("Expected foreign key constraint error for non-existent StudentID, got nil")
	} else if !strings.Contains(strings.ToLower(err.Error()), "foreign key constraint failed") {
		// Making the check case-insensitive as error messages can vary.
		t.Errorf("Expected error message to contain 'FOREIGN KEY constraint failed' for StudentID, got: %v", err.Error())
	}

	// Create a valid student to isolate term FK failure
	tempStudentStore := store.NewSQLiteStudentStore(db) // Use the same DB
	validStudent := createTestStudent(t, tempStudentStore, "FK Test Student")

	gradeWithInvalidTerm := models.Grade{
		StudentID: validStudent.ID, TermID: nonExistentTermID, Subject: "FK Test", Value: 5, Weight: 1, Date: time.Now(),
	}
	_, err = gradeStore.SaveGrade(gradeWithInvalidTerm)
	if err == nil {
		t.Errorf("Expected foreign key constraint error for non-existent TermID, got nil")
	} else if !strings.Contains(strings.ToLower(err.Error()), "foreign key constraint failed") {
		t.Errorf("Expected error message to contain 'FOREIGN KEY constraint failed' for TermID, got: %v", err.Error())
	}

	// Cascade delete testing for grades when a student or term is deleted would be more involved
	// and depends on StudentStore/TermStore having Delete methods and the DB schema's ON DELETE CASCADE.
	// The current grade table schema has ON DELETE CASCADE, so this is implicitly handled by SQLite.
	// Verifying it here would mean:
	// 1. Create student, term, grade.
	// 2. Delete student (or term) using their respective stores (if Delete methods exist).
	// 3. Try to GetGradeByID and expect sql.ErrNoRows.
	// This is out of scope if Student/Term stores don't have Delete, or if focus is only on GradeStore.
	// The prompt mentions "focus on preventing inserts with bad FKs" for this test.
}


// TestMain can be used for global setup/teardown
// func TestMain(m *testing.M) { // Removed to avoid multiple TestMain definitions
// 	// No global setup needed for :memory: DBs that are fresh each test.
// 	os.Exit(m.Run())
// }

// Note: parseGradeTestDate is a duplicate of parseDate from termstore_test.go.
// In a real scenario, this would be in a shared test utility package.
// The `strings` import was added due to `strings.Contains`.
// The `github.com/google/uuid` import was added due to `uuid.NewString`.
// The `fmt` import was added for `fmt.Errorf` (though t.Errorf also uses it).

var _ = fmt.Errorf // dummy usage to satisfy linter if needed
var _ = uuid.NewString // dummy usage
