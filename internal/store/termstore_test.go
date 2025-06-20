package store_test

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3
	"vickgenda/internal/models"
	"vickgenda/internal/store"
)

// setupTermDB initializes an in-memory SQLite database and a TermStore for testing.
// It returns the database connection, the TermStore, and a teardown function.
func setupTermDB(t *testing.T) (*sql.DB, store.TermStore, func()) {
	t.Helper()

	// Using file::memory:?cache=shared to ensure the connection is shareable if needed,
	// though for simple tests, ":memory:" is often fine.
	// Using a temporary file-based DB can also be useful for inspection.
	// For this test, :memory: should be sufficient.
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	termStore := store.NewSQLiteTermStore(db)
	if err := termStore.Init(); err != nil {
		db.Close()
		t.Fatalf("Failed to initialize term store: %v", err)
	}

	teardown := func() {
		db.Close()
	}

	return db, termStore, teardown
}

func parseDate(t *testing.T, dateStr string) time.Time {
	t.Helper()
	layout := "2006-01-02"
	tm, err := time.Parse(layout, dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date string '%s': %v", dateStr, err)
	}
	return tm
}

func TestTermStore_SaveAndGetTerm(t *testing.T) {
	_, termStore, teardown := setupTermDB(t)
	defer teardown()

	term := models.Term{
		Name:      "1º Bimestre",
		StartDate: parseDate(t, "2024-02-01"),
		EndDate:   parseDate(t, "2024-04-15"),
	}

	savedTerm, err := termStore.SaveTerm(term)
	if err != nil {
		t.Fatalf("SaveTerm failed: %v", err)
	}
	if savedTerm.ID == "" {
		t.Errorf("Expected saved term to have an ID, got empty string")
	}

	retrievedTerm, err := termStore.GetTermByID(savedTerm.ID)
	if err != nil {
		t.Fatalf("GetTermByID failed: %v", err)
	}
	if retrievedTerm.Name != term.Name || !retrievedTerm.StartDate.Equal(term.StartDate) || !retrievedTerm.EndDate.Equal(term.EndDate) {
		t.Errorf("Retrieved term does not match saved term. Got %+v, expected %+v (ignoring ID)", retrievedTerm, term)
	}

	_, err = termStore.GetTermByID("non-existent-id")
	if err == nil {
		t.Errorf("Expected error when getting non-existent term, got nil")
	} else {
		// Check if the error is or wraps sql.ErrNoRows, or contains "not found"
		// The exact error check depends on how the store wraps it.
		// For SQLiteTermStore, it's fmt.Errorf("term with ID '%s' not found: %w", id, err)
		if !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected sql.ErrNoRows or 'not found' error, got: %v", err)
		}
	}
}

func TestTermStore_ListTermsByYear(t *testing.T) {
	_, termStore, teardown := setupTermDB(t)
	defer teardown()

	term2024_1 := models.Term{Name: "1º Bim 2024", StartDate: parseDate(t, "2024-02-01"), EndDate: parseDate(t, "2024-04-15")}
	term2024_2 := models.Term{Name: "2º Bim 2024", StartDate: parseDate(t, "2024-05-01"), EndDate: parseDate(t, "2024-07-15")}
	term2023_1 := models.Term{Name: "1º Bim 2023", StartDate: parseDate(t, "2023-02-01"), EndDate: parseDate(t, "2023-04-15")}

	_, _ = termStore.SaveTerm(term2024_1)
	_, _ = termStore.SaveTerm(term2024_2)
	_, _ = termStore.SaveTerm(term2023_1)

	terms2024, err := termStore.ListTermsByYear(2024)
	if err != nil {
		t.Fatalf("ListTermsByYear(2024) failed: %v", err)
	}
	if len(terms2024) != 2 {
		t.Errorf("Expected 2 terms for year 2024, got %d", len(terms2024))
	}
	// Basic check for names, could be more thorough
	found2024_1 := false
	found2024_2 := false
	for _, term := range terms2024 {
		if term.Name == term2024_1.Name {
			found2024_1 = true
		}
		if term.Name == term2024_2.Name {
			found2024_2 = true
		}
	}
	if !found2024_1 || !found2024_2 {
		t.Errorf("Did not find all expected terms for 2024. Found1: %v, Found2: %v", found2024_1, found2024_2)
	}
    // Check order (by StartDate ASC)
    if len(terms2024) == 2 && terms2024[0].StartDate.After(terms2024[1].StartDate) {
        t.Errorf("Terms for 2024 are not sorted by StartDate ASC. Got %s then %s", terms2024[0].Name, terms2024[1].Name)
    }


	terms2025, err := termStore.ListTermsByYear(2025)
	if err != nil {
		t.Fatalf("ListTermsByYear(2025) failed: %v", err)
	}
	if len(terms2025) != 0 {
		t.Errorf("Expected 0 terms for year 2025, got %d", len(terms2025))
	}
}

