package rotina

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vickgenda/internal/models" // Assuming this is the correct path to models
	// Task creation will eventually be needed for task generation,
	// but for model management, we only need the models.Routine struct.
	// "vickgenda/internal/commands/tarefa"
)

const dateTimeLayoutRotina = "2006-01-02 15:04"

// rotinasStore é o nosso banco de dados em memória para modelos de rotina.
var (
	rotinasStore = make(map[string]models.Routine)
	nextRotinaID = 1
	muRotinas    sync.Mutex // Mutex para proteger o acesso concorrente
)

// generateNewRotinaID gera um ID único para um novo modelo de rotina.
func generateNewRotinaID() string {
	id := fmt.Sprintf("routine-%d", nextRotinaID)
	nextRotinaID++
	return id
}

// isValidFrequency valida o formato da frequência.
// Exemplos válidos: "diaria", "semanal:seg,qua,sex", "mensal:15", "manual"
func isValidFrequency(freq string) bool {
	lowerFreq := strings.ToLower(freq)
	if lowerFreq == "diaria" || lowerFreq == "manual" {
		return true
	}
	parts := strings.Split(lowerFreq, ":")
	if len(parts) == 2 {
		freqType := parts[0]
		// freqValue := parts[1] // Value validation can be more complex
		if freqType == "semanal" || freqType == "mensal" {
			// Basic check, more detailed validation (e.g., valid days/dates) can be added.
			return len(parts[1]) > 0
		}
	}
	return false
}

