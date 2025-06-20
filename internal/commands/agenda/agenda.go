package agenda

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vickgenda-cli/internal/models"
)

// dateTimeLayout define o formato padrão para parsing de data e hora (YYYY-MM-DD HH:MM).
const dateTimeLayout = "2006-01-02 15:04"
// dateLayout define o formato padrão para parsing de data (YYYY-MM-DD).
const dateLayout = "2006-01-02"

// eventosStore é o nosso banco de dados em memória para eventos.
// Não é exportado; o acesso é gerenciado pelas funções públicas.
var (
	eventosStore = make(map[string]models.Event)
	nextEventID  = 1
	muEventos    sync.Mutex // muEventos protege o acesso concorrente ao eventosStore e nextEventID.
)

// generateNewEventID gera um ID único para um novo evento de forma sequencial.
// Não exportada, pois é um detalhe de implementação. Ex: "event-1".
func generateNewEventID() string {
	id := fmt.Sprintf("event-%d", nextEventID)
	nextEventID++
	return id
}

// AdicionarEvento cria e armazena um novo evento na agenda.
// Requer título, data/hora de início e data/hora de término.
// inicioStr e fimStr devem estar no formato "YYYY-MM-DD HH:MM".
// A data/hora de término deve ser posterior à de início.
// Descrição e local são opcionais.
// Retorna o evento criado ou um erro se a validação falhar.
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

// ListarEventos retorna uma lista de eventos com base em filtros de período e ordenação.
// periodo: Define o intervalo de tempo ("dia", "semana", "mes", "proximos", "custom").
//          Se "custom", dataInicioStr e dataFimStr (YYYY-MM-DD) são obrigatórios.
//          Padrão é "proximos" se inválido ou vazio.
// sortBy: Campo para ordenação ("inicio", "titulo", "fim"). Padrão: "inicio".
// sortOrder: Ordem ("asc", "desc"). Padrão: "asc".
// Retorna uma lista de eventos filtrados e ordenados, ou um erro em caso de parâmetros inválidos.
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
		rangeStart = now // Início da semana como agora
		rangeEnd = now.AddDate(0, 0, 7) // Próximos 7 dias
	case "mes":
		rangeStart = now // Início do mês como agora
		rangeEnd = now.AddDate(0, 1, 0) // Próximos 30 dias (aproximadamente)
	case "proximos":
		rangeStart = now // A partir de agora
		rangeEnd = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC) // Um futuro distante
	case "custom":
		if dataInicioStr == "" || dataFimStr == "" {
			return nil, errors.New("para período customizado, forneça data de início e fim (YYYY-MM-DD)")
		}
		var err error
		rangeStart, err = time.Parse(dateLayout, dataInicioStr)
		if err != nil {
			return nil, errors.New("formato de data inválido para data de início. Use YYYY-MM-DD")
		}
		rangeEndTmp, err := time.Parse(dateLayout, dataFimStr)
		if err != nil {
			return nil, errors.New("formato de data inválido para data de fim. Use YYYY-MM-DD")
		}
		// Ajusta rangeEnd para o final do dia especificado.
		rangeEnd = time.Date(rangeEndTmp.Year(), rangeEndTmp.Month(), rangeEndTmp.Day(), 23, 59, 59, 999999999, rangeEndTmp.Location())
	default:
		// Padrão para "proximos" se período for inválido ou não especificado.
		rangeStart = now
		rangeEnd = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}


	for _, evento := range eventosStore {
		// Um evento está no intervalo se sua duração sobrepõe o intervalo [rangeStart, rangeEnd].
		// Condição: (event.StartTime <= rangeEnd) && (event.EndTime >= rangeStart)
		if evento.StartTime.Before(rangeEnd) && evento.EndTime.After(rangeStart) {
			result = append(result, evento)
		}
	}

	// Ordenação
	if sortBy == "" {
		sortBy = "inicio" // Padrão de ordenação
	}
	if sortOrder == "" {
		sortOrder = "asc" // Padrão de ordem
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

// VerDia lista todos os eventos que ocorrem em um dia específico.
// diaStr deve estar no formato "YYYY-MM-DD".
// Retorna uma lista de eventos para o dia, ordenados por hora de início, ou um erro.
func VerDia(diaStr string) ([]models.Event, error) {
	muEventos.Lock()
	defer muEventos.Unlock()

	dia, err := time.Parse(dateLayout, diaStr)
	if err != nil {
		return nil, errors.New("formato de data inválido. Use YYYY-MM-DD")
	}

	startOfDay := time.Date(dia.Year(), dia.Month(), dia.Day(), 0, 0, 0, 0, dia.Location())
	endOfDay := startOfDay.Add(24*time.Hour - 1*time.Nanosecond) // Até o último nanosegundo do dia.

	var result []models.Event
	for _, evento := range eventosStore {
		// Evento ocorre no dia se sua duração intercepta o dia.
		if evento.StartTime.Before(endOfDay) && evento.EndTime.After(startOfDay) {
			result = append(result, evento)
		}
	}

	// Ordenar por hora de início.
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})

	return result, nil
}