func TestTermStore_TermOverlapValidation(t *testing.T) {
	_, termStore, teardown := setupTermDB(t)
	defer teardown()

	initialTerm := models.Term{Name: "Initial Term", StartDate: parseDate(t, "2024-03-01"), EndDate: parseDate(t, "2024-04-15")}
	savedInitialTerm, err := termStore.SaveTerm(initialTerm)
	if err != nil {
		t.Fatalf("Failed to save initial term: %v", err)
	}

	overlapTests := []struct {
		name        string
		term        models.Term
		expectError bool
	}{
		{"OverlapEnd", models.Term{Name: "Overlap End", StartDate: parseDate(t, "2024-04-01"), EndDate: parseDate(t, "2024-05-15")}, true},
		{"OverlapStart", models.Term{Name: "Overlap Start", StartDate: parseDate(t, "2024-02-01"), EndDate: parseDate(t, "2024-03-15")}, true},
		{"OverlapMiddle", models.Term{Name: "Overlap Middle", StartDate: parseDate(t, "2024-03-15"), EndDate: parseDate(t, "2024-04-01")}, true},
		{"FullyContained", models.Term{Name: "Fully Contained", StartDate: parseDate(t, "2024-03-05"), EndDate: parseDate(t, "2024-04-01")}, true},
		{"ContainsInitial", models.Term{Name: "Contains Initial", StartDate: parseDate(t, "2024-02-15"), EndDate: parseDate(t, "2024-04-30")}, true},
		{"NonOverlappingSameYear", models.Term{Name: "Non-Overlap Same Year", StartDate: parseDate(t, "2024-05-01"), EndDate: parseDate(t, "2024-06-15")}, false},
		{"OverlappingDifferentYear", models.Term{Name: "Overlap Diff Year", StartDate: parseDate(t, "2025-03-01"), EndDate: parseDate(t, "2025-04-15")}, false},
        {"ExactSameDatesSameYear", models.Term{Name: "Exact Dates Same Year Diff Name", StartDate: parseDate(t, "2024-03-01"), EndDate: parseDate(t, "2024-04-15")}, true},
	}

	for _, tc := range overlapTests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := termStore.SaveTerm(tc.term)
			if tc.expectError && err == nil {
				t.Errorf("Expected overlap error for term '%s', but got nil", tc.term.Name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for term '%s', but got: %v", tc.term.Name, err)
			}
			if tc.expectError && err != nil && !strings.Contains(err.Error(), "overlaps with existing term") {
				t.Errorf("Expected error message to contain 'overlaps with existing term', got: %v", err.Error())
			}
		})
	}

	// Test updating the same term (should not cause overlap with itself)
	updateInitialTerm := savedInitialTerm
	updateInitialTerm.Name = "Initial Term Updated"
	_, err = termStore.SaveTerm(updateInitialTerm)
	if err != nil {
		t.Errorf("Updating the same term should not result in an overlap error, but got: %v", err)
	}
}

func TestTermStore_SaveTerm_IDGeneration(t *testing.T) {
	_, termStore, teardown := setupTermDB(t)
	defer teardown()

	termWithoutID := models.Term{
		Name:      "Term No ID",
		StartDate: parseDate(t, "2024-01-01"),
		EndDate:   parseDate(t, "2024-01-31"),
	}
	savedTerm1, err := termStore.SaveTerm(termWithoutID)
	if err != nil {
		t.Fatalf("SaveTerm for termWithoutID failed: %v", err)
	}
	if savedTerm1.ID == "" {
		t.Errorf("Expected ID to be generated for termWithoutID, but it's empty")
	}
	_, err = uuid.Parse(savedTerm1.ID)
	if err != nil {
		t.Errorf("Expected generated ID to be a UUID, but parsing failed: %v", err)
	}

	predefinedID := uuid.NewString()
	termWithID := models.Term{
		ID:        predefinedID,
		Name:      "Term With ID",
		StartDate: parseDate(t, "2024-02-01"),
		EndDate:   parseDate(t, "2024-02-28"),
	}
	savedTerm2, err := termStore.SaveTerm(termWithID)
	if err != nil {
		t.Fatalf("SaveTerm for termWithID failed: %v", err)
	}
	if savedTerm2.ID != predefinedID {
		t.Errorf("Expected ID to be '%s', but got '%s'", predefinedID, savedTerm2.ID)
	}
}

// Helper function to check if a string is in a slice of strings (for testing purposes)
// This might be needed if ListTermsByYear doesn't guarantee order or if we only check a subset.
// However, ListTermsByYear *should* guarantee order by start_date.
func containsTermName(terms []models.Term, name string) bool {
	for _, term := range terms {
		if term.Name == name {
			return true
		}
	}
	return false
}

// Add a TestMain to manage global setup/teardown if necessary,
// e.g., for creating a temporary DB file instead of :memory: for inspection.
func TestMain(m *testing.M) {
	// Setup code here, if any
	exitCode := m.Run()
	// Teardown code here, if any
	os.Exit(exitCode)
}

// It's good practice to import "strings" when using strings.Contains
// Even if the linter doesn't complain, it makes dependencies explicit.
// This was missing in the initial prompt but added here for completeness.
import "strings"
