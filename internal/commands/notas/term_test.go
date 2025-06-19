package notas

import (
	"reflect"
	"testing"
	"time" // Required for parsing time in test setup if needed

	"vickgenda/internal/models"
	"vickgenda/internal/store" // Required for Init and Clear
)

// Helper function to setup for each test or group of tests
func setupTermTestTable(t *testing.T) {
	nextTermID = 1 // Reset package-level ID counter for predictable IDs
	// Ensure db path is set for testing, e.g., to an in-memory db
	// This might be done in a TestMain or globally, but for safety, ensure db is usable.
	// Example: db.SetDBPath(":memory:") or a test-specific file.
	// For now, assume db.GetDB() will work or is configured elsewhere.
	err := store.InitTermsTable() // Ensure table exists
	if err != nil {
		t.Fatalf("Failed to initialize terms table for test: %v", err)
	}
	err = store.ClearTermsTableForTesting() // Clear data before each test
	if err != nil {
		t.Fatalf("Failed to clear terms table for test: %v", err)
	}
}

func TestConfigurarBimestreAdicionar(t *testing.T) {
	tests := []struct {
		name            string
		anoLetivo       int
		nome            string
		inicioStr       string
		fimStr          string
		setupPreexisting func(t *testing.T) // Function to setup pre-existing conflicting data
		expectError     bool
		expectedTermName string
		expectedID      string // For checking predictable ID
	}{
		{
			name: "bimestre válido", anoLetivo: 2024, nome: "1º Bimestre", inicioStr: "01-02-2024", fimStr: "15-04-2024",
			expectError: false, expectedTermName: "1º Bimestre", expectedID: "term001",
		},
		{
			name: "data fim antes de inicio", anoLetivo: 2024, nome: "Inválido", inicioStr: "15-04-2024", fimStr: "01-02-2024",
			expectError: true,
		},
		{
			name: "campos obrigatórios faltando", anoLetivo: 2024, nome: "", inicioStr: "01-02-2024", fimStr: "15-04-2024",
			expectError: true,
		},
		{
			name: "formato data inválido", anoLetivo: 2024, nome: "Bimestre X", inicioStr: "2024/02/01", fimStr: "2024/04/15",
			expectError: true,
		},
		{
			name: "adicionar segundo bimestre válido após o primeiro", anoLetivo: 2024, nome: "2º Bimestre", inicioStr: "16-04-2024", fimStr: "30-06-2024",
			setupPreexisting: func(t *testing.T) {
				// nextTermID is 1 initially in this sub-test due to setupTermTestTable
				// It will be consumed by the pre-existing term.
				_, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024") // Consumes term001
				if err != nil {
					t.Fatalf("Setup error for 'adicionar segundo bimestre': %v", err)
				}
				// nextTermID is now 2 for the term being tested in this specific sub-test.
			},
			expectError: false, expectedTermName: "2º Bimestre", expectedID: "term002",
		},
		{
			name: "tentar adicionar bimestre com mesmo nome no mesmo ano", anoLetivo: 2024, nome: "1º Bimestre", inicioStr: "01-05-2024", fimStr: "15-07-2024",
			setupPreexisting: func(t *testing.T) {
				_, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024")
				if err != nil {
					t.Fatalf("Setup error for 'mesmo nome no mesmo ano': %v", err)
				}
			},
			expectError: true, // DB unique constraint (name, year) should trigger this
		},
        {
			// This test's expectation changes. With the previous in-memory logic, this might have been an error
			// if the validation was strict about any date overlap.
			// However, the DB constraint is UNIQUE(name, strftime('%Y', start_date)).
			// It does NOT prevent overlapping date ranges if the names are different.
			// The Go code for ConfigurarBimestreAdicionar now primarily relies on the DB for this.
			// If strict overlap prevention for different names is desired, it needs to be re-added as a Go check.
			// For now, this test reflects that different names can have overlapping dates.
			name: "bimestre com sobreposição de datas nomes diferentes", anoLetivo: 2024, nome: "Sobreposto", inicioStr: "10-04-2024", fimStr: "20-04-2024",
			setupPreexisting: func(t *testing.T) {
				_, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024") // term001
                if err != nil { t.Fatalf("Setup error for 'sobreposição datas nomes diferentes': %v", err) }
			},
			expectError: false, expectedTermName: "Sobreposto", expectedID: "term002", // Will be term002 as 1st Bimestre was term001
		},
        {
			name: "bimestre válido em outro ano", anoLetivo: 2025, nome: "1º Bimestre", inicioStr: "01-02-2025", fimStr: "15-04-2025",
			// No setupPreexisting needed as setupTermTestTable clears and it's a different year.
			expectError: false, expectedTermName: "1º Bimestre", expectedID: "term001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTermTestTable(t) // Clear, Init table, and reset nextTermID for each sub-test

			if tt.setupPreexisting != nil {
				tt.setupPreexisting(t) // This will affect nextTermID for the actual call below
			}

			term, err := ConfigurarBimestreAdicionar(tt.anoLetivo, tt.nome, tt.inicioStr, tt.fimStr)
			if tt.expectError {
				if err == nil {
					t.Errorf("esperado erro, mas obteve nil. Teste: %s", tt.name)
				}
			} else {
				if err != nil {
					t.Fatalf("esperado sem erro, mas obteve: %v. Teste: %s", err, tt.name)
				}
				if term.ID == "" {
					t.Errorf("esperado ID do termo preenchido. Teste: %s", tt.name)
				}
                if term.ID != tt.expectedID {
					t.Errorf("esperado ID do termo '%s', obteve '%s'. Teste: %s", tt.expectedID, term.ID, tt.name)
				}
				if term.Name != tt.expectedTermName {
					t.Errorf("esperado nome do termo '%s', obteve '%s'. Teste: %s", tt.expectedTermName, term.Name, tt.name)
				}
				// Verify it's in the DB
				_, dbErr := store.GetTermByID(term.ID)
				if dbErr != nil {
					t.Errorf("Termo adicionado não encontrado no DB: %v. Teste: %s", dbErr, tt.name)
				}
			}
		})
	}
}

