package store_test

import (
	"database/sql"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3
	"vickgenda-cli/internal/models"
	"vickgenda-cli/internal/store"
)

// setupStudentDB initializes an in-memory SQLite database and a StudentStore for testing.
func setupStudentDB(t *testing.T) (*sql.DB, store.StudentStore, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	studentStore := store.NewSQLiteStudentStore(db)
	if err := studentStore.Init(); err != nil {
		db.Close()
		t.Fatalf("Failed to initialize student store: %v", err)
	}

	teardown := func() {
		db.Close()
	}
	return db, studentStore, teardown
}

func TestStudentStore_SaveAndGetStudent(t *testing.T) {
	_, studentStore, teardown := setupStudentDB(t)
	defer teardown()

	student := models.Student{Name: "Alice Wonderland"}
	savedStudent, err := studentStore.SaveStudent(student)
	if err != nil {
		t.Fatalf("SaveStudent failed: %v", err)
	}
	if savedStudent.ID == "" {
		t.Errorf("Expected saved student to have an ID, got empty string")
	}
	if savedStudent.Name != student.Name {
		t.Errorf("Expected student name to be '%s', got '%s'", student.Name, savedStudent.Name)
	}

	retrievedStudent, err := studentStore.GetStudentByID(savedStudent.ID)
	if err != nil {
		t.Fatalf("GetStudentByID failed: %v", err)
	}
	if retrievedStudent.ID != savedStudent.ID || retrievedStudent.Name != savedStudent.Name {
		t.Errorf("Retrieved student %+v does not match saved student %+v", retrievedStudent, savedStudent)
	}

	_, err = studentStore.GetStudentByID("non-existent-id")
	if err == nil {
		t.Errorf("Expected error when getting non-existent student, got nil")
	} else {
		if !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected sql.ErrNoRows or 'not found' error, got: %v", err)
		}
	}
}

func TestStudentStore_ListStudents(t *testing.T) {
	_, studentStore, teardown := setupStudentDB(t)
	defer teardown()

	// Test with no students
	students, err := studentStore.ListStudents()
	if err != nil {
		t.Fatalf("ListStudents failed when empty: %v", err)
	}
	if len(students) != 0 {
		t.Errorf("Expected 0 students when store is empty, got %d", len(students))
	}

	// Save multiple students
	student1 := models.Student{Name: "Charlie Brown"}
	student2 := models.Student{Name: "Alice Wonderland"} // Alice should come before Charlie when sorted
	student3 := models.Student{Name: "Bob The Builder"}  // Bob between Alice and Charlie

	s1, _ := studentStore.SaveStudent(student1)
	s2, _ := studentStore.SaveStudent(student2)
	s3, _ := studentStore.SaveStudent(student3)

	expectedOrder := []models.Student{s2, s3, s1} // Sorted by name: Alice, Bob, Charlie

	students, err = studentStore.ListStudents()
	if err != nil {
		t.Fatalf("ListStudents failed: %v", err)
	}
	if len(students) != 3 {
		t.Errorf("Expected 3 students, got %d", len(students))
	}

	// Verify sorting (ListStudents should return them sorted by name ASC)
	correctOrder := true
	for i := 0; i < len(students); i++ {
		if students[i].ID != expectedOrder[i].ID || students[i].Name != expectedOrder[i].Name {
			correctOrder = false
			break
		}
	}
	if !correctOrder {
		var gotNames []string
		for _, s := range students {
			gotNames = append(gotNames, s.Name)
		}
		var expectedNames []string
		for _, s := range expectedOrder {
			expectedNames = append(expectedNames, s.Name)
		}
		t.Errorf("ListStudents did not return students in expected sorted order by name.\nExpected: %v\nGot: %v", expectedNames, gotNames)
	}
}

func TestStudentStore_SaveStudent_UpdateExisting(t *testing.T) {
	_, studentStore, teardown := setupStudentDB(t)
	defer teardown()

	student := models.Student{Name: "Initial Name"}
	savedStudent, err := studentStore.SaveStudent(student)
	if err != nil {
		t.Fatalf("Initial SaveStudent failed: %v", err)
	}

	updatedStudent := savedStudent
	updatedStudent.Name = "Updated Name"

	_, err = studentStore.SaveStudent(updatedStudent) // SaveStudent uses INSERT OR REPLACE
	if err != nil {
		t.Fatalf("Updating student with SaveStudent failed: %v", err)
	}

	retrievedStudent, err := studentStore.GetStudentByID(savedStudent.ID)
	if err != nil {
		t.Fatalf("GetStudentByID after update failed: %v", err)
	}
	if retrievedStudent.Name != "Updated Name" {
		t.Errorf("Expected student name to be 'Updated Name' after update, got '%s'", retrievedStudent.Name)
	}
}

func TestStudentStore_SaveStudent_IDGeneration(t *testing.T) {
	_, studentStore, teardown := setupStudentDB(t)
	defer teardown()

	studentWithoutID := models.Student{Name: "No ID Given"}
	savedStudent1, err := studentStore.SaveStudent(studentWithoutID)
	if err != nil {
		t.Fatalf("SaveStudent for studentWithoutID failed: %v", err)
	}
	if savedStudent1.ID == "" {
		t.Errorf("Expected ID to be generated for studentWithoutID, but it's empty")
	}
	_, err = uuid.Parse(savedStudent1.ID)
	if err != nil {
		t.Errorf("Expected generated ID to be a UUID for studentWithoutID, but parsing failed: %v", err)
	}

	predefinedID := uuid.NewString()
	studentWithID := models.Student{
		ID:   predefinedID,
		Name: "Has ID Given",
	}
	savedStudent2, err := studentStore.SaveStudent(studentWithID)
	if err != nil {
		t.Fatalf("SaveStudent for studentWithID failed: %v", err)
	}
	if savedStudent2.ID != predefinedID {
		t.Errorf("Expected ID to be '%s' for studentWithID, but got '%s'", predefinedID, savedStudent2.ID)
	}
}

// TestMain can be used for global setup/teardown if needed
// func TestMain(m *testing.M) { // Removed to avoid multiple TestMain definitions
// 	os.Exit(m.Run())
// }

// Explicitly import sort for clarity, though it's used via sort.Slice in ListStudents test verification.
// The student store implementation itself uses `ORDER BY name ASC` in SQL.
var _ = sort.Interface(nil) // Use sort to satisfy linter if no direct calls in test.
// Actually, the sorting check is manual.
// The `store.ListStudents` is expected to return sorted results.
// Let's remove the unused sort import if not directly used in test logic.
// The test logic above manually defines `expectedOrder` and compares.
// No, `sort` is not directly used in the test logic itself. The SQL query handles sorting.
// The `strings` import was added due to `strings.Contains`
// The `github.com/google/uuid` import was added due to `uuid.Parse` and `uuid.NewString`