// EditarEvento atualiza os campos de um evento existente, identificado pelo seu ID.
// Pelo menos um dos campos opcionais (novoTitulo, novoInicioStr, etc.) deve ser fornecido para alteração.
// Se datas/horas forem alteradas, valida se o novo fim é posterior ao novo início.
// Retorna o evento atualizado ou um erro se não encontrado, nenhuma alteração especificada, ou formato inválido.
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

	// Usa os tempos atuais como base se não forem fornecidos novos tempos.
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

	// Valida a consistência dos tempos apenas se um deles foi alterado.
	if hasNewStartTime || hasNewEndTime {
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

// RemoverEvento remove um evento da agenda, identificado pelo seu ID.
// Retorna um erro se o evento não for encontrado.
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

// GetEventoByID busca e retorna um evento específico pelo seu ID.
// Função auxiliar, útil para testes ou acesso por outros pacotes.
// Retorna o evento encontrado ou um erro se não existir.
func GetEventoByID(id string) (models.Event, error) {
    muEventos.Lock()
    defer muEventos.Unlock()
    evento, existe := eventosStore[id]
    if !existe {
        return models.Event{}, fmt.Errorf("evento com ID '%s' não encontrado", id)
    }
    return evento, nil
}

// LimparEventosStore remove todos os eventos do armazenamento em memória.
// Destinada primariamente para uso em testes para garantir um estado limpo.
func LimparEventosStore() {
	muEventos.Lock()
	defer muEventos.Unlock()
	eventosStore = make(map[string]models.Event)
	nextEventID = 1
}

// ListarProximosXEventos retorna uma lista dos próximos 'count' eventos a partir de agora.
// Os eventos são ordenados pela data de início.
// Se 'count' for 0 ou negativo, ou se houver menos de 'count' eventos, todos os próximos eventos são retornados.
// Inclui eventos que já começaram mas ainda não terminaram (em andamento).
func ListarProximosXEventos(count int) ([]models.Event, error) {
	muEventos.Lock()
	defer muEventos.Unlock()

	var proximosEventos []models.Event
	now := time.Now()

	for _, evento := range eventosStore {
		// Inclui eventos que começam no futuro ou que já começaram mas ainda não terminaram.
		if evento.StartTime.After(now) || (now.After(evento.StartTime) && now.Before(evento.EndTime)) {
			proximosEventos = append(proximosEventos, evento)
		}
	}

	sort.SliceStable(proximosEventos, func(i, j int) bool {
		return proximosEventos[i].StartTime.Before(proximosEventos[j].StartTime)
	})

	if count > 0 && len(proximosEventos) > count {
		return proximosEventos[:count], nil
	}
	return proximosEventos, nil
}
