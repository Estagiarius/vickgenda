package tarefa

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vickgenda-cli/internal/models"
)

// tarefasStore é o nosso banco de dados em memória para tarefas.
// Não é exportado, pois o acesso é gerenciado pelas funções públicas deste pacote.
var (
	tarefasStore = make(map[string]models.Task)
	nextTaskID   = 1
	mu           sync.Mutex // mu protege o acesso concorrente ao tarefasStore e nextTaskID.
)

// generateNewTaskID gera um ID único para uma nova tarefa de forma sequencial.
// Esta função não é exportada pois é um detalhe de implementação.
// Exemplo: "task-1", "task-2".
func generateNewTaskID() string {
	id := fmt.Sprintf("task-%d", nextTaskID)
	nextTaskID++
	return id
}

// CriarTarefa adiciona uma nova tarefa ao sistema de gerenciamento de tarefas.
// Requer uma descrição não vazia.
// dueDateStr deve estar no formato "YYYY-MM-DD"; se vazio, a tarefa não tem prazo.
// A prioridade, se não especificada (<=0), assume o valor padrão 2 (Média).
// tagsStr é uma string de tags separadas por vírgula (ex: "importante,trabalho").
// Retorna a tarefa criada e armazenada ou um erro se a validação dos campos falhar.
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
		Status:      "Pendente", // Status inicial padrão para novas tarefas.
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tarefasStore[novaTarefa.ID] = novaTarefa
	return novaTarefa, nil
}

// ListarTarefas retorna uma lista de tarefas com base nos filtros e ordenação fornecidos.
// Todos os parâmetros de filtro são opcionais.
// statusFilter: filtra tarefas pelo status (ex: "Pendente", "Concluída"). Case-insensitive.
// priorityFilter: filtra tarefas pela prioridade (ex: 1, 2, 3).
// dueDateFilterStr: filtra tarefas com prazo até a data especificada ("YYYY-MM-DD").
// tagFilter: filtra tarefas que contenham a tag especificada. Case-insensitive.
// sortBy: campo para ordenação ("descricao", "prazo", "prioridade", "status", "CreatedAt"). Padrão: "CreatedAt".
// sortOrder: ordem de classificação ("asc" para ascendente, "desc" para descendente). Padrão: "asc".
// Retorna uma lista de tarefas ou um erro se, por exemplo, o formato de data do filtro for inválido.
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

// EditarTarefa atualiza os campos de uma tarefa existente, identificada pelo seu ID.
// Pelo menos um dos campos a serem alterados (novaDesc, novoPrazoStr, etc.) deve ser fornecido.
// Se um campo de string opcional for vazio, ele não será alterado.
// Se novaPrioridade for 0 ou negativo, não será alterada.
// novasTagsStr substitui completamente as tags existentes; se vazia, as tags são mantidas ou limpas dependendo da interpretação desejada (aqui, string vazia de tags = sem tags).
// Retorna a tarefa atualizada ou um erro se a tarefa não for encontrada, nenhuma alteração for especificada, ou houver erro de formato.
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
	if novasTagsStr != "" { // Permitir limpar tags passando uma string que resulte em slice vazio, ex: "" ou " "
		var tags []string
		trimmedTagsStr := strings.TrimSpace(novasTagsStr)
		if trimmedTagsStr != "" { // Só faz split se não for vazia após trim
			tags = strings.Split(trimmedTagsStr, ",")
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

// ConcluirTarefa marca uma tarefa especificada pelo ID como "Concluída".
// Retorna a tarefa atualizada ou um erro se a tarefa não for encontrada ou já estiver concluída.
func ConcluirTarefa(id string) (models.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	tarefa, existe := tarefasStore[id]
	if !existe {
		return models.Task{}, fmt.Errorf("tarefa com ID '%s' não encontrada", id)
	}

	if tarefa.Status == "Concluída" {
		return tarefa, errors.New("tarefa já está concluída")
	}

	tarefa.Status = "Concluída"
	tarefa.UpdatedAt = time.Now()
	tarefasStore[id] = tarefa
	return tarefa, nil
}

// RemoverTarefa remove uma tarefa do sistema, identificada pelo seu ID.
// Retorna um erro se a tarefa não for encontrada.
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

// GetTarefaByID busca e retorna uma tarefa específica pelo seu ID.
// É uma função auxiliar que pode ser usada por outros pacotes ou para testes.
// Retorna a tarefa encontrada ou um erro se nenhuma tarefa com o ID fornecido existir.
func GetTarefaByID(id string) (models.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	tarefa, existe := tarefasStore[id]
	if !existe {
		return models.Task{}, fmt.Errorf("tarefa com ID '%s' não encontrada", id)
	}
	return tarefa, nil
}

// LimparTarefasStore remove todas as tarefas do armazenamento em memória.
// Esta função é primariamente destinada a ser usada em testes para garantir um estado limpo.
func LimparTarefasStore() {
	mu.Lock()
	defer mu.Unlock()
	tarefasStore = make(map[string]models.Task)
	nextTaskID = 1
}

// ContarTarefas retorna a contagem de tarefas com base nos filtros fornecidos.
// Filtros são opcionais. Se um filtro não for desejado, passe uma string vazia ou 0.
// statusFilter: filtra tarefas pelo status. Case-insensitive.
// priorityFilter: filtra tarefas pela prioridade.
// tagFilter: filtra tarefas que contenham a tag especificada. Case-insensitive.
// Retorna o número de tarefas que correspondem aos critérios ou um erro (atualmente sempre nil).
func ContarTarefas(statusFilter string, priorityFilter int, tagFilter string) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	count := 0
	for _, tarefa := range tarefasStore {
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
		count++
	}
	return count, nil
}
