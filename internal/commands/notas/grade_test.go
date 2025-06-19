package notas

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"vickgenda/internal/models"
	"vickgenda/internal/store"
)

// setupGradeAndDependenciesTests initializes tables and clears data for a clean test environment.
// It now also handles the grades table.
func setupGradeAndDependenciesTests(t *testing.T) {
	// Reset package-level ID counters for predictable test IDs
	ResetGradeIDCounterForTesting() // Resets notas.currentGradeID
	nextTermID = 1                  // Resets notas.nextTermID (package notas)

	// Initialize database tables
	err := store.InitStudentsTable()
	if err != nil { t.Fatalf("Failed to initialize students table: %v", err) }
	err = store.InitTermsTable()
	if err != nil { t.Fatalf("Failed to initialize terms table: %v", err) }
	err = store.InitGradesTable() // Initialize Grades table
	if err != nil { t.Fatalf("Failed to initialize grades table: %v", err) }

	// Clear data from tables
	// Order matters if there are foreign key constraints without ON DELETE CASCADE,
	// but with ON DELETE CASCADE (as in grades->students/terms), order is less critical.
	// Clearing grades first, then terms, then students is safest if unsure.
	err = store.ClearGradesTableForTesting() // Clear Grades table first
	if err != nil { t.Fatalf("Failed to clear grades table: %v", err) }
	err = store.ClearTermsTableForTesting()
	if err != nil { t.Fatalf("Failed to clear terms table: %v", err) }
	err = store.ClearStudentsTableForTesting()
	if err != nil { t.Fatalf("Failed to clear students table: %v", err) }
}

// setupGradeTests remains largely the same but benefits from the more comprehensive setup.
func setupGradeTests(t *testing.T) (studentID string, studentID2 string, termID string) {
	setupGradeAndDependenciesTests(t) // This now also handles grades table init/clear

	student1 := models.Student{ID: "stud001", Name: "Aluno Teste Principal"}
	if err := store.CreateStudent(student1); err != nil {
		t.Fatalf("setupGradeTests: Failed to create student %s: %v", student1.ID, err)
	}

	student2Model := models.Student{ID: "stud002", Name: "Aluno Teste Secundário"}
	if err := store.CreateStudent(student2Model); err != nil {
		t.Fatalf("setupGradeTests: Failed to create student %s: %v", student2Model.ID, err)
	}

	// ConfigurarBimestreAdicionar uses nextTermID, which is reset by setupGradeAndDependenciesTests.
	term, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre Teste Global", "01-02-2024", "15-04-2024") // Uses term001
	if err != nil {
		t.Fatalf("setupGradeTests: Falha ao criar termo mock: %v", err)
	}
	return student1.ID, student2Model.ID, term.ID
}

