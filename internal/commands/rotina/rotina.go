package rotina

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vickgenda-cli/internal/models"
	"vickgenda-cli/internal/commands/tarefa"
)

// dateTimeLayoutRotina define o formato para parsing de data/hora para rotinas.
const dateTimeLayoutRotina = "2006-01-02 15:04"

// rotinasStore é o banco de dados em memória para modelos de rotina.
// Não exportado; acesso gerenciado pelas funções públicas.
var (
	rotinasStore = make(map[string]models.Routine)
	nextRotinaID = 1
	muRotinas    sync.Mutex // muRotinas protege o acesso concorrente ao rotinasStore e nextRotinaID.
)

// generateNewRotinaID gera um ID único para um novo modelo de rotina.
// Não exportada, detalhe de implementação. Ex: "routine-1".
func generateNewRotinaID() string {
	id := fmt.Sprintf("routine-%d", nextRotinaID)
	nextRotinaID++
	return id
}

// isValidFrequencyInternal valida o formato da string de frequência.
// Esta é uma função auxiliar interna.
// Exemplos válidos: "diaria", "semanal:seg,qua,sex", "mensal:15", "manual".
func isValidFrequencyInternal(freq string) bool {
	lowerFreq := strings.ToLower(freq)
	if lowerFreq == "diaria" || lowerFreq == "manual" {
		return true
	}
	parts := strings.Split(lowerFreq, ":")
	if len(parts) == 2 {
		freqType := parts[0]
		freqValue := parts[1]
		if (freqType == "semanal" || freqType == "mensal") && len(freqValue) > 0 {
			// Validação mais detalhada (dias válidos, formato do dia do mês) pode ser adicionada aqui.
			return true
		}
	}
	return false
}

