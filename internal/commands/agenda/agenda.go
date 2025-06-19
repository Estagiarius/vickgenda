package agenda

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vickgenda/internal/models" // Assuming this is the correct path to models
)

const dateTimeLayout = "2006-01-02 15:04"
const dateLayout = "2006-01-02"

// eventosStore é o nosso banco de dados em memória para eventos.
var (
	eventosStore = make(map[string]models.Event)
	nextEventID  = 1
	muEventos    sync.Mutex // Mutex para proteger o acesso concorrente ao store e nextEventID
)

// generateNewEventID gera um ID único para um novo evento.
func generateNewEventID() string {
	id := fmt.Sprintf("event-%d", nextEventID)
	nextEventID++
	return id
}

// AdicionarEvento adiciona um novo evento à agenda.
// Args: titulo, inicioStr (YYYY-MM-DD HH:MM), fimStr (YYYY-MM-DD HH:MM), descricao, local
func AdicionarEvento(titulo string, inicioStr string, fimStr string, descricao string, local string) (models.Event, error) {
	muEventos.Lock()
	defer muEventos.Unlock()

	if strings.TrimSpace(titulo) == "" {
		return models.Event{}, errors.New("o título do evento é obrigatório")
	}
	if strings.TrimSpace(inicioStr) == "" {
		return models.Event{}, errors.New("a data e hora de início são obrigatórias")
	}
	if strings.TrimSpace(fimStr) == "" {
		return models.Event{}, errors.New("a data e hora de término são obrigatórias")
	}

	inicio, err := time.Parse(dateTimeLayout, inicioStr)
	if err != nil {
		return models.Event{}, errors.New("formato de data/hora inválido para início. Use YYYY-MM-DD HH:MM")
	}
	fim, err := time.Parse(dateTimeLayout, fimStr)
	if err != nil {
		return models.Event{}, errors.New("formato de data/hora inválido para término. Use YYYY-MM-DD HH:MM")
	}

	if !fim.After(inicio) {
		return models.Event{}, errors.New("a hora de término deve ser posterior à hora de início")
	}

	now := time.Now()
	novoEvento := models.Event{
		ID:          generateNewEventID(),
		Title:       titulo,
		Description: descricao,
		StartTime:   inicio,
		EndTime:     fim,
		Location:    local,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	eventosStore[novoEvento.ID] = novoEvento
	return novoEvento, nil
}

// ListarEventos retorna uma lista de eventos, com filtros e ordenação.
// Args: periodo ("dia", "semana", "mes", "proximos", "custom"), dataInicioStr, dataFimStr, sortBy, sortOrder
func ListarEventos(periodo string, dataInicioStr string, dataFimStr string, sortBy string, sortOrder string) ([]models.Event, error) {
	muEventos.Lock()
	defer muEventos.Unlock()

	var result []models.Event
	now := time.Now()
	var rangeStart, rangeEnd time.Time

	switch strings.ToLower(periodo) {
	case "dia":
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
		rangeStart = startOfDay
		rangeEnd = endOfDay
	case "semana":
		rangeStart = now
		rangeEnd = now.AddDate(0, 0, 7)
	case "mes":
		rangeStart = now
		rangeEnd = now.AddDate(0, 1, 0)
	case "proximos":
		rangeStart = now
		rangeEnd = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC) // Far future
	case "custom":
		if dataInicioStr == "" || dataFimStr == "" {
			return nil, errors.New("para período customizado, forneça data de início e fim (YYYY-MM-DD)")
		}
		var err error
		rangeStart, err = time.Parse(dateLayout, dataInicioStr)
		if err != nil {
			return nil, errors.New("formato de data inválido para data de início. Use YYYY-MM-DD")
		}
		rangeEnd, err = time.Parse(dateLayout, dataFimStr)
		if err != nil {
			return nil, errors.New("formato de data inválido para data de fim. Use YYYY-MM-DD")
		}
		// Adjust rangeEnd to be end of the day
		rangeEnd = time.Date(rangeEnd.Year(), rangeEnd.Month(), rangeEnd.Day(), 23, 59, 59, 999999999, rangeEnd.Location())

	default:
		// Default to "proximos" if period is empty or invalid
		rangeStart = now
		rangeEnd = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}


	for _, evento := range eventosStore {
		// Event is within the range if (event.StartTime <= rangeEnd) and (event.EndTime >= rangeStart)
		if evento.StartTime.Before(rangeEnd) && evento.EndTime.After(rangeStart) {
			result = append(result, evento)
		}
	}

	// Ordenação
	if sortBy == "" {
		sortBy = "inicio" // Padrão
	}
	if sortOrder == "" {
		sortOrder = "asc" // Padrão
	}

	sort.SliceStable(result, func(i, j int) bool {
		e1 := result[i]
		e2 := result[j]
		var less bool
		switch strings.ToLower(sortBy) {
		case "titulo":
			less = e1.Title < e2.Title
		case "fim":
			less = e1.EndTime.Before(e2.EndTime)
		default: // "inicio" ou qualquer outro
			less = e1.StartTime.Before(e2.StartTime)
		}
		if strings.ToLower(sortOrder) == "desc" {
			return !less
		}
		return less
	})

	return result, nil
}