// CriarModeloRotina adiciona um novo modelo de rotina.
// Args: nome, frequencia, descTarefa, prioridadeTarefa, tagsTarefaStr (comma-separated), proximaExecucaoStr (YYYY-MM-DD HH:MM)
func CriarModeloRotina(nome, frequencia, descTarefa string, prioridadeTarefa int, tagsTarefaStr string, proximaExecucaoStr string) (models.Routine, error) {
	muRotinas.Lock()
	defer muRotinas.Unlock()

	if strings.TrimSpace(nome) == "" {
		return models.Routine{}, errors.New("o nome do modelo de rotina é obrigatório")
	}
	if !isValidFrequency(frequencia) {
		return models.Routine{}, errors.New("formato de frequência inválido. Exemplos: 'diaria', 'semanal:seg,qua', 'mensal:1', 'manual'")
	}
	if strings.TrimSpace(descTarefa) == "" {
		return models.Routine{}, errors.New("a descrição modelo para tarefas é obrigatória")
	}

	var proximaExecucao time.Time
	var err error
	if strings.ToLower(frequencia) != "manual" {
		if proximaExecucaoStr == "" {
			// For automatic routines, proximaExecucao can be set to a default (e.g., now) or be required.
			// As per spec: "Se não especificado para rotinas automáticas, pode ser calculado..."
			// For this step, we'll make it simpler: if not manual, it can be zero, implying it needs processing.
			// Or, let's make it required for non-manual for now to ensure it's always there if needed by a processor.
			// Let's default it to time.Now() if not provided and not manual, to be adjusted by a scheduler later.
			proximaExecucao = time.Now() // Placeholder, scheduler should refine this.
		} else {
			proximaExecucao, err = time.Parse(dateTimeLayoutRotina, proximaExecucaoStr)
			if err != nil {
				return models.Routine{}, errors.New("formato de data/hora inválido para próxima execução. Use YYYY-MM-DD HH:MM")
			}
		}
	}


	if prioridadeTarefa <= 0 {
		prioridadeTarefa = 2 // Padrão: Média
	}

	var tagsTarefa []string
	if strings.TrimSpace(tagsTarefaStr) != "" {
		tagsTarefa = strings.Split(tagsTarefaStr, ",")
		for i, tag := range tagsTarefa {
			tagsTarefa[i] = strings.TrimSpace(tag)
		}
	}

	now := time.Now()
	novoModelo := models.Routine{
		ID:                generateNewRotinaID(),
		Name:              nome,
		Description:       "", // Description for the routine model itself, not the task. Can be added.
		Frequency:         frequencia,
		TaskDescription:   descTarefa,
		TaskPriority:      prioridadeTarefa,
		TaskTags:          tagsTarefa,
		NextRunTime:       proximaExecucao,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	rotinasStore[novoModelo.ID] = novoModelo
	return novoModelo, nil
}

// ListarModelosRotina retorna todos os modelos de rotina.
// Args: sortBy, sortOrder
func ListarModelosRotina(sortBy string, sortOrder string) ([]models.Routine, error) {
	muRotinas.Lock()
	defer muRotinas.Unlock()

	var result []models.Routine
	for _, modelo := range rotinasStore {
		result = append(result, modelo)
	}

	if sortBy == "" {
		sortBy = "nome" // Padrão
	}
	if sortOrder == "" {
		sortOrder = "asc" // Padrão
	}

	sort.SliceStable(result, func(i, j int) bool {
		r1 := result[i]
		r2 := result[j]
		var less bool
		switch strings.ToLower(sortBy) {
		case "frequencia":
			less = r1.Frequency < r2.Frequency
		case "proxima_execucao":
			less = r1.NextRunTime.Before(r2.NextRunTime)
		default: // "nome" ou qualquer outro
			less = r1.Name < r2.Name
		}
		if strings.ToLower(sortOrder) == "desc" {
			return !less
		}
		return less
	})

	return result, nil
}

// EditarModeloRotina atualiza um modelo de rotina existente.
func EditarModeloRotina(id, novoNome, novaFreq, novaDescTarefa string, novaPrioTarefa int, novasTagsTarefaStr, novaProxExecStr string) (models.Routine, error) {
	muRotinas.Lock()
	defer muRotinas.Unlock()

	modelo, existe := rotinasStore[id]
	if !existe {
		return models.Routine{}, fmt.Errorf("modelo de rotina com ID '%s' não encontrado", id)
	}

	updated := false
	if novoNome != "" {
		modelo.Name = novoNome
		updated = true
	}
	if novaFreq != "" {
		if !isValidFrequency(novaFreq) {
			return models.Routine{}, errors.New("formato de frequência inválido")
		}
		modelo.Frequency = novaFreq
		updated = true
		// If frequency changes to/from manual, NextRunTime might need adjustment.
		// For now, if it becomes non-manual and NextRunTime is zero, it might need to be set.
        if strings.ToLower(novaFreq) != "manual" && novaProxExecStr == "" && modelo.NextRunTime.IsZero() {
            // If changing to an automatic type and no new next run time is provided,
            // and current is zero, set it to now (to be processed by scheduler)
             modelo.NextRunTime = time.Now()
        } else if strings.ToLower(novaFreq) == "manual" {
            modelo.NextRunTime = time.Time{} // Manual routines don't have a next run time
        }
	}
	if novaDescTarefa != "" {
		modelo.TaskDescription = novaDescTarefa
		updated = true
	}
	if novaPrioTarefa > 0 {
		modelo.TaskPriority = novaPrioTarefa
		updated = true
	}
	if novasTagsTarefaStr != "" {
		var tags []string
		if strings.TrimSpace(novasTagsTarefaStr) != "" {
			tags = strings.Split(novasTagsTarefaStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}
		modelo.TaskTags = tags
		updated = true
	}
    if novaProxExecStr != "" {
        if strings.ToLower(modelo.Frequency) == "manual" && novaProxExecStr != "" {
            return models.Routine{}, errors.New("não é possível definir próxima execução para rotina manual")
        }
        newNextRun, err := time.Parse(dateTimeLayoutRotina, novaProxExecStr)
        if err != nil {
            return models.Routine{}, errors.New("formato de data/hora inválido para próxima execução")
        }
        modelo.NextRunTime = newNextRun
        updated = true
    } else if strings.ToLower(modelo.Frequency) == "manual" {
		// If no new next run time is provided, and the frequency is manual (or changed to manual)
		// ensure NextRunTime is zeroed out.
		if !modelo.NextRunTime.IsZero() {
			modelo.NextRunTime = time.Time{}
			updated = true
		}
	}


	if !updated {
		return models.Routine{}, errors.New("nenhuma alteração especificada")
	}

	modelo.UpdatedAt = time.Now()
	rotinasStore[id] = modelo
	return modelo, nil
}

// RemoverModeloRotina remove um modelo de rotina.
func RemoverModeloRotina(id string) error {
	muRotinas.Lock()
	defer muRotinas.Unlock()

	_, existe := rotinasStore[id]
	if !existe {
		return fmt.Errorf("modelo de rotina com ID '%s' não encontrado", id)
	}

	delete(rotinasStore, id)
	return nil
}

// GetModeloRotinaByID helper
func GetModeloRotinaByID(id string) (models.Routine, error) {
    muRotinas.Lock()
    defer muRotinas.Unlock()
    modelo, existe := rotinasStore[id]
    if !existe {
        return models.Routine{}, fmt.Errorf("modelo de rotina com ID '%s' não encontrado", id)
    }
    return modelo, nil
}

// LimparRotinasStore helper for tests
func LimparRotinasStore() {
	muRotinas.Lock()
	defer muRotinas.Unlock()
	rotinasStore = make(map[string]models.Routine)
	nextRotinaID = 1
}

// GerarTarefasFromModelo gera tarefas a partir de um modelo de rotina específico.
// dataBaseStr é opcional (formato YYYY-MM-DD), usada para substituir placeholders como {data}.
func GerarTarefasFromModelo(modeloID string, dataBaseStr string) ([]models.Task, error) {
	muRotinas.Lock() // Lock for reading the routine model
	modelo, existe := rotinasStore[modeloID]
	muRotinas.Unlock() // Unlock immediately after reading

	if !existe {
		return nil, fmt.Errorf("modelo de rotina com ID '%s' não encontrado", modeloID)
	}

	var dataBase time.Time
	var err error
	if dataBaseStr != "" {
		dataBase, err = time.Parse("2006-01-02", dataBaseStr)
		if err != nil {
			return nil, fmt.Errorf("formato de data inválido para data base: %w", err)
		}
	} else {
		dataBase = time.Now()
	}

	// Substituir placeholders na descrição da tarefa
	taskDesc := modelo.TaskDescription
	taskDesc = strings.ReplaceAll(taskDesc, "{data}", dataBase.Format("2006-01-02"))
	taskDesc = strings.ReplaceAll(taskDesc, "{nome_rotina}", modelo.Name)
	// Adicionar mais placeholders conforme necessário

	// Gerar a tarefa usando a função CriarTarefa do pacote tarefa
	// Tarefas geradas por rotina não terão um DueDate específico, a menos que
	// o modelo de rotina seja estendido para suportar um offset de prazo.
	// Tags são convertidas de []string para string separada por vírgulas.
	tagsStr := strings.Join(modelo.TaskTags, ",")

	// CriarTarefa é uma função que pode retornar erro, precisamos lidar com isso.
	// E como CriarTarefa já lida com sua própria concorrência (travando tarefasStore),
	// não precisamos de um lock global aqui para a criação de tarefa em si.
	novaTarefa, err := tarefa.CriarTarefa(taskDesc, "", modelo.TaskPriority, tagsStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar tarefa a partir do modelo '%s': %w", modeloID, err)
	}

	// Lógica para atualizar NextRunTime da rotina (simplificado por agora)
	// A atualização real de NextRunTime exigiria uma lógica de agendamento mais complexa
	// baseada na frequência da rotina.
	// Por exemplo, se a rotina for 'diaria', NextRunTime seria +24h.
	// Se for 'semanal:seg', seria a próxima segunda-feira.
	// Esta parte é complexa e pode ser responsabilidade de um "scheduler" dedicado.
	// Por enquanto, apenas demonstramos que a tarefa foi gerada.
	// Se quisermos atualizar NextRunTime, precisamos de um Lock novamente.
	/*
	if strings.ToLower(modelo.Frequency) != "manual" {
		muRotinas.Lock()
		// Aqui viria a lógica de cálculo do próximo NextRunTime.
		// Exemplo muito simples: adicionar 24h se for diária.
		// if strings.ToLower(modelo.Frequency) == "diaria" {
		//	 modelo.NextRunTime = modelo.NextRunTime.Add(24 * time.Hour)
		// }
		// rotinasStore[modeloID] = modelo // Salvar a atualização
		muRotinas.Unlock()
	}
	*/

	return []models.Task{novaTarefa}, nil
}
```
