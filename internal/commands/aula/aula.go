package aula

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"vickgenda/internal/models"
)

// Simulação de armazenamento em memória para aulas
var lessonsStore = make(map[string]models.Lesson)
var nextLessonID = 1

// generateID gera um ID simples e único para novas aulas.
func generateID() string {
	id := fmt.Sprintf("aula%03d", nextLessonID)
	nextLessonID++
	return id
}

// CriarAula adiciona uma nova aula.
func CriarAula(disciplina, topico, dataStr, horaStr, turma, plano, obs string) (models.Lesson, error) {
	if disciplina == "" || topico == "" || dataStr == "" || turma == "" {
		return models.Lesson{}, errors.New("disciplina, tópico, data e turma são obrigatórios")
	}

	layout := "02-01-2006"
	aulaData, err := time.Parse(layout, dataStr)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("formato de data inválido: %s. Use dd-mm-aaaa", dataStr)
	}

	if horaStr != "" {
		horaLayout := "15:04"
		aulaHora, err := time.Parse(horaLayout, horaStr)
		if err != nil {
			return models.Lesson{}, fmt.Errorf("formato de hora inválido: %s. Use hh:mm", horaStr)
		}
		aulaData = time.Date(aulaData.Year(), aulaData.Month(), aulaData.Day(), aulaHora.Hour(), aulaHora.Minute(), 0, 0, time.Local)
	}

	newID := generateID()
	lesson := models.Lesson{
		ID:           newID,
		Subject:      disciplina,
		Topic:        topico,
		Date:         aulaData,
		ClassID:      turma,
		Plan:         plano,
		Observations: obs,
	}
	lessonsStore[newID] = lesson
	return lesson, nil
}

// ListarAulas retorna uma lista de aulas, com filtros opcionais.
func ListarAulas(disciplina, turma, periodo, mes, ano string) ([]models.Lesson, error) {
	var result []models.Lesson
	var periodoInicio, periodoFim time.Time
	var err error

	if periodo != "" {
		partes := strings.Split(periodo, ":")
		if len(partes) != 2 {
			return nil, errors.New("formato de período inválido. Use <data_inicio>:<data_fim>")
		}
		layout := "02-01-2006"
		periodoInicio, err = time.Parse(layout, partes[0])
		if err != nil {
			return nil, fmt.Errorf("formato de data de início inválido: %s", partes[0])
		}
		periodoFim, err = time.Parse(layout, partes[1])
		if err != nil {
			return nil, fmt.Errorf("formato de data de fim inválido: %s", partes[1])
		}
		periodoFim = periodoFim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	var mesTime time.Time
	if mes != "" {
		layout := "01-2006" // mm-aaaa
		mesTime, err = time.Parse(layout, mes)
		if err != nil {
			return nil, fmt.Errorf("formato de mês inválido: %s. Use mm-aaaa", mes)
		}
	}

    var anoTime time.Time
    if ano != "" {
        layout := "2006" // aaaa
        anoTime, err = time.Parse(layout, ano)
        if err != nil {
            return nil, fmt.Errorf("formato de ano inválido: %s. Use aaaa", ano)
        }
    }

	for _, lesson := range lessonsStore {
		match := true
		if disciplina != "" && !strings.EqualFold(lesson.Subject, disciplina) {
			match = false
		}
		if turma != "" && !strings.EqualFold(lesson.ClassID, turma) {
			match = false
		}
		if periodo != "" && (lesson.Date.Before(periodoInicio) || lesson.Date.After(periodoFim)) {
			match = false
		}
		if mes != "" && (lesson.Date.Month() != mesTime.Month() || lesson.Date.Year() != mesTime.Year()) {
			match = false
		}
        if ano != "" && lesson.Date.Year() != anoTime.Year() {
            match = false
        }
		if match {
			result = append(result, lesson)
		}
	}
	return result, nil
}

// VerAula retorna os detalhes de uma aula específica pelo ID.
func VerAula(id string) (models.Lesson, error) {
	lesson, found := lessonsStore[id]
	if !found {
		return models.Lesson{}, fmt.Errorf("aula com ID '%s' não encontrada", id)
	}
	return lesson, nil
}

// EditarPlanoAula atualiza o plano e/ou observações de uma aula existente.
func EditarPlanoAula(id, novoPlano, novasObservacoes string) (models.Lesson, error) {
	lesson, found := lessonsStore[id]
	if !found {
		return models.Lesson{}, fmt.Errorf("aula com ID '%s' não encontrada para edição", id)
	}

	// These flags track if an actual change was intended by passing a non-empty string.
	intentToChangePlano := false
	if novoPlano != "" {
		intentToChangePlano = true
	}

	intentToChangeObs := false
	if novasObservacoes != "" {
		intentToChangeObs = true
	}

	// If neither field was intended to be changed (i.e., both inputs were empty strings,
	// which is how the test "nenhuma_alteracao" signals "no change desired for these fields"),
	// then return an error.
	if !intentToChangePlano && !intentToChangeObs {
		return lesson, errors.New("nenhuma alteração fornecida para plano ou observações")
	}

	// Apply changes if intended
	if intentToChangePlano {
		lesson.Plan = novoPlano // Set to new value (could be empty string to clear it)
	}
	if intentToChangeObs {
		lesson.Observations = novasObservacoes // Set to new value
	}

	lessonsStore[id] = lesson
	return lesson, nil
}

// ExcluirAula remove uma aula do armazenamento.
// Conforme especificação: vickgenda aula excluir <id_aula> [--confirmar]
// A flag --confirmar será tratada pela CLI. Esta função apenas exclui.
func ExcluirAula(id string) error {
    _, found := lessonsStore[id]
    if !found {
        return fmt.Errorf("aula com ID '%s' não encontrada para exclusão", id)
    }
    delete(lessonsStore, id)
    return nil
}

// LimparStoreAulas é uma função auxiliar para testes, para limpar o map entre testes.
func LimparStoreAulas() {
    lessonsStore = make(map[string]models.Lesson)
    nextLessonID = 1
}