// CriarModeloRotina adiciona um novo modelo de rotina ao sistema.
// nome: Nome descritivo para o modelo.
// frequencia: Define a recorrência ("diaria", "semanal:dias", "mensal:dia_do_mes", "manual").
// descTarefa: Modelo para a descrição das tarefas geradas (pode usar placeholders como {nome_rotina}, {data}).
// prioridadeTarefa: Prioridade padrão para tarefas geradas (1-Alta, 2-Média, 3-Baixa). Padrão 2 se <= 0.
// tagsTarefaStr: String de tags separadas por vírgula para as tarefas geradas.
// proximaExecucaoStr: Data/hora ("YYYY-MM-DD HH:MM") da primeira execução.
//                     Se frequência não for "manual" e este campo for vazio, NextRunTime é time.Now().
// Retorna o modelo de rotina criado ou um erro de validação.
func CriarModeloRotina(nome, frequencia, descTarefa string, prioridadeTarefa int, tagsTarefaStr string, proximaExecucaoStr string) (models.Routine, error) {
	muRotinas.Lock()
	defer muRotinas.Unlock()

	if strings.TrimSpace(nome) == "" {
		return models.Routine{}, errors.New("o nome do modelo de rotina é obrigatório")
	}
	if !isValidFrequencyInternal(frequencia) { // Alterado para função interna
		return models.Routine{}, errors.New("formato de frequência inválido. Exemplos: 'diaria', 'semanal:seg,qua', 'mensal:1', 'manual'")
	}
	if strings.TrimSpace(descTarefa) == "" {
		return models.Routine{}, errors.New("a descrição modelo para tarefas é obrigatória")
	}

	var proximaExecucao time.Time
	var err error
	// Rotinas manuais não têm NextRunTime por padrão.
	if strings.ToLower(frequencia) != "manual" {
		if proximaExecucaoStr == "" {
			// Para rotinas automáticas, se não especificado, NextRunTime é agora.
			// Um sistema de agendamento poderia refinar isso com base na frequência.
			proximaExecucao = time.Now()
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
		Description:       "", // Campo de descrição do modelo de rotina, pode ser adicionado como parâmetro.
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

// ListarModelosRotina retorna uma lista de todos os modelos de rotina existentes.
// sortBy: Campo para ordenação ("nome", "frequencia", "proxima_execucao"). Padrão: "nome".
// sortOrder: Ordem ("asc", "desc"). Padrão: "asc".
// Retorna uma lista de modelos ou um erro (atualmente sempre nil).
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
			// Rotinas com NextRunTime zero (ex: manuais) vêm depois.
			if r1.NextRunTime.IsZero() { return false }
			if r2.NextRunTime.IsZero() { return true }
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
// id: ID do modelo a ser editado.
// Os demais parâmetros são os novos valores; se vazios ou zero (para prioridade), não são alterados,
// exceto tagsTarefaStr que substitui as tags existentes (string vazia = sem tags).
// novaProxExecStr: Se fornecida e a rotina não for manual, atualiza NextRunTime.
//                  Se a frequência for alterada para "manual", NextRunTime é zerado.
// Retorna o modelo atualizado ou um erro se não encontrado, validação falhar, ou nenhuma alteração for feita.
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
		if !isValidFrequencyInternal(novaFreq) { // Alterado para função interna
			return models.Routine{}, errors.New("formato de frequência inválido")
		}
		modelo.Frequency = novaFreq
		updated = true
		// Ajustar NextRunTime se a frequência mudar para/de manual.
		if strings.ToLower(novaFreq) == "manual" {
			modelo.NextRunTime = time.Time{} // Zera para manual
		} else if modelo.NextRunTime.IsZero() && novaProxExecStr == "" {
			// Se tornou automática, NextRunTime era zero e não foi fornecido novo, default para Now.
			modelo.NextRunTime = time.Now()
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
	if novasTagsTarefaStr != "" { // Permitir limpar tags
		var tags []string
		trimmedTags := strings.TrimSpace(novasTagsTarefaStr)
		if trimmedTags != "" {
			tags = strings.Split(trimmedTags, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}
		modelo.TaskTags = tags
		updated = true
	}

    if novaProxExecStr != "" {
        if strings.ToLower(modelo.Frequency) == "manual" {
            return models.Routine{}, errors.New("não é possível definir próxima execução para rotina manual")
        }
        newNextRun, err := time.Parse(dateTimeLayoutRotina, novaProxExecStr)
        if err != nil {
            return models.Routine{}, errors.New("formato de data/hora inválido para próxima execução")
        }
        modelo.NextRunTime = newNextRun
        updated = true
    } else if strings.ToLower(modelo.Frequency) == "manual" {
		// Se frequência é manual (ou mudou para manual) e não foi fornecida nova data, garante NextRunTime zero.
		if !modelo.NextRunTime.IsZero() {
			modelo.NextRunTime = time.Time{}
			updated = true // Considera uma atualização se NextRunTime foi zerado.
		}
	}


	if !updated {
		return models.Routine{}, errors.New("nenhuma alteração especificada")
	}

	modelo.UpdatedAt = time.Now()
	rotinasStore[id] = modelo
	return modelo, nil
}

// RemoverModeloRotina remove um modelo de rotina do sistema.
// id: ID do modelo a ser removido.
// Retorna um erro se o modelo não for encontrado.
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

// GetModeloRotinaByID busca e retorna um modelo de rotina pelo seu ID.
// Função auxiliar, útil para testes ou acesso por outros pacotes.
// Retorna o modelo encontrado ou um erro se não existir.
func GetModeloRotinaByID(id string) (models.Routine, error) {
    muRotinas.Lock()
    defer muRotinas.Unlock()
    modelo, existe := rotinasStore[id]
    if !existe {
        return models.Routine{}, fmt.Errorf("modelo de rotina com ID '%s' não encontrado", id)
    }
    return modelo, nil
}

// LimparRotinasStore remove todos os modelos de rotina do armazenamento.
// Destinada primariamente para uso em testes.
func LimparRotinasStore() {
	muRotinas.Lock()
	defer muRotinas.Unlock()
	rotinasStore = make(map[string]models.Routine)
	nextRotinaID = 1
}

// GerarTarefasFromModelo cria tarefas com base em um modelo de rotina específico.
// modeloID: ID do modelo de rotina a ser usado.
// dataBaseStr: Data base opcional ("YYYY-MM-DD") para substituir placeholders como {data}.
//              Se vazia, usa a data atual.
// Retorna uma lista de tarefas criadas (atualmente sempre uma) ou um erro.
// A lógica de atualização de NextRunTime do modelo é simplificada e comentada,
// pois um agendador mais complexo seria necessário para o cálculo correto.
func GerarTarefasFromModelo(modeloID string, dataBaseStr string) ([]models.Task, error) {
	muRotinas.Lock() // Lock para ler o modelo de rotina.
	modelo, existe := rotinasStore[modeloID]
	muRotinas.Unlock() // Desbloqueia imediatamente após a leitura.

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

	// Substituir placeholders na descrição da tarefa.
	taskDesc := modelo.TaskDescription
	taskDesc = strings.ReplaceAll(taskDesc, "{data}", dataBase.Format("2006-01-02"))
	taskDesc = strings.ReplaceAll(taskDesc, "{nome_rotina}", modelo.Name)
	// Outros placeholders podem ser adicionados aqui.

	tagsStr := strings.Join(modelo.TaskTags, ",")

	// CriarTarefa lida com sua própria concorrência.
	novaTarefa, err := tarefa.CriarTarefa(taskDesc, "", modelo.TaskPriority, tagsStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar tarefa a partir do modelo '%s': %w", modeloID, err)
	}

	// A lógica de atualização de NextRunTime está comentada pois requer um agendador.
	/*
	if strings.ToLower(modelo.Frequency) != "manual" {
		muRotinas.Lock()
		// Lógica de cálculo do próximo NextRunTime (ex: modelo.NextRunTime.Add(24 * time.Hour))
		// rotinasStore[modeloID] = modelo // Salvar atualização
		muRotinas.Unlock()
	}
	*/

	return []models.Task{novaTarefa}, nil
}