func TestConfigurarBimestreListar(t *testing.T) {
    setupTermTestTable(t) // Clear, Init table, and reset nextTermID

    // Add terms directly using the function being tested or store.CreateTerm for setup.
    // Using ConfigurarBimestreAdicionar ensures IDs are generated as expected by tests.
    term1_2024, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024") // term001
    if err != nil { t.Fatalf("TestConfigurarBimestreListar setup error term1_2024: %v", err) }
    term2_2024, err := ConfigurarBimestreAdicionar(2024, "2º Bimestre", "16-04-2024", "30-06-2024") // term002
    if err != nil { t.Fatalf("TestConfigurarBimestreListar setup error term2_2024: %v", err) }
    term1_2025, err := ConfigurarBimestreAdicionar(2025, "1º Bimestre", "01-02-2025", "15-04-2025") // term003
    if err != nil { t.Fatalf("TestConfigurarBimestreListar setup error term1_2025: %v", err) }

    tests := []struct {
        name          string
        anoLetivo     int
        expectError   bool
        expectedCount int
        expectedTerms []models.Term
    }{
        // Note: The store.ListTermsByYear already sorts by start_date.
        {"listar bimestres 2024", 2024, false, 2, []models.Term{term1_2024, term2_2024}},
        {"listar bimestres 2025", 2025, false, 1, []models.Term{term1_2025}},
        {"listar ano sem bimestres", 2026, false, 0, []models.Term{}},
        {"listar com ano 0 (erro)", 0, true, 0, nil},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            terms, err := ConfigurarBimestreListar(tt.anoLetivo)
            if tt.expectError {
                if err == nil {
                    t.Errorf("esperado erro, mas obteve nil. Teste: %s", tt.name)
                }
            } else {
                if err != nil {
                    t.Fatalf("esperado sem erro, mas obteve: %v. Teste: %s", err, tt.name)
                }
                if len(terms) != tt.expectedCount {
                    t.Errorf("esperado %d termos, obteve %d. Teste: %s", tt.expectedCount, len(terms), tt.name)
                }
                // DeepEqual checks order as well.
                if tt.expectedCount > 0 { // Only compare if we expect terms
                    if !reflect.DeepEqual(terms, tt.expectedTerms) {
                        t.Errorf("lista de termos não corresponde ao esperado. Teste: %s\nEsperado: %+v\nObtido:   %+v", tt.name, tt.expectedTerms, terms)
                    }
                } else if len(terms) != 0 { // If expectedCount is 0, terms should be empty
                     t.Errorf("esperado lista vazia, obteve %d termos. Teste: %s", len(terms), tt.name)
                }
            }
        })
    }
}