// VerDia lista todos os eventos de um dia específico.
// Arg: diaStr (YYYY-MM-DD)
func VerDia(diaStr string) ([]models.Event, error) {
	muEventos.Lock()
	defer muEventos.Unlock()

	dia, err := time.Parse(dateLayout, diaStr)
	if err != nil {
		return nil, errors.New("formato de data inválido. Use YYYY-MM-DD")
	}

	startOfDay := time.Date(dia.Year(), dia.Month(), dia.Day(), 0, 0, 0, 0, dia.Location())
	endOfDay := startOfDay.Add(24*time.Hour - 1*time.Nanosecond)

	var result []models.Event
	for _, evento := range eventosStore {
		if evento.StartTime.Before(endOfDay) && evento.EndTime.After(startOfDay) {
			result = append(result, evento)
		}
	}

	// Ordenar por hora de início
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})

	return result, nil
}


// EditarEvento atualiza um evento existente.
func EditarEvento(id string, novoTitulo, novoInicioStr, novoFimStr, novaDesc, novoLocal string) (models.Event, error) {
	muEventos.Lock()
	defer muEventos.Unlock()

	evento, existe := eventosStore[id]
	if !existe {
		return models.Event{}, fmt.Errorf("evento com ID '%s' não encontrado", id)
	}

	updated := false
	if novoTitulo != "" {
		evento.Title = novoTitulo
		updated = true
	}
	if novaDesc != "" {
		evento.Description = novaDesc
		updated = true
	}
	if novoLocal != "" {
		evento.Location = novoLocal
		updated = true
	}

	var novoInicio, novoFim time.Time
	var err error
	hasNewStartTime := novoInicioStr != ""
	hasNewEndTime := novoFimStr != ""

	currentStartTime := evento.StartTime
	currentEndTime := evento.EndTime

	if hasNewStartTime {
		novoInicio, err = time.Parse(dateTimeLayout, novoInicioStr)
		if err != nil {
			return models.Event{}, errors.New("formato de data/hora inválido para novo início. Use YYYY-MM-DD HH:MM")
		}
		updated = true
	} else {
		novoInicio = currentStartTime
	}

	if hasNewEndTime {
		novoFim, err = time.Parse(dateTimeLayout, novoFimStr)
		if err != nil {
			return models.Event{}, errors.New("formato de data/hora inválido para novo fim. Use YYYY-MM-DD HH:MM")
		}
		updated = true
	} else {
		novoFim = currentEndTime
	}

	if hasNewStartTime || hasNewEndTime { // Only validate if one of them changed
	    if !novoFim.After(novoInicio) {
		    return models.Event{}, errors.New("a nova hora de término deve ser posterior à nova hora de início")
	    }
    }
    evento.StartTime = novoInicio
	evento.EndTime = novoFim


	if !updated {
		return models.Event{}, errors.New("nenhuma alteração especificada")
	}

	evento.UpdatedAt = time.Now()
	eventosStore[id] = evento
	return evento, nil
}

// RemoverEvento remove um evento.
func RemoverEvento(id string) error {
	muEventos.Lock()
	defer muEventos.Unlock()

	_, existe := eventosStore[id]
	if !existe {
		return fmt.Errorf("evento com ID '%s' não encontrado", id)
	}

	delete(eventosStore, id)
	return nil
}

// GetEventoByID é uma função auxiliar
func GetEventoByID(id string) (models.Event, error) {
    muEventos.Lock()
    defer muEventos.Unlock()
    evento, existe := eventosStore[id]
    if !existe {
        return models.Event{}, fmt.Errorf("evento com ID '%s' não encontrado", id)
    }
    return evento, nil
}

// LimparEventosStore é uma função auxiliar para testes.
func LimparEventosStore() {
	muEventos.Lock()
	defer muEventos.Unlock()
	eventosStore = make(map[string]models.Event)
	nextEventID = 1
}

```