func TestLancarNota(t *testing.T) {
	studentIDForTest, _, termIDForTest := setupGradeTests(t) // stud001, term001

	tests := []struct {
		name            string
		alunoID         string
		bimestreID      string
		disciplina      string
		avaliacaoDesc   string
		valorNota       float64
		pesoNota        float64
		dataStr         string
		setupStudentForSubtest *models.Student
		expectError     bool
		expectedValue   float64
		expectedGradeID string // generateGradeID still produces predictable IDs
	}{
		{"nota válida", studentIDForTest, termIDForTest, "Matemática", "Prova 1", 8.5, 0.4, "10-03-2024", nil, false, 8.5, "nota001"},
		{"aluno inexistente", "stud999", termIDForTest, "Matemática", "Prova 1", 8.0, 1, "10-03-2024", nil, true, 0, ""},
		{"bimestre inexistente", studentIDForTest, "term999", "Matemática", "Prova 1", 8.0, 1, "10-03-2024", nil, true, 0, ""},
		// Add more specific validation error cases if needed (e.g. valorNota out of range)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ResetGradeIDCounterForTesting is called by setupGradeAndDependenciesTests,
			// but call it again if subtests need truly independent ID generation starting from nota001.
			// And clear grades table for isolation of this specific LancarNota call.
			ResetGradeIDCounterForTesting()
			if err := store.ClearGradesTableForTesting(); err != nil {
				t.Fatalf("Subtest %s: Failed to clear grades table: %v", tt.name, err)
			}

			currentStudentID := tt.alunoID
			if tt.setupStudentForSubtest != nil {
				if err := store.CreateStudent(*tt.setupStudentForSubtest); err != nil {
					// This might fail if student ID conflicts with global setup. Test design should ensure uniqueness.
					t.Fatalf("Subtest %s: Failed to create specific student %s: %v", tt.name, tt.setupStudentForSubtest.ID, err)
				}
				currentStudentID = tt.setupStudentForSubtest.ID
			}

			grade, err := LancarNota(currentStudentID, tt.bimestreID, tt.disciplina, tt.avaliacaoDesc, tt.valorNota, tt.pesoNota, tt.dataStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Subtest %s: esperado erro, mas obteve nil", tt.name)
				}
			} else {
				if err != nil {
					t.Fatalf("Subtest %s: esperado sem erro, mas obteve: %v", tt.name, err)
				}
				if grade.ID != tt.expectedGradeID {
				    t.Errorf("Subtest %s: esperado ID da nota '%s', mas obteve '%s'", tt.name, tt.expectedGradeID, grade.ID)
				}
				if !floatEquals(grade.Value, tt.expectedValue) { // Use floatEquals for float comparison
					t.Errorf("Subtest %s: esperado valor da nota %.2f, obteve %.2f", tt.name, tt.expectedValue, grade.Value)
				}

				// Verify the grade was actually persisted in the database
				persistedGrade, dbErr := store.GetGradeByID(grade.ID)
				if dbErr != nil {
					t.Errorf("Subtest %s: Nota lançada não foi encontrada no DB: %v", tt.name, dbErr)
				}
				// Normalize time for comparison if necessary, as DB might store with different precision/timezone
                // For simplicity, if dates are set via time.Parse with same layout, they should be comparable.
				if !reflect.DeepEqual(persistedGrade, grade) {
					t.Errorf("Subtest %s: Nota persistida no DB não corresponde à retornada.\nDB:  %+v\nRet: %+v", tt.name, persistedGrade, grade)
				}
			}
		})
	}
}

// TestVerNotas remains largely the same as it already uses VerNotas, which now reads from DB.
// The setup of initial grades via LancarNota will now persist to DB.
func TestVerNotas(t *testing.T) {
	studentID, otherStudentID, termID := setupGradeTests(t) // stud001, stud002, term001
    otherTerm, _ := ConfigurarBimestreAdicionar(2024, "2º Bimestre TVN", "16-04-2024", "30-06-2024") // term002
    otherTermID := otherTerm.ID

	ResetGradeIDCounterForTesting() // Reset before this block of LancarNota calls
	nota1, _ := LancarNota(studentID, termID, "Matemática", "P1 M", 8.0, 0.4, "10-03-2024")     // nota001
	nota2, _ := LancarNota(studentID, termID, "Matemática", "T1 M", 7.0, 0.3, "15-03-2024")     // nota002
	nota3, _ := LancarNota(studentID, termID, "Português", "R1 P", 9.0, 0.5, "12-03-2024")     // nota003
    nota4, _ := LancarNota(studentID, otherTermID, "Matemática", "P2 M OT", 8.5, 0.4, "05-05-2024") // nota004
    LancarNota(otherStudentID, termID, "Matemática", "P1 M OS", 6.0, 0.4, "10-03-2024")         // nota005

	tests := []struct {
		name             string; alunoID string; bimestreIDFilter string; disciplinaFilter string; expectError bool; expectedCount int; expectedGrades []models.Grade
	}{
		{"todas do aluno1 no bim1", studentID, termID, "", false, 3, []models.Grade{nota1, nota3, nota2}}, // Sorted by date by ListGrades
		{"matemática do aluno1 no bim1", studentID, termID, "Matemática", false, 2, []models.Grade{nota1, nota2}},
		{"português do aluno1 no bim1", studentID, termID, "Português", false, 1, []models.Grade{nota3}},
		{"todas do aluno1 (todos bimestres)", studentID, "", "", false, 4, []models.Grade{nota1, nota3, nota2, nota4}},
		{"aluno2 sem notas no bim2", otherStudentID, otherTermID, "", false, 0, []models.Grade{}},
		{"aluno inexistente", "stud999", "", "", true, 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grades, err := VerNotas(tt.alunoID, tt.bimestreIDFilter, tt.disciplinaFilter)
			if tt.expectError { if err == nil { t.Errorf("esperado erro") }
			} else {
				if err != nil { t.Fatalf("sem erro esperado, obteve: %v", err) }
				if len(grades) != tt.expectedCount { t.Errorf("esperado %d, obteve %d. Notas: %+v", tt.expectedCount, len(grades), grades) }
				// VerNotas now gets data from store.ListGrades, which sorts by date.
				// Ensure tt.expectedGrades are also sorted by date if DeepEqual is to succeed.
                // The test data for expectedGrades is already ordered by date.
				if tt.expectedCount > 0 && !reflect.DeepEqual(grades, tt.expectedGrades) {
					t.Errorf("lista não corresponde.\nEsperado: %+v\nObtido:   %+v", tt.expectedGrades, grades)
				}
			}
		})
	}
}