func TestGetTermByID(t *testing.T) {
    setupTermTestTable(t) // Clear, Init table, and reset nextTermID

    // ID will be term001 due to reset in setupTermTestTable
    addedTerm, err := ConfigurarBimestreAdicionar(2024, "Único Bimestre", "01-03-2024", "30-04-2024")
    if err != nil { t.Fatalf("TestGetTermByID setup error: %v", err) }

    t.Run("buscar termo existente", func(t *testing.T) {
        term, err := GetTermByID(addedTerm.ID) // addedTerm.ID should be "term001"
        if err != nil {
            t.Fatalf("Erro ao buscar termo existente: %v", err)
        }
        if term.ID != addedTerm.ID || term.Name != addedTerm.Name {
            // For more detailed comparison, especially with time.Time
            if !reflect.DeepEqual(term, addedTerm) {
                 t.Errorf("Termo recuperado não corresponde ao adicionado.\nEsperado: %+v\nObtido:   %+v", addedTerm, term)
            }
        }
    })

    t.Run("buscar termo inexistente", func(t *testing.T) {
        _, err := GetTermByID("term_nao_existe")
        if err == nil {
            t.Error("Esperado erro ao buscar termo inexistente, mas obteve nil")
        }
    })

    t.Run("buscar termo com ID vazio", func(t *testing.T) {
        _, err := GetTermByID("")
        if err == nil {
            t.Error("Esperado erro ao buscar termo com ID vazio, mas obteve nil")
        }
    })
}
package notas

import (
	"reflect"
	"testing"
	"time" // Required for parsing time in test setup if needed

	"vickgenda/internal/models"
	"vickgenda/internal/store" // Required for Init and Clear
)

// Helper function to setup for each test or group of tests
func setupTermTestTable(t *testing.T) {
	nextTermID = 1 // Reset package-level ID counter for predictable IDs
	// Ensure db path is set for testing, e.g., to an in-memory db
	// This might be done in a TestMain or globally, but for safety, ensure db is usable.
	// Example: db.SetDBPath(":memory:") or a test-specific file.
	// For now, assume db.GetDB() will work or is configured elsewhere.
	err := store.InitTermsTable() // Ensure table exists
	if err != nil {
		t.Fatalf("Failed to initialize terms table for test: %v", err)
	}
	err = store.ClearTermsTableForTesting() // Clear data before each test
	if err != nil {
		t.Fatalf("Failed to clear terms table for test: %v", err)
	}
}

