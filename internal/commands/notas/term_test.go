package notas

import (
	"reflect"
	"testing"

	"vickgenda/internal/models"
)

func TestConfigurarBimestreAdicionar(t *testing.T) {
    LimparStoreTermos()
	tests := []struct {
		name        string
		anoLetivo   int
		nome        string
		inicioStr   string
		fimStr      string
		expectError bool
        expectedTermName string
	}{
		{
			name: "bimestre válido", anoLetivo: 2024, nome: "1º Bimestre", inicioStr: "01-02-2024", fimStr: "15-04-2024",
			expectError: false, expectedTermName: "1º Bimestre",
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
			name: "adicionar segundo bimestre válido", anoLetivo: 2024, nome: "2º Bimestre", inicioStr: "16-04-2024", fimStr: "30-06-2024",
			expectError: false, expectedTermName: "2º Bimestre",
		},
		{
			name: "tentar adicionar bimestre com mesmo nome no mesmo ano", anoLetivo: 2024, nome: "1º Bimestre", inicioStr: "01-05-2024", fimStr: "15-07-2024",
			expectError: true, // Já existe "1º Bimestre" de um teste anterior bem sucedido
		},
        {
			name: "bimestre com sobreposição de datas", anoLetivo: 2024, nome: "Sobreposto", inicioStr: "10-04-2024", fimStr: "20-04-2024", // Sobrepõe com 1º e 2º Bimestres
			expectError: true,
		},
        {
			name: "bimestre válido em outro ano", anoLetivo: 2025, nome: "1º Bimestre", inicioStr: "01-02-2025", fimStr: "15-04-2025",
			expectError: false, expectedTermName: "1º Bimestre",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Limpar store apenas para testes que não dependem de estado anterior (ex: nome duplicado)
            // Para este conjunto de testes, é melhor limpar a cada run para isolar,
            // exceto para os de duplicação/sobreposição que precisam de dados pré-existentes.
            // Vamos refatorar para limpar e adicionar base se necessário.
            if tt.name == "bimestre válido" || tt.name == "bimestre válido em outro ano" || tt.name == "data fim antes de inicio" || tt.name == "campos obrigatórios faltando" || tt.name == "formato data inválido" {
                LimparStoreTermos()
            }
            // Para testes de conflito, precisamos garantir que o estado base exista
            if tt.name == "tentar adicionar bimestre com mesmo nome no mesmo ano" || tt.name == "bimestre com sobreposição de datas" || tt.name == "adicionar segundo bimestre válido" {
                // Se o store estiver vazio (como no início ou após LimparStoreTermos), adicione o primeiro bimestre.
                if len(termList) == 0 || termList[0].Name != "1º Bimestre" || termList[0].StartDate.Year() != 2024 {
                    LimparStoreTermos() // Limpa para garantir que não haja lixo de execuções anteriores de outros testes
                    _, _ = ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024")
                }
            }
             if tt.name == "bimestre com sobreposição de datas" {
                // Garante que o 2º Bimestre também exista para testar sobreposição com ele
                // Esta lógica de setup está ficando um pouco complexa, idealmente cada teste é mais independente.
                // Mas para este caso, é necessário ter ambos os bimestres.
                 foundSecond := false
                 for _, term := range termList {
                     if term.Name == "2º Bimestre" && term.StartDate.Year() == 2024 {
                         foundSecond = true
                         break
                     }
                 }
                 if !foundSecond {
                    _, _ = ConfigurarBimestreAdicionar(2024, "2º Bimestre", "16-04-2024", "30-06-2024")
                 }
            }


			term, err := ConfigurarBimestreAdicionar(tt.anoLetivo, tt.nome, tt.inicioStr, tt.fimStr)
			if tt.expectError {
				if err == nil {
					t.Errorf("esperado erro, mas obteve nil")
				}
			} else {
				if err != nil {
					t.Fatalf("esperado sem erro, mas obteve: %v", err)
				}
				if term.ID == "" {
					t.Errorf("esperado ID do termo preenchido")
				}
				if term.Name != tt.expectedTermName {
					t.Errorf("esperado nome do termo '%s', obteve '%s'", tt.expectedTermName, term.Name)
				}
			}
		})
	}
}

func TestConfigurarBimestreListar(t *testing.T) {
    LimparStoreTermos()
    // Adicionar termos para teste
    term1_2024, _ := ConfigurarBimestreAdicionar(2024, "1º Bimestre", "01-02-2024", "15-04-2024")
    term2_2024, _ := ConfigurarBimestreAdicionar(2024, "2º Bimestre", "16-04-2024", "30-06-2024")
    term1_2025, _ := ConfigurarBimestreAdicionar(2025, "1º Bimestre", "01-02-2025", "15-04-2025")

    tests := []struct {
        name          string
        anoLetivo     int
        expectError   bool
        expectedCount int
        expectedTerms []models.Term // Ordem importa aqui
    }{
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
                    t.Errorf("esperado erro, mas obteve nil")
                }
            } else {
                if err != nil {
                    t.Fatalf("esperado sem erro, mas obteve: %v", err)
                }
                if len(terms) != tt.expectedCount {
                    t.Errorf("esperado %d termos, obteve %d", tt.expectedCount, len(terms))
                }
                if tt.expectedCount > 0 && !reflect.DeepEqual(terms, tt.expectedTerms) {
                    t.Errorf(`lista de termos não corresponde ao esperado.
Esperado: %+v
Obtido:   %+v`, tt.expectedTerms, terms)
                }
            }
        })
    }
}

func TestGetTermByID(t *testing.T) {
    LimparStoreTermos()
    addedTerm, _ := ConfigurarBimestreAdicionar(2024, "Único Bimestre", "01-03-2024", "30-04-2024")

    t.Run("buscar termo existente", func(t *testing.T) {
        term, err := GetTermByID(addedTerm.ID)
        if err != nil {
            t.Fatalf("Erro ao buscar termo existente: %v", err)
        }
        if term.ID != addedTerm.ID || term.Name != addedTerm.Name {
            t.Errorf("Termo recuperado não corresponde ao adicionado. Esperado %+v, obtido %+v", addedTerm, term)
        }
    })

    t.Run("buscar termo inexistente", func(t *testing.T) {
        _, err := GetTermByID("term_nao_existe")
        if err == nil {
            t.Error("Esperado erro ao buscar termo inexistente, mas obteve nil")
        }
    })
}
