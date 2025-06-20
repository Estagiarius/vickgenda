package notas

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"vickgenda/internal/models"
)

// ... (Conteúdo anterior de grade.go: stores, IDs, mocks, LancarNota, VerNotas, MediaInfo, CalcularMedia) ...
var gradesStore = make(map[string]models.Grade)
var gradeList = []models.Grade{}
var nextGradeID = 1
var mockStudentStore = make(map[string]models.Student)

func generateGradeID() string {
	id := fmt.Sprintf("nota%03d", nextGradeID)
	nextGradeID++
	return id
}

func LimparStoreGrades() {
	gradesStore = make(map[string]models.Grade)
	gradeList = []models.Grade{}
	nextGradeID = 1
	mockStudentStore = make(map[string]models.Student)
}

func MockAddStudent(id, name string) {
	mockStudentStore[id] = models.Student{ID: id, Name: name}
}

func mockStudentExiste(studentID string) bool {
	_, found := mockStudentStore[studentID]
	return found
}

func LancarNota(alunoID, bimestreID, disciplina, avaliacaoDesc string, valorNota, pesoNota float64, dataStr string) (models.Grade, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" || avaliacaoDesc == "" {
		return models.Grade{}, errors.New("ID do aluno, ID do bimestre, disciplina e descrição da avaliação são obrigatórios")
	}
	if !mockStudentExiste(alunoID) {
		return models.Grade{}, fmt.Errorf("aluno com ID '%s' não encontrado", alunoID)
	}
	_, err := GetTermByID(bimestreID)
	if err != nil {
		return models.Grade{}, fmt.Errorf("bimestre com ID '%s' não encontrado: %w", bimestreID, err)
	}
	if valorNota < 0 || valorNota > 10 {
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
	newID := generateGradeID()
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
	gradesStore[newID] = grade

    var foundInList bool
    for i, g := range gradeList {
        if g.ID == newID {
            gradeList[i] = grade
            foundInList = true
            break
        }
    }
    if !foundInList {
        gradeList = append(gradeList, grade)
    }
	return grade, nil
}

func VerNotas(alunoID, bimestreID, disciplina string) ([]models.Grade, error) {
	if alunoID == "" {
		return nil, errors.New("ID do aluno é obrigatório para ver notas")
	}
	if !mockStudentExiste(alunoID) {
		return nil, fmt.Errorf("aluno com ID '%s' não encontrado", alunoID)
	}
    if bimestreID != "" {
        _, err := GetTermByID(bimestreID)
        if err != nil {
            return nil, fmt.Errorf("bimestre com ID '%s' não encontrado ao tentar filtrar notas: %w", bimestreID, err)
        }
    }
	var result []models.Grade
    for _, grade := range gradeList {
		match := true
		if grade.StudentID != alunoID {
			match = false
		}
		if bimestreID != "" && grade.TermID != bimestreID {
			match = false
		}
		if disciplina != "" && !strings.EqualFold(grade.Subject, disciplina) {
			match = false
		}
		if match {
			result = append(result, grade)
		}
	}
    sort.Slice(result, func(i, j int) bool {
        return result[i].Date.Before(result[j].Date)
    })
	return result, nil
}

type MediaInfo struct {
    AlunoID         string
    BimestreID      string
    Disciplina      string
    MediaPonderada  float64
    SomaPesos       float64
    NotasConsideradas []models.Grade
}

func CalcularMedia(alunoID, bimestreID, disciplina string) (MediaInfo, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" {
		return MediaInfo{}, errors.New("ID do aluno, ID do bimestre e disciplina são obrigatórios para calcular a média")
	}
	if !mockStudentExiste(alunoID) {
		return MediaInfo{}, fmt.Errorf("aluno com ID '%s' não encontrado", alunoID)
	}
	_, err := GetTermByID(bimestreID)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("bimestre com ID '%s' não encontrado: %w", bimestreID, err)
	}
	notasDoAlunoNoBimestreDisciplina, err := VerNotas(alunoID, bimestreID, disciplina)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("erro ao buscar notas para cálculo da média: %w", err)
	}
	if len(notasDoAlunoNoBimestreDisciplina) == 0 {
		return MediaInfo{}, fmt.Errorf("nenhuma nota encontrada para o aluno '%s' no bimestre '%s' para a disciplina '%s'. Média não pode ser calculada", alunoID, bimestreID, disciplina)
	}
	var somaValoresPonderados float64
	var somaPesos float64
	for _, nota := range notasDoAlunoNoBimestreDisciplina {
		somaValoresPonderados += nota.Value * nota.Weight
		somaPesos += nota.Weight
	}
	if somaPesos == 0 {
		return MediaInfo{
            AlunoID:         alunoID,
            BimestreID:      bimestreID,
            Disciplina:      disciplina,
            MediaPonderada:  0,
            SomaPesos:       0,
            NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
        }, errors.New("soma dos pesos das notas é zero, média não pode ser calculada")
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

// EditarNota permite alterar campos de uma nota existente.
func EditarNota(idNota string, novoValor *float64, novoPeso *float64, novaDesc *string, novaDataStr *string) (models.Grade, error) {
    grade, found := gradesStore[idNota]
    if !found {
        return models.Grade{}, fmt.Errorf("nota com ID '%s' não encontrada para edição", idNota)
    }

    algoAlterado := false

    if novoValor != nil {
        if *novoValor < 0 || *novoValor > 10 { // Revalidar
            return models.Grade{}, fmt.Errorf("novo valor da nota (%.2f) fora do intervalo permitido (0-10)", *novoValor)
        }
        if grade.Value != *novoValor {
            grade.Value = *novoValor
            algoAlterado = true
        }
    }

    if novoPeso != nil {
        if *novoPeso <= 0 { // Revalidar
            return models.Grade{}, errors.New("novo peso da nota deve ser um valor positivo")
        }
        if grade.Weight != *novoPeso {
            grade.Weight = *novoPeso
            algoAlterado = true
        }
    }

    if novaDesc != nil {
        if grade.Description != *novaDesc {
            grade.Description = *novaDesc
            algoAlterado = true
        }
    }

    if novaDataStr != nil {
        layout := "02-01-2006"
        dataAvaliacao, err := time.Parse(layout, *novaDataStr)
        if err != nil {
            return models.Grade{}, fmt.Errorf("formato de nova data inválido: %s. Use dd-mm-aaaa", *novaDataStr)
        }
        if !grade.Date.Equal(dataAvaliacao) {
            grade.Date = dataAvaliacao
            algoAlterado = true
        }
    }

    if !algoAlterado {
        return grade, errors.New("nenhuma alteração fornecida para a nota")
    }

    gradesStore[idNota] = grade
    for i, g := range gradeList {
        if g.ID == idNota {
            gradeList[i] = grade
            break
        }
    }
    return grade, nil
}

// ExcluirNota remove uma nota do armazenamento.
func ExcluirNota(idNota string) error {
    _, found := gradesStore[idNota]
    if !found {
        return fmt.Errorf("nota com ID '%s' não encontrada para exclusão", idNota)
    }
    delete(gradesStore, idNota)

    newList := []models.Grade{}
    for _, g := range gradeList {
        if g.ID != idNota {
            newList = append(newList, g)
        }
    }
    gradeList = newList
    return nil
}