func TestConfigurarBimestreAdicionar(t *testing.T) {
	tests := []struct {
		name            string
		anoLetivo       int
		nome            string
		inicioStr       string
		fimStr          string
		setupPreexisting func(t *testing.T) // Function to setup pre-existing conflicting data
		expectError     bool
		expectedTermName string
		expectedID      string // For checking predictable ID
	}{
		{
			name: "bimestre válido", anoLetivo: 2024, nome: "1º Bimestre", inicioStr: "01-02-2024", fimStr: "15-04-2024",
			expectError: false, expectedTermName: "1º Bimestre", expectedID: "term001",
		},
		{
			name: "data fim antes de inicio", anoLetivo: 2024, nome: "Inválido", inicioStr: "15-04-2024", fimStr: "01-02-2024",
			expectError: true,
		},
		{
			name: "campos obrigatórios faltando", anoLetivo: 2024, nome: "", inicioStr: "01-02-2024", fimStr: "15-04-2024",
			expectError: true,
		},
		{
			name: "formato data inválido", anoLetivo: 2024, nome: "Bimestre X", inicioStr: "2024/02/01", fimStr: "2024/04/15",
			expectError: true,
		},
		{
			name: "adicionar segundo bimestre válido após o primeiro", anoLetivo: 2024, nome: "2º Bimestre", inicioStr: "16-04-2024", fimStr: "30-06-2024",
			setupPreexisting: func(t *testing.T) {
				// nextTermID is 1 initially in this sub-test due to setupTermTestTable
				// It will be consumed by the pre-existing term.
				_, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024") // Consumes term001
				if err != nil {
					t.Fatalf("Setup error for 'adicionar segundo bimestre': %v", err)
				}
				// nextTermID is now 2 for the term being tested in this specific sub-test.
			},
			expectError: false, expectedTermName: "2º Bimestre", expectedID: "term002",
		},
		{
			name: "tentar adicionar bimestre com mesmo nome no mesmo ano", anoLetivo: 2024, nome: "1º Bimestre", inicioStr: "01-05-2024", fimStr: "15-07-2024",
			setupPreexisting: func(t *testing.T) {
				_, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024")
				if err != nil {
					t.Fatalf("Setup error for 'mesmo nome no mesmo ano': %v", err)
				}
			},
			expectError: true, // DB unique constraint (name, year) should trigger this
		},
        {
			// This test's expectation changes. With the previous in-memory logic, this might have been an error
			// if the validation was strict about any date overlap.
			// However, the DB constraint is UNIQUE(name, strftime('%Y', start_date)).
			// It does NOT prevent overlapping date ranges if the names are different.
			// The Go code for ConfigurarBimestreAdicionar now primarily relies on the DB for this.
			// If strict overlap prevention for different names is desired, it needs to be re-added as a Go check.
			// For now, this test reflects that different names can have overlapping dates.
			name: "bimestre com sobreposição de datas nomes diferentes", anoLetivo: 2024, nome: "Sobreposto", inicioStr: "10-04-2024", fimStr: "20-04-2024",
			setupPreexisting: func(t *testing.T) {
				_, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024") // term001
                if err != nil { t.Fatalf("Setup error for 'sobreposição datas nomes diferentes': %v", err) }
			},
			expectError: false, expectedTermName: "Sobreposto", expectedID: "term002", // Will be term002 as 1st Bimestre was term001
		},
        {
			name: "bimestre válido em outro ano", anoLetivo: 2025, nome: "1º Bimestre", inicioStr: "01-02-2025", fimStr: "15-04-2025",
			// No setupPreexisting needed as setupTermTestTable clears and it's a different year.
			expectError: false, expectedTermName: "1º Bimestre", expectedID: "term001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTermTestTable(t) // Clear, Init table, and reset nextTermID for each sub-test

			if tt.setupPreexisting != nil {
				tt.setupPreexisting(t) // This will affect nextTermID for the actual call below
			}

			term, err := ConfigurarBimestreAdicionar(tt.anoLetivo, tt.nome, tt.inicioStr, tt.fimStr)
			if tt.expectError {
				if err == nil {
					t.Errorf("esperado erro, mas obteve nil. Teste: %s", tt.name)
				}
			} else {
				if err != nil {
					t.Fatalf("esperado sem erro, mas obteve: %v. Teste: %s", err, tt.name)
				}
				if term.ID == "" {
					t.Errorf("esperado ID do termo preenchido. Teste: %s", tt.name)
				}
                if term.ID != tt.expectedID {
					t.Errorf("esperado ID do termo '%s', obteve '%s'. Teste: %s", tt.expectedID, term.ID, tt.name)
				}
				if term.Name != tt.expectedTermName {
					t.Errorf("esperado nome do termo '%s', obteve '%s'. Teste: %s", tt.expectedTermName, term.Name, tt.name)
				}
				// Verify it's in the DB
				_, dbErr := store.GetTermByID(term.ID)
				if dbErr != nil {
					t.Errorf("Termo adicionado não encontrado no DB: %v. Teste: %s", dbErr, tt.name)
				}
			}
		})
	}
}

