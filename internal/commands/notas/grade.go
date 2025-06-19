package notas

import (
	"errors"
	"fmt"
	// "sort" // No longer needed here if store sorts VerNotas
	"strings"
	"time"

	"vickgenda/internal/models"
	"vickgenda/internal/store"
)

// generateGradeID remains for now, called by LancarNota before CreateGrade.
// Consider moving to UUIDs or DB-generated IDs later.
var currentGradeID = 1 // Renamed from nextGradeID to avoid confusion, as it's used to generate current ID

func generateGradeID() string {
	id := fmt.Sprintf("nota%03d", currentGradeID)
	currentGradeID++
	return id
}

// ResetGradeIDCounterForTesting resets the grade ID counter. To be called in test setups.
// This is needed because grade tests run sequentially and expect predictable IDs when using generateGradeID.
func ResetGradeIDCounterForTesting() {
	currentGradeID = 1
}

// LancarNota adiciona uma nova nota, now using the store.
func LancarNota(alunoID, bimestreID, disciplina, avaliacaoDesc string, valorNota, pesoNota float64, dataStr string) (models.Grade, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" || avaliacaoDesc == "" {
		return models.Grade{}, errors.New("ID do aluno, ID do bimestre, disciplina e descrição da avaliação são obrigatórios")
	}

	// Validate student and term existence
	_, err := store.GetStudentByID(alunoID)
	if err != nil {
		return models.Grade{}, fmt.Errorf("validação do aluno falhou: %w", err)
	}

	_, err = GetTermByID(bimestreID) // This GetTermByID is from the notas package, which already uses store.GetTermByID
	if err != nil {
		return models.Grade{}, fmt.Errorf("bimestre com ID '%s' não encontrado: %w", bimestreID, err)
	}

	// Validate grade values
	if valorNota < 0 || valorNota > 10 { // Assuming 0-10 scale
		return models.Grade{}, fmt.Errorf("valor da nota (%.2f) fora do intervalo permitido (0-10)", valorNota)
	}
	if pesoNota <= 0 {
		return models.Grade{}, errors.New("peso da nota deve ser um valor positivo")
	}

	var dataAvaliacao time.Time
	if dataStr != "" {
		layout := "02-01-2006"
		dataAvaliacao, err = time.Parse(layout, dataStr)
		if err != nil {
			return models.Grade{}, fmt.Errorf("formato de data inválido: %s. Use dd-mm-aaaa", dataStr)
		}
	} else {
		dataAvaliacao = time.Now()
	}

	newID := generateGradeID() // Still generating ID here; store will just use it.
	grade := models.Grade{
		ID:          newID,
		StudentID:   alunoID,
		TermID:      bimestreID,
		Subject:     disciplina,
		Description: avaliacaoDesc,
		Value:       valorNota,
		Weight:      pesoNota,
		Date:        dataAvaliacao,
	}

	err = store.CreateGrade(grade)
	if err != nil {
		// Potential errors: DB connection, constraint violations (e.g. foreign key if student/term deleted mid-op)
		return models.Grade{}, fmt.Errorf("erro ao salvar nota no banco de dados: %w", err)
	}

	return grade, nil
}

// VerNotas visualiza as notas de um aluno, now using the store.
func VerNotas(alunoID, bimestreID, disciplina string) ([]models.Grade, error) {
	if alunoID == "" {
		return nil, errors.New("ID do aluno é obrigatório para ver notas")
	}

	// Validate student existence (optional, as ListGrades might return empty if student has no grades,
	// but good for consistent error reporting if student must exist).
	_, err := store.GetStudentByID(alunoID)
	if err != nil {
		return nil, fmt.Errorf("validação do aluno para VerNotas falhou: %w", err)
	}

	// Validate term existence if provided (optional, ListGrades handles empty termID for no filter by term)
	if bimestreID != "" {
		_, err := GetTermByID(bimestreID)
		if err != nil {
			return nil, fmt.Errorf("bimestre com ID '%s' não encontrado ao tentar filtrar notas: %w", bimestreID, err)
		}
	}

	// Fetch grades from the store.
	// The store's ListGrades function handles filtering by studentID, termID (optional), and subject (optional).
	// It also handles sorting by date.
	grades, err := store.ListGrades(alunoID, bimestreID, disciplina)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar notas no banco de dados: %w", err)
	}

	return grades, nil
}

