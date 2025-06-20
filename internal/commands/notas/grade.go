package notas

import (
	"database/sql"
	"errors"
	"fmt"
	"strings" // Added missing import
	"time"

	"vickgenda-cli/internal/models"
	"vickgenda-cli/internal/store"
)

var ActiveGradeStore store.GradeStore
var ActiveStudentStoreForGrades store.StudentStore // To avoid naming conflict if student command has its own

func InitializeGradeStore(db *sql.DB) {
	ActiveGradeStore = store.NewSQLiteGradeStore(db)
}
func InitializeStudentStoreForGrades(db *sql.DB) {
	ActiveStudentStoreForGrades = store.NewSQLiteStudentStore(db)
}

func LancarNota(alunoID, bimestreID, disciplina, avaliacaoDesc string, valorNota, pesoNota float64, dataStr string) (models.Grade, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" || avaliacaoDesc == "" {
		return models.Grade{}, errors.New("ID do aluno, ID do bimestre, disciplina e descrição da avaliação são obrigatórios")
	}

	if ActiveStudentStoreForGrades == nil {
		return models.Grade{}, errors.New("StudentStore para grades não inicializado")
	}
	_, err := ActiveStudentStoreForGrades.GetStudentByID(alunoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") { // Adapt to specific error from store
			return models.Grade{}, fmt.Errorf("aluno com ID '%s' não encontrado: %w", alunoID, err)
		}
		return models.Grade{}, fmt.Errorf("erro ao verificar aluno com ID '%s': %w", alunoID, err)
	}

	// GetTermByID já usa ActiveTermStore (do pacote notas)
	_, err = GetTermByID(bimestreID)
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

	grade := models.Grade{
		// ID é gerado pelo store
		StudentID:   alunoID,
		TermID:      bimestreID,
		Subject:     disciplina,
		Description: avaliacaoDesc,
		Value:       valorNota,
		Weight:      pesoNota,
		Date:        dataAvaliacao,
	}

	if ActiveGradeStore == nil {
		return models.Grade{}, errors.New("GradeStore não inicializado")
	}
	savedGrade, err := ActiveGradeStore.SaveGrade(grade)
	if err != nil {
		return models.Grade{}, fmt.Errorf("erro ao lançar nota: %w", err)
	}
	return savedGrade, nil
}

func VerNotas(alunoID, bimestreID, disciplina string) ([]models.Grade, error) {
	if alunoID == "" {
		return nil, errors.New("ID do aluno é obrigatório para ver notas")
	}

	if ActiveStudentStoreForGrades == nil {
		return nil, errors.New("StudentStore para grades não inicializado")
	}
	_, err := ActiveStudentStoreForGrades.GetStudentByID(alunoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("aluno com ID '%s' não encontrado: %w", alunoID, err)
		}
		return nil, fmt.Errorf("erro ao verificar aluno com ID '%s': %w", alunoID, err)
	}

	if bimestreID != "" {
		_, err := GetTermByID(bimestreID) // Valida se o bimestre existe
		if err != nil {
			return nil, fmt.Errorf("bimestre com ID '%s' não encontrado ao tentar filtrar notas: %w", bimestreID, err)
		}
	}

	if ActiveGradeStore == nil {
		return nil, errors.New("GradeStore não inicializado")
	}
	grades, err := ActiveGradeStore.ListGradesByStudent(alunoID, bimestreID, disciplina)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar notas: %w", err)
	}
	return grades, nil
}

type MediaInfo struct {
	AlunoID           string
	BimestreID        string
	Disciplina        string
	MediaPonderada    float64
	SomaPesos         float64
	NotasConsideradas []models.Grade
}