func TestConfigurarBimestreListar(t *testing.T) {
    setupTermTestTable(t) // Clear, Init table, and reset nextTermID

    // Add terms directly using the function being tested or store.CreateTerm for setup.
    // Using ConfigurarBimestreAdicionar ensures IDs are generated as expected by tests.
    term1_2024, err := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024") // term001
    if err != nil { t.Fatalf("TestConfigurarBimestreListar setup error term1_2024: %v", err) }
    term2_2024, err := ConfigurarBimestreAdicionar(2024, "2º Bimestre", "16-04-2024", "30-06-2024") // term002
    if err != nil { t.Fatalf("TestConfigurarBimestreListar setup error term2_2024: %v", err) }
    term1_2025, err := ConfigurarBimestreAdicionar(2025, "1º Bimestre", "01-02-2025", "15-04-2025") // term003
    if err != nil { t.Fatalf("TestConfigurarBimestreListar setup error term1_2025: %v", err) }

    tests := []struct {
        name          string
        anoLetivo     int
        expectError   bool
        expectedCount int
        expectedTerms []models.Term
    }{
        // Note: The store.ListTermsByYear already sorts by start_date.
        {"listar bimestres 2024", 2024, false, 2, []models.Term{term1_2024, term2_2024}},
        {"listar bimestres 2025", 2025, false, 1, []models.Term{term1_2025}},
        {"listar ano sem bimestres", 2026, false, 0, []models.Term{}},
        {"listar com ano 0 (erro)", 0, true, 0, nil},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            terms, err := ConfigurarBimestreListar(tt.anoLetivo)
            if tt.expectError {
                if err == nil {
                    t.Errorf("esperado erro, mas obteve nil. Teste: %s", tt.name)
                }
            } else {
                if err != nil {
                    t.Fatalf("esperado sem erro, mas obteve: %v. Teste: %s", err, tt.name)
                }
                if len(terms) != tt.expectedCount {
                    t.Errorf("esperado %d termos, obteve %d. Teste: %s", tt.expectedCount, len(terms), tt.name)
                }
                // DeepEqual checks order as well.
                if tt.expectedCount > 0 { // Only compare if we expect terms
                    if !reflect.DeepEqual(terms, tt.expectedTerms) {
                        t.Errorf("lista de termos não corresponde ao esperado. Teste: %s\nEsperado: %+v\nObtido:   %+v", tt.name, tt.expectedTerms, terms)
                    }
                } else if len(terms) != 0 { // If expectedCount is 0, terms should be empty
                     t.Errorf("esperado lista vazia, obteve %d termos. Teste: %s", len(terms), tt.name)
                }
            }
        })
    }
}

func TestGetTermByID(t *testing.T) {
    setupTermTestTable(t) // Clear, Init table, and reset nextTermID

    // ID will be term001 due to reset in setupTermTestTable
    addedTerm, err := ConfigurarBimestreAdicionar(2024, "Único Bimestre", "01-03-2024", "30-04-2024")
    if err != nil { t.Fatalf("TestGetTermByID setup error: %v", err) }

    t.Run("buscar termo existente", func(t *testing.T) {
        term, err := GetTermByID(addedTerm.ID) // addedTerm.ID should be "term001"
        if err != nil {
            t.Fatalf("Erro ao buscar termo existente: %v", err)
        }
        if term.ID != addedTerm.ID || term.Name != addedTerm.Name {
            // For more detailed comparison, especially with time.Time
            if !reflect.DeepEqual(term, addedTerm) {
                 t.Errorf("Termo recuperado não corresponde ao adicionado.\nEsperado: %+v\nObtido:   %+v", addedTerm, term)
            }
        }
    })

    t.Run("buscar termo inexistente", func(t *testing.T) {
        _, err := GetTermByID("term_nao_existe")
        if err == nil {
            t.Error("Esperado erro ao buscar termo inexistente, mas obteve nil")
        }
    })

    t.Run("buscar termo com ID vazio", func(t *testing.T) {
        _, err := GetTermByID("")
        if err == nil {
            t.Error("Esperado erro ao buscar termo com ID vazio, mas obteve nil")
        }
    })
}