// MediaInfo struct remains the same.
type MediaInfo struct {
    AlunoID         string
    BimestreID      string
    Disciplina      string
    MediaPonderada  float64
    SomaPesos       float64
    NotasConsideradas []models.Grade
}

// CalcularMedia calculates the weighted average for a student in a subject/term.
// It now relies on VerNotas which fetches data from the store.
func CalcularMedia(alunoID, bimestreID, disciplina string) (MediaInfo, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" {
		return MediaInfo{}, errors.New("ID do aluno, ID do bimestre e disciplina são obrigatórios para calcular a média")
	}

	// Student and Term validation can be done upfront.
	// VerNotas also does this, but explicit checks here provide clear context.
	_, err := store.GetStudentByID(alunoID)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("validação do aluno para CalcularMedia falhou: %w", err)
	}
	_, err = GetTermByID(bimestreID)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("validação do bimestre para CalcularMedia falhou: %w", err)
	}

	notasDoAlunoNoBimestreDisciplina, err := VerNotas(alunoID, bimestreID, disciplina)
	if err != nil {
		// This error would be from store.ListGrades via VerNotas
		return MediaInfo{}, fmt.Errorf("erro ao buscar notas para cálculo da média: %w", err)
	}

	if len(notasDoAlunoNoBimestreDisciplina) == 0 {
		return MediaInfo{ // Return info with zeroed fields and specific error
            AlunoID:    alunoID,
            BimestreID: bimestreID,
            Disciplina: disciplina,
        }, fmt.Errorf("nenhuma nota encontrada para o aluno '%s' no bimestre '%s' para a disciplina '%s'. Média não pode ser calculada", alunoID, bimestreID, disciplina)
	}

	var somaValoresPonderados float64
	var somaPesos float64
	for _, nota := range notasDoAlunoNoBimestreDisciplina {
		somaValoresPonderados += nota.Value * nota.Weight
		somaPesos += nota.Weight
	}

	if somaPesos == 0 { // Should ideally not happen if notes exist and have Weight > 0 defined in business logic
		return MediaInfo{
            AlunoID:         alunoID,
            BimestreID:      bimestreID,
            Disciplina:      disciplina,
            MediaPonderada:  0, // Or handle as NaN, or maintain error
            SomaPesos:       0,
            NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
        }, errors.New("soma dos pesos das notas é zero, média não pode ser calculada (verifique se as notas têm pesos válidos atribuídos)")
	}

	mediaFinal := somaValoresPonderados / somaPesos
	return MediaInfo{
		AlunoID:         alunoID,
		BimestreID:      bimestreID,
		Disciplina:      disciplina,
		MediaPonderada:  mediaFinal,
		SomaPesos:       somaPesos,
		NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
	}, nil
}

