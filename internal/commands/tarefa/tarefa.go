package tarefa

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vickgenda/internal/models" // Assuming this is the correct path to models
)

// tarefasStore é o nosso banco de dados em memória para tarefas.
var (
	tarefasStore = make(map[string]models.Task)
	nextTaskID   = 1
	mu           sync.Mutex // Mutex para proteger o acesso concorrente ao store e nextTaskID
)

// generateNewTaskID gera um ID único para uma nova tarefa.
func generateNewTaskID() string {
	id := fmt.Sprintf("task-%d", nextTaskID)
	nextTaskID++
	return id
}

// CriarTarefa adiciona uma nova tarefa.
// Args: description, dueDateStr (YYYY-MM-DD), priority, tags (comma-separated string)
func CriarTarefa(description string, dueDateStr string, priority int, tagsStr string) (models.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	if strings.TrimSpace(description) == "" {
		return models.Task{}, errors.New("a descrição da tarefa é obrigatória")
	}

	var dueDate time.Time
	var err error
	if dueDateStr != "" {
		dueDate, err = time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			return models.Task{}, errors.New("formato de data inválido para --prazo. Use YYYY-MM-DD")
		}
	}

	if priority <= 0 {
		priority = 2 // Padrão: Média
	}

	var tags []string
	if strings.TrimSpace(tagsStr) != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	now := time.Now()
	novaTarefa := models.Task{
		ID:          generateNewTaskID(),
		Description: description,
		DueDate:     dueDate,
		Priority:    priority,
		Status:      "Pendente", // Status inicial
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tarefasStore[novaTarefa.ID] = novaTarefa
	return novaTarefa, nil
}

// ListarTarefas retorna uma lista de tarefas, com filtros e ordenação.
// Args: statusFilter, priorityFilter, dueDateFilter (YYYY-MM-DD), tagFilter, sortBy, sortOrder (asc/desc)
func ListarTarefas(statusFilter string, priorityFilter int, dueDateFilterStr string, tagFilter string, sortBy string, sortOrder string) ([]models.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	var result []models.Task
	for _, tarefa := range tarefasStore {
		// Aplicar filtros
		if statusFilter != "" && !strings.EqualFold(tarefa.Status, statusFilter) {
			continue
		}
		if priorityFilter > 0 && tarefa.Priority != priorityFilter {
			continue
		}
		if tagFilter != "" {
			foundTag := false
			for _, t := range tarefa.Tags {
				if strings.EqualFold(t, tagFilter) {
					foundTag = true
					break
				}
			}
			if !foundTag {
				continue
			}
		}
		if dueDateFilterStr != "" {
			dueDateF, err := time.Parse("2006-01-02", dueDateFilterStr)
			if err != nil {
				return nil, errors.New("formato de data inválido para filtro de prazo")
			}
			// Consider tasks due on or before the filter date, if the task has a due date
			if !tarefa.DueDate.IsZero() && tarefa.DueDate.After(dueDateF) {
				continue
			}
		}
		result = append(result, tarefa)
	}

	// Ordenação
	if sortBy == "" {
		sortBy = "CreatedAt" // Padrão
	}
	if sortOrder == "" {
		sortOrder = "asc" // Padrão
	}

	sort.SliceStable(result, func(i, j int) bool {
		t1 := result[i]
		t2 := result[j]
		var less bool
		switch strings.ToLower(sortBy) {
		case "descricao":
			less = t1.Description < t2.Description
		case "prazo":
			if t1.DueDate.IsZero() { // Tarefas sem prazo vêm depois
				return false
			}
			if t2.DueDate.IsZero() {
				return true
			}
			less = t1.DueDate.Before(t2.DueDate)
		case "prioridade":
			less = t1.Priority < t2.Priority // Menor número = maior prioridade
		case "status":
			less = t1.Status < t2.Status
		default: // "CreatedAt" ou qualquer outro
			less = t1.CreatedAt.Before(t2.CreatedAt)
		}
		if strings.ToLower(sortOrder) == "desc" {
			return !less
		}
		return less
	})

	return result, nil
}

// EditarTarefa atualiza uma tarefa existente.
// Pelo menos um dos campos opcionais (novaDesc, novoPrazoStr, etc.) deve ser fornecido.
func EditarTarefa(id string, novaDesc, novoPrazoStr string, novaPrioridade int, novoStatus string, novasTagsStr string) (models.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	tarefa, existe := tarefasStore[id]
	if !existe {
		return models.Task{}, fmt.Errorf("tarefa com ID '%s' não encontrada", id)
	}

	updated := false
	if novaDesc != "" {
		tarefa.Description = novaDesc
		updated = true
	}
	if novoPrazoStr != "" {
		newDueDate, err := time.Parse("2006-01-02", novoPrazoStr)
		if err != nil {
			return models.Task{}, errors.New("formato de data inválido para novo prazo. Use YYYY-MM-DD")
		}
		tarefa.DueDate = newDueDate
		updated = true
	}
	if novaPrioridade > 0 {
		tarefa.Priority = novaPrioridade
		updated = true
	}
	if novoStatus != "" {
		// Poderia haver uma validação de status aqui (ex: Pendente, Em Andamento, Concluída)
		tarefa.Status = novoStatus
		updated = true
	}
	if novasTagsStr != "" {
		var tags []string
		if strings.TrimSpace(novasTagsStr) != "" {
			tags = strings.Split(novasTagsStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}
		tarefa.Tags = tags // Substitui as tags existentes
		updated = true
	}

	if !updated {
		return models.Task{}, errors.New("nenhuma alteração especificada")
	}

	tarefa.UpdatedAt = time.Now()
	tarefasStore[id] = tarefa
	return tarefa, nil
}

// ConcluirTarefa marca uma tarefa como concluída.
func ConcluirTarefa(id string) (models.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	tarefa, existe := tarefasStore[id]
	if !existe {
		return models.Task{}, fmt.Errorf("tarefa com ID '%s' não encontrada", id)
	}

	if tarefa.Status == "Concluída" {
		return tarefa, errors.New("tarefa já está concluída") // Ou apenas retornar a tarefa sem erro
	}

	tarefa.Status = "Concluída"
	tarefa.UpdatedAt = time.Now()
	tarefasStore[id] = tarefa
	return tarefa, nil
}

// RemoverTarefa remove uma tarefa.
func RemoverTarefa(id string) error {
	mu.Lock()
	defer mu.Unlock()

	_, existe := tarefasStore[id]
	if !existe {
		return fmt.Errorf("tarefa com ID '%s' não encontrada", id)
	}

	delete(tarefasStore, id)
	return nil
}

// GetTarefaByID é uma função auxiliar para buscar uma tarefa (poderia ser usada por outros pacotes ou testes)
func GetTarefaByID(id string) (models.Task, error) {
    mu.Lock()
    defer mu.Unlock()

    tarefa, existe := tarefasStore[id]
    if !existe {
        return models.Task{}, fmt.Errorf("tarefa com ID '%s' não encontrada", id)
    }
    return tarefa, nil
}

// LimparTarefasStore é uma função auxiliar para testes, para limpar o store.
func LimparTarefasStore() {
	mu.Lock()
	defer mu.Unlock()
	tarefasStore = make(map[string]models.Task)
	nextTaskID = 1
}
```