const floatEqualityThreshold = 1e-9
func floatEquals(a, b float64) bool { return math.Abs(a-b) < floatEqualityThreshold }

// TestCalcularMedia also remains largely the same in structure.
func TestCalcularMedia(t *testing.T) {
    studentID, stud002ID, termID := setupGradeTests(t)
	ResetGradeIDCounterForTesting()
    n1Mat, _ := LancarNota(studentID, termID, "Matemática", "P1", 8.0, 0.4, "10-03-2024")
    n2Mat, _ := LancarNota(studentID, termID, "Matemática", "T1", 7.0, 0.3, "15-03-2024")
    n3Mat, _ := LancarNota(studentID, termID, "Matemática", "Part", 10.0, 0.3, "20-03-2024")
    LancarNota(stud002ID, termID, "Matemática", "P.Única", 5.0, 1.0, "10-03-2024")

    tests := []struct { name string; alunoID string; bimestreID string; disciplina string; expectError bool; expectedMedia float64; expectedSomaPesos float64; expectedNotasCount int }{
        {"média mat aluno1", studentID, termID, "Matemática", false, 8.3, 1.0, 3},
        {"aluno inexistente", "stud999", termID, "Matemática", true, 0, 0, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mediaInfo, err := CalcularMedia(tt.alunoID, tt.bimestreID, tt.disciplina)
            if tt.expectError { if err == nil { t.Errorf("esperado erro") } } else {
                if err != nil { t.Fatalf("sem erro esperado: %v", err) }
                if !floatEquals(mediaInfo.MediaPonderada, tt.expectedMedia) { t.Errorf("média esperada %.2f, obteve %.2f", tt.expectedMedia, mediaInfo.MediaPonderada) }
                if len(mediaInfo.NotasConsideradas) != tt.expectedNotasCount {t.Errorf("esperado %d notas, obteve %d", tt.expectedNotasCount, len(mediaInfo.NotasConsideradas))}
            }
        })
    }
}


func float64Ptr(v float64) *float64 { return &v }
func stringPtr(v string) *string    { return &v }