// EditarNota updates an existing grade using the store.
func EditarNota(idNota string, novoValor *float64, novoPeso *float64, novaDesc *string, novaDataStr *string) (models.Grade, error) {
    grade, err := store.GetGradeByID(idNota) // Fetch current grade from DB
    if err != nil {
        // Error from store.GetGradeByID is already descriptive (e.g., "grade with ID '...' not found")
        return models.Grade{}, fmt.Errorf("falha ao buscar nota para edição: %w", err)
    }

    algoAlterado := false
    if novoValor != nil {
        if *novoValor < 0 || *novoValor > 10 {
            return models.Grade{}, fmt.Errorf("novo valor da nota (%.2f) fora do intervalo permitido (0-10)", *novoValor)
        }
        if grade.Value != *novoValor { grade.Value = *novoValor; algoAlterado = true }
    }
    if novoPeso != nil {
        if *novoPeso <= 0 {
            return models.Grade{}, errors.New("novo peso da nota deve ser um valor positivo")
        }
        if grade.Weight != *novoPeso { grade.Weight = *novoPeso; algoAlterado = true }
    }
    if novaDesc != nil {
        if grade.Description != *novaDesc { grade.Description = *novaDesc; algoAlterado = true }
    }
    if novaDataStr != nil {
        layout := "02-01-2006"
        dataAvaliacao, errParseDate := time.Parse(layout, *novaDataStr)
        if errParseDate != nil {
            return models.Grade{}, fmt.Errorf("formato de nova data inválido: '%s'. Use dd-mm-aaaa", *novaDataStr)
        }
        // Compare date part only if time part is not relevant, or ensure full time comparison
        if !grade.Date.Equal(dataAvaliacao) { grade.Date = dataAvaliacao; algoAlterado = true }
    }

    if !algoAlterado {
        return grade, errors.New("nenhuma alteração fornecida para a nota")
    }

    err = store.UpdateGrade(grade) // Update in the database
    if err != nil {
        return models.Grade{}, fmt.Errorf("erro ao atualizar nota no banco de dados: %w", err)
    }
    return grade, nil
}

// ExcluirNota removes a grade using the store.
func ExcluirNota(idNota string) error {
    // Optional: Check if grade exists using GetGradeByID first, to provide a potentially more specific error
    // before attempting delete. store.DeleteGrade itself will return an error if ID not found.
    _, err := store.GetGradeByID(idNota)
    if err != nil {
         return fmt.Errorf("falha ao buscar nota para exclusão (verificação prévia): %w", err)
    }

    err = store.DeleteGrade(idNota) // Attempt to delete from the database
    if err != nil {
        // This error could be "no grade found with ID" or other DB issue.
        return fmt.Errorf("erro ao excluir nota do banco de dados: %w", err)
    }
    return nil
}

// LimparStoreGrades foi REMOVIDA.
// Testes devem agora usar store.ClearGradesTableForTesting() para limpar dados da tabela 'grades'
// e ResetGradeIDCounterForTesting() se dependerem da geração sequencial de IDs por generateGradeID().
package notas

import (
	"errors"
	"fmt"
	// "sort" // No longer needed here if store sorts VerNotas
	"strings"
	"time"

	"vickgenda/internal/models"
	"vickgenda/internal/store"
)

// generateGradeID remains for now, called by LancarNota before CreateGrade.
// Consider moving to UUIDs or DB-generated IDs later.
var currentGradeID = 1 // Renamed from nextGradeID to avoid confusion, as it's used to generate current ID

func generateGradeID() string {
	id := fmt.Sprintf("nota%03d", currentGradeID)
	currentGradeID++
	return id
}

// ResetGradeIDCounterForTesting resets the grade ID counter. To be called in test setups.
// This is needed because grade tests run sequentially and expect predictable IDs when using generateGradeID.
func ResetGradeIDCounterForTesting() {
	currentGradeID = 1
}

// LancarNota adiciona uma nova nota, now using the store.
func LancarNota(alunoID, bimestreID, disciplina, avaliacaoDesc string, valorNota, pesoNota float64, dataStr string) (models.Grade, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" || avaliacaoDesc == "" {
		return models.Grade{}, errors.New("ID do aluno, ID do bimestre, disciplina e descrição da avaliação são obrigatórios")
	}

	// Validate student and term existence
	_, err := store.GetStudentByID(alunoID)
	if err != nil {
		return models.Grade{}, fmt.Errorf("validação do aluno falhou: %w", err)
	}

	_, err = GetTermByID(bimestreID) // This GetTermByID is from the notas package, which already uses store.GetTermByID
	if err != nil {
		return models.Grade{}, fmt.Errorf("bimestre com ID '%s' não encontrado: %w", bimestreID, err)
	}

	// Validate grade values
	if valorNota < 0 || valorNota > 10 { // Assuming 0-10 scale
		return models.Grade{}, fmt.Errorf("valor da nota (%.2f) fora do intervalo permitido (0-10)", valorNota)
	}
	if pesoNota <= 0 {
		return models.Grade{}, errors.New("peso da nota deve ser um valor positivo")
	}

	var dataAvaliacao time.Time
	if dataStr != "" {
		layout := "02-01-2006"
		dataAvaliacao, err = time.Parse(layout, dataStr)
		if err != nil {
			return models.Grade{}, fmt.Errorf("formato de data inválido: %s. Use dd-mm-aaaa", dataStr)
		}
	} else {
		dataAvaliacao = time.Now()
	}

	newID := generateGradeID() // Still generating ID here; store will just use it.
	grade := models.Grade{
		ID:          newID,
		StudentID:   alunoID,
		TermID:      bimestreID,
		Subject:     disciplina,
		Description: avaliacaoDesc,
		Value:       valorNota,
		Weight:      pesoNota,
		Date:        dataAvaliacao,
	}

	err = store.CreateGrade(grade)
	if err != nil {
		// Potential errors: DB connection, constraint violations (e.g. foreign key if student/term deleted mid-op)
		return models.Grade{}, fmt.Errorf("erro ao salvar nota no banco de dados: %w", err)
	}

	return grade, nil
}