func CalcularMedia(alunoID, bimestreID, disciplina string) (MediaInfo, error) {
	if alunoID == "" || bimestreID == "" || disciplina == "" {
		return MediaInfo{}, errors.New("ID do aluno, ID do bimestre e disciplina são obrigatórios para calcular a média")
	}

	if ActiveStudentStoreForGrades == nil {
		return MediaInfo{}, errors.New("StudentStore para grades não inicializado")
	}
	_, err := ActiveStudentStoreForGrades.GetStudentByID(alunoID) // Validar aluno
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return MediaInfo{}, fmt.Errorf("aluno com ID '%s' não encontrado: %w", alunoID, err)
		}
		return MediaInfo{}, fmt.Errorf("erro ao verificar aluno '%s': %w", alunoID, err)
	}

	_, err = GetTermByID(bimestreID) // Validar bimestre
	if err != nil {
		return MediaInfo{}, fmt.Errorf("bimestre com ID '%s' não encontrado: %w", bimestreID, err)
	}

	notasDoAlunoNoBimestreDisciplina, err := VerNotas(alunoID, bimestreID, disciplina)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("erro ao buscar notas para cálculo da média: %w", err)
	}
	if len(notasDoAlunoNoBimestreDisciplina) == 0 {
		// Retornar Média 0 e notas vazias em vez de erro direto, para consistência ou decisão do chamador
		return MediaInfo{
			AlunoID:           alunoID,
			BimestreID:        bimestreID,
			Disciplina:        disciplina,
			MediaPonderada:    0,
			SomaPesos:         0,
			NotasConsideradas: []models.Grade{},
		}, fmt.Errorf("nenhuma nota encontrada para o aluno '%s' no bimestre '%s' para a disciplina '%s'. Média não pode ser calculada", alunoID, bimestreID, disciplina)
	}

	var somaValoresPonderados float64
	var somaPesos float64
	for _, nota := range notasDoAlunoNoBimestreDisciplina {
		somaValoresPonderados += nota.Value * nota.Weight
		somaPesos += nota.Weight
	}

	if somaPesos == 0 {
		return MediaInfo{
			AlunoID:           alunoID,
			BimestreID:        bimestreID,
			Disciplina:        disciplina,
			MediaPonderada:    0, // Ou NaN, ou algum indicador de impossibilidade
			SomaPesos:         0,
			NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
		}, errors.New("soma dos pesos das notas é zero, média não pode ser calculada") // Ou retorna média 0 com aviso
	}

	mediaFinal := somaValoresPonderados / somaPesos
	return MediaInfo{
		AlunoID:           alunoID,
		BimestreID:        bimestreID,
		Disciplina:        disciplina,
		MediaPonderada:    mediaFinal,
		SomaPesos:         somaPesos,
		NotasConsideradas: notasDoAlunoNoBimestreDisciplina,
	}, nil
}

// EditarNota permite alterar campos de uma nota existente.
func EditarNota(idNota string, novoValor *float64, novoPeso *float64, novaDesc *string, novaDataStr *string) (models.Grade, error) {
	if idNota == "" {
		return models.Grade{}, errors.New("ID da nota é obrigatório para edição")
	}
	if ActiveGradeStore == nil {
		return models.Grade{}, errors.New("GradeStore não inicializado")
	}

	grade, err := ActiveGradeStore.GetGradeByID(idNota)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return models.Grade{}, fmt.Errorf("nota com ID '%s' não encontrada para edição: %w", idNota, err)
		}
		return models.Grade{}, fmt.Errorf("erro ao buscar nota para edição: %w", err)
	}

	algoAlterado := false
	if novoValor != nil {
		if *novoValor < 0 || *novoValor > 10 {
			return models.Grade{}, fmt.Errorf("novo valor da nota (%.2f) fora do intervalo permitido (0-10)", *novoValor)
		}
		if grade.Value != *novoValor {
			grade.Value = *novoValor
			algoAlterado = true
		}
	}
	if novoPeso != nil {
		if *novoPeso <= 0 {
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
		dataAvaliacao, errDate := time.Parse(layout, *novaDataStr)
		if errDate != nil {
			return models.Grade{}, fmt.Errorf("formato de nova data inválido: %s. Use dd-mm-aaaa: %w", *novaDataStr, errDate)
		}
		if !grade.Date.Equal(dataAvaliacao) {
			grade.Date = dataAvaliacao
			algoAlterado = true
		}
	}

	if !algoAlterado {
		return grade, errors.New("nenhuma alteração fornecida para a nota")
	}

	updatedGrade, err := ActiveGradeStore.SaveGrade(grade) // SaveGrade usa INSERT OR REPLACE
	if err != nil {
		return models.Grade{}, fmt.Errorf("erro ao salvar alterações da nota ID '%s': %w", idNota, err)
	}
	return updatedGrade, nil
}

// ExcluirNota remove uma nota do armazenamento.
func ExcluirNota(idNota string) error {
	if idNota == "" {
		return errors.New("ID da nota é obrigatório para exclusão")
	}
	if ActiveGradeStore == nil {
		return errors.New("GradeStore não inicializado")
	}
	err := ActiveGradeStore.DeleteGrade(idNota)
	if err != nil {
		// O store já pode retornar um erro formatado para "not found"
		return fmt.Errorf("erro ao excluir nota ID '%s': %w", idNota, err)
	}
	return nil
}