func TestEditarNota(t *testing.T) {
    baseStudentID, _, baseTermID := setupGradeTests(t) // stud001, term001

    // Helper to create a fresh note for each subtest, ensuring isolation
    createTestNoteForEdit := func(t *testing.T) models.Grade {
        ResetGradeIDCounterForTesting() // Reset ID for nota001
		if err := store.ClearGradesTableForTesting(); err != nil {t.Fatalf("ClearGrades in createTestNoteForEdit: %v",err)}
        nota, err := LancarNota(baseStudentID, baseTermID, "Ciências Edit", "Original Desc", 7.5, 0.5, "20-03-2024") // Creates nota001
        if err != nil {
            t.Fatalf("Failed to create test note for edit: %v", err)
        }
        return nota
    }

    tests := []struct {
        name string; novoValor *float64; novoPeso *float64; novaDesc *string; novaDataStr *string; expectError bool; expectedValue float64; expectedWeight float64; expectedDesc string; expectedDateStr string
    }{
        {"editar todos os campos", float64Ptr(8.0), float64Ptr(0.6), stringPtr("Desc Atualizada"), stringPtr("22-03-2024"), false, 8.0, 0.6, "Desc Atualizada", "22-03-2024"},
        {"editar apenas valor", float64Ptr(9.0), nil, nil, nil, false, 9.0, 0.5, "Original Desc", "20-03-2024"},
        {"nota não encontrada (ID errado)", float64Ptr(5.0), nil, nil, nil, true, 0,0,"",""}, // expected fields don't matter on error
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            originalNota := createTestNoteForEdit(t)
            idDaNotaParaTeste := originalNota.ID
            if tt.name == "nota não encontrada (ID errado)" {
                idDaNotaParaTeste = "nota_fantasma_para_editar"
            }

			// Determine expected values based on what's being changed
            expectedVal := originalNota.Value
			if tt.novoValor != nil { expectedVal = tt.expectedValue }
			expectedWgt := originalNota.Weight
			if tt.novoPeso != nil { expectedWgt = tt.expectedWeight }
			expectedDsc := originalNota.Description
			if tt.novaDesc != nil { expectedDsc = tt.expectedDesc }
			expectedDtStr := originalNota.Date.Format("02-01-2006")
			if tt.novaDataStr != nil { expectedDtStr = tt.expectedDateStr }

            _, err := EditarNota(idDaNotaParaTeste, tt.novoValor, tt.novoPeso, tt.novaDesc, tt.novaDataStr)

            if tt.expectError {
                if err == nil { t.Errorf("esperado erro, mas obteve nil. Teste: %s", tt.name) }
            } else {
                if err != nil { t.Fatalf("sem erro esperado, mas obteve: %v. Teste: %s", err, tt.name) }

                persistedGrade, dbErr := store.GetGradeByID(idDaNotaParaTeste)
                if dbErr != nil { t.Fatalf("Erro ao buscar nota editada do DB: %v. Teste: %s", dbErr, tt.name) }

                if !floatEquals(persistedGrade.Value, expectedVal) { t.Errorf("DB valor %.2f vs esperado %.2f. Teste: %s", persistedGrade.Value, expectedVal, tt.name) }
                if !floatEquals(persistedGrade.Weight, expectedWgt) { t.Errorf("DB peso %.2f vs esperado %.2f. Teste: %s", persistedGrade.Weight, expectedWgt, tt.name) }
                if persistedGrade.Description != expectedDsc { t.Errorf("DB desc '%s' vs esperado '%s'. Teste: %s", persistedGrade.Description, expectedDsc, tt.name) }
                parsedExpectedDate, _ := time.Parse("02-01-2006", expectedDtStr)
                if !persistedGrade.Date.Equal(parsedExpectedDate) { // Direct time.Time comparison after parsing
                     t.Errorf("DB data '%s' vs esperado '%s'. Teste: %s", persistedGrade.Date.Format("02-01-2006"), expectedDtStr, tt.name)
                }
            }
        })
    }
}

func TestExcluirNota(t *testing.T) {
    studentID, _, termID := setupGradeTests(t)
	ResetGradeIDCounterForTesting() // nota001, nota002
    notaParaExcluir, _ := LancarNota(studentID, termID, "Geografia", "Mapa Para Excluir", 9.0, 1.0, "01-04-2024")
    LancarNota(studentID, termID, "Geografia", "Capitais Para Manter", 8.0, 1.0, "02-04-2024")

    t.Run("excluir nota existente", func(t *testing.T) {
        err := ExcluirNota(notaParaExcluir.ID)
        if err != nil { t.Fatalf("Erro ao excluir nota: %v", err) }

        _, errGet := store.GetGradeByID(notaParaExcluir.ID)
        if errGet == nil { // Should be an error (sql.ErrNoRows or similar wrapped)
            t.Errorf("Nota '%s' ainda encontrada no DB após exclusão, mas deveria ter sido removida.", notaParaExcluir.ID)
        }
    })

    t.Run("excluir nota inexistente", func(t *testing.T) {
        err := ExcluirNota("nota_fantasma_para_excluir")
        if err == nil {
            t.Error("Esperado erro ao excluir nota inexistente, mas não ocorreu")
        }
    })
}

```