// VerNotas visualiza as notas de um aluno, now using the store.
func VerNotas(alunoID, bimestreID, disciplina string) ([]models.Grade, error) {
	if alunoID == "" {
		return nil, errors.New("ID do aluno é obrigatório para ver notas")
	}

	// Validate student existence (optional, as ListGrades might return empty if student has no grades,
	// but good for consistent error reporting if student must exist).
	_, err := store.GetStudentByID(alunoID)
	if err != nil {
		return nil, fmt.Errorf("validação do aluno para VerNotas falhou: %w", err)
	}

	// Validate term existence if provided (optional, ListGrades handles empty termID for no filter by term)
	if bimestreID != "" {
		_, err := GetTermByID(bimestreID)
		if err != nil {
			return nil, fmt.Errorf("bimestre com ID '%s' não encontrado ao tentar filtrar notas: %w", bimestreID, err)
		}
	}

	// Fetch grades from the store.
	// The store's ListGrades function handles filtering by studentID, termID (optional), and subject (optional).
	// It also handles sorting by date.
	grades, err := store.ListGrades(alunoID, bimestreID, disciplina)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar notas no banco de dados: %w", err)
	}

	return grades, nil
}

// MediaInfo struct remains the same.
type MediaInfo struct {
    AlunoID         string
    BimestreID      string
    Disciplina      string
    MediaPonderada  float64
    SomaPesos       float64
    NotasConsideradas []models.Grade
}

// CalcularMedia calculates the weighted average for a student in a subject/term.
// It now relies on VerNotas which fetches data from the store.
func CalcularMedia(alunoID, bimestreID, disciplina string) (MediaInfo, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" {
		return MediaInfo{}, errors.New("ID do aluno, ID do bimestre e disciplina são obrigatórios para calcular a média")
	}

	// Student and Term validation can be done upfront.
	// VerNotas also does this, but explicit checks here provide clear context.
	_, err := store.GetStudentByID(alunoID)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("validação do aluno para CalcularMedia falhou: %w", err)
	}
	_, err = GetTermByID(bimestreID)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("validação do bimestre para CalcularMedia falhou: %w", err)
	}

	notasDoAlunoNoBimestreDisciplina, err := VerNotas(alunoID, bimestreID, disciplina)
	if err != nil {
		// This error would be from store.ListGrades via VerNotas
		return MediaInfo{}, fmt.Errorf("erro ao buscar notas para cálculo da média: %w", err)
	}

	if len(notasDoAlunoNoBimestreDisciplina) == 0 {
		return MediaInfo{ // Return info with zeroed fields and specific error
            AlunoID:    alunoID,
            BimestreID: bimestreID,
            Disciplina: disciplina,
        }, fmt.Errorf("nenhuma nota encontrada para o aluno '%s' no bimestre '%s' para a disciplina '%s'. Média não pode ser calculada", alunoID, bimestreID, disciplina)
	}

	var somaValoresPonderados float64
	var somaPesos float64
	for _, nota := range notasDoAlunoNoBimestreDisciplina {
		somaValoresPonderados += nota.Value * nota.Weight
		somaPesos += nota.Weight
	}

	if somaPesos == 0 { // Should ideally not happen if notes exist and have Weight > 0 defined in business logic
		return MediaInfo{
            AlunoID:         alunoID,
            BimestreID:      bimestreID,
            Disciplina:      disciplina,
            MediaPonderada:  0, // Or handle as NaN, or maintain error
            SomaPesos:       0,
            NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
        }, errors.New("soma dos pesos das notas é zero, média não pode ser calculada (verifique se as notas têm pesos válidos atribuídos)")
	}

	mediaFinal := somaValoresPonderados / somaPesos
	return MediaInfo{
		AlunoID:         alunoID,
		BimestreID:      bimestreID,
		Disciplina:      disciplina,
		MediaPonderada:  mediaFinal,
		SomaPesos:       somaPesos,
		NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
	}, nil
}

// EditarNota updates an existing grade using the store.
func EditarNota(idNota string, novoValor *float64, novoPeso *float64, novaDesc *string, novaDataStr *string) (models.Grade, error) {
    grade, err := store.GetGradeByID(idNota) // Fetch current grade from DB
    if err != nil {
        // Error from store.GetGradeByID is already descriptive (e.g., "grade with ID '...' not found")
        return models.Grade{}, fmt.Errorf("falha ao buscar nota para edição: %w", err)
    }

    algoAlterado := false
    if novoValor != nil {
        if *novoValor < 0 || *novoValor > 10 {
            return models.Grade{}, fmt.Errorf("novo valor da nota (%.2f) fora do intervalo permitido (0-10)", *novoValor)
        }
        if grade.Value != *novoValor { grade.Value = *novoValor; algoAlterado = true }
    }
    if novoPeso != nil {
        if *novoPeso <= 0 {
            return models.Grade{}, errors.New("novo peso da nota deve ser um valor positivo")
        }
        if grade.Weight != *novoPeso { grade.Weight = *novoPeso; algoAlterado = true }
    }
    if novaDesc != nil {
        if grade.Description != *novaDesc { grade.Description = *novaDesc; algoAlterado = true }
    }
    if novaDataStr != nil {
        layout := "02-01-2006"
        dataAvaliacao, errParseDate := time.Parse(layout, *novaDataStr)
        if errParseDate != nil {
            return models.Grade{}, fmt.Errorf("formato de nova data inválido: '%s'. Use dd-mm-aaaa", *novaDataStr)
        }
        // Compare date part only if time part is not relevant, or ensure full time comparison
        if !grade.Date.Equal(dataAvaliacao) { grade.Date = dataAvaliacao; algoAlterado = true }
    }

    if !algoAlterado {
        return grade, errors.New("nenhuma alteração fornecida para a nota")
    }

    err = store.UpdateGrade(grade) // Update in the database
    if err != nil {
        return models.Grade{}, fmt.Errorf("erro ao atualizar nota no banco de dados: %w", err)
    }
    return grade, nil
}

// ExcluirNota removes a grade using the store.
func ExcluirNota(idNota string) error {
    // Optional: Check if grade exists using GetGradeByID first, to provide a potentially more specific error
    // before attempting delete. store.DeleteGrade itself will return an error if ID not found.
    _, err := store.GetGradeByID(idNota)
    if err != nil {
         return fmt.Errorf("falha ao buscar nota para exclusão (verificação prévia): %w", err)
    }

    err = store.DeleteGrade(idNota) // Attempt to delete from the database
    if err != nil {
        // This error could be "no grade found with ID" or other DB issue.
        return fmt.Errorf("erro ao excluir nota do banco de dados: %w", err)
    }
    return nil
}

// LimparStoreGrades foi REMOVIDA.
// Testes devem agora usar store.ClearGradesTableForTesting() para limpar dados da tabela 'grades'
// e ResetGradeIDCounterForTesting() se dependerem da geração sequencial de IDs por generateGradeID().
