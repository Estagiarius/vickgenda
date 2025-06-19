package tarefa

import (
	"testing"
	"time"
	"vickgenda/internal/models"
	"strings"
)

// Helper to check if a slice of tasks contains a task with a specific ID
func containsTask(tasks []models.Task, id string) bool {
	for _, task := range tasks {
		if task.ID == id {
			return true
		}
	}
	return false
}


func TestCriarTarefa(t *testing.T) {
	LimparTarefasStore() // Clean up before test

	t.Run("Criação bem-sucedida", func(t *testing.T) {
		desc := "Nova tarefa de teste"
		prazo := "2024-12-31"
		prio := 1
		tags := "importante,teste"

		tarefa, err := CriarTarefa(desc, prazo, prio, tags)
		if err != nil {
			t.Fatalf("CriarTarefa falhou: %v", err)
		}
		if tarefa.Description != desc {
			t.Errorf("Esperado descrição '%s', obtido '%s'", desc, tarefa.Description)
		}
		if tarefa.DueDate.Format("2006-01-02") != prazo {
			t.Errorf("Esperado prazo '%s', obtido '%s'", prazo, tarefa.DueDate.Format("2006-01-02"))
		}
		if tarefa.Priority != prio {
			t.Errorf("Esperado prioridade %d, obtido %d", prio, tarefa.Priority)
		}
		if len(tarefa.Tags) != 2 || tarefa.Tags[0] != "importante" || tarefa.Tags[1] != "teste" {
			t.Errorf("Tags incorretas: esperado '[importante teste]', obtido '%v'", tarefa.Tags)
		}
		if tarefa.Status != "Pendente" {
			t.Errorf("Esperado status 'Pendente', obtido '%s'", tarefa.Status)
		}
	})

	t.Run("Descrição obrigatória", func(t *testing.T) {
		_, err := CriarTarefa("", "2024-12-31", 1, "tag")
		if err == nil {
			t.Error("Esperado erro para descrição vazia, mas não houve erro")
		} else if !strings.Contains(err.Error(), "descrição da tarefa é obrigatória") {
            t.Errorf("Mensagem de erro inesperada para descrição vazia: %v", err)
        }
	})

	t.Run("Formato de prazo inválido", func(t *testing.T) {
		_, err := CriarTarefa("Tarefa com prazo inválido", "31/12/2024", 1, "")
		if err == nil {
			t.Error("Esperado erro para formato de prazo inválido, mas não houve erro")
		} else if !strings.Contains(err.Error(), "formato de data inválido para --prazo") {
            t.Errorf("Mensagem de erro inesperada para prazo inválido: %v", err)
        }
	})

    t.Run("Prioridade padrão se não especificada ou zero", func(t *testing.T) {
        tarefa, err := CriarTarefa("Tarefa prioridade padrão", "", 0, "")
        if err != nil {
            t.Fatalf("CriarTarefa falhou: %v", err)
        }
        if tarefa.Priority != 2 { // Padrão é 2 (Média)
            t.Errorf("Esperado prioridade padrão 2, obtido %d", tarefa.Priority)
        }
    })
}

func TestListarTarefas(t *testing.T) {
	LimparTarefasStore()
	t1, _ := CriarTarefa("Tarefa A", "2024-01-10", 1, "alta")
	t2, _ := CriarTarefa("Tarefa B", "2024-01-15", 2, "media,teste")
	t3, _ := CriarTarefa("Tarefa C", "2024-01-05", 1, "alta,teste")
    _, _ = ConcluirTarefa(t3.ID) // t3 está concluída
    t3Concluida, _ := GetTarefaByID(t3.ID)


	t.Run("Listar todas", func(t *testing.T) {
		tarefas, err := ListarTarefas("", 0, "", "", "", "")
		if err != nil {
			t.Fatalf("ListarTarefas falhou: %v", err)
		}
		if len(tarefas) != 3 {
			t.Errorf("Esperado 3 tarefas, obtido %d", len(tarefas))
		}
	})

	t.Run("Filtrar por status Pendente", func(t *testing.T) {
		tarefas, err := ListarTarefas("Pendente", 0, "", "", "", "")
		if err != nil {
			t.Fatalf("ListarTarefas falhou: %v", err)
		}
		if len(tarefas) != 2 { // t1, t2
			t.Errorf("Esperado 2 tarefas pendentes, obtido %d", len(tarefas))
		}
        if containsTask(tarefas, t3Concluida.ID) {
            t.Error("Lista de pendentes não deveria incluir tarefa concluída t3")
        }
	})

	t.Run("Filtrar por status Concluída", func(t *testing.T) {
		tarefas, err := ListarTarefas("Concluída", 0, "", "", "", "")
		if err != nil {
			t.Fatalf("ListarTarefas falhou: %v", err)
		}
		if len(tarefas) != 1 {
			t.Errorf("Esperado 1 tarefa concluída, obtido %d", len(tarefas))
		}
        if !containsTask(tarefas, t3Concluida.ID) {
            t.Error("Lista de concluídas deveria incluir tarefa t3")
        }
	})

	t.Run("Filtrar por prioridade 1", func(t *testing.T) {
		tarefas, err := ListarTarefas("", 1, "", "", "", "")
		if err != nil {
			t.Fatalf("ListarTarefas falhou: %v", err)
		}
		if len(tarefas) != 2 { // t1, t3
			t.Errorf("Esperado 2 tarefas com prioridade 1, obtido %d", len(tarefas))
		}
	})

	t.Run("Filtrar por tag 'teste'", func(t *testing.T) {
		tarefas, err := ListarTarefas("", 0, "", "teste", "", "")
		if err != nil {
			t.Fatalf("ListarTarefas falhou: %v", err)
		}
		if len(tarefas) != 2 { // t2, t3
			t.Errorf("Esperado 2 tarefas com tag 'teste', obtido %d", len(tarefas))
		}
	})

    t.Run("Ordenar por prazo ascendente", func(t *testing.T) {
        tarefas, err := ListarTarefas("", 0, "", "", "prazo", "asc")
        if err != nil {
            t.Fatalf("ListarTarefas falhou: %v", err)
        }
        if len(tarefas) != 3 {
            t.Fatalf("Esperado 3 tarefas, obtido %d", len(tarefas))
        }
        // t3 (01-05), t1 (01-10), t2 (01-15)
        if tarefas[0].ID != t3.ID || tarefas[1].ID != t1.ID || tarefas[2].ID != t2.ID {
            t.Errorf("Ordem incorreta: esperado IDs %s, %s, %s; obtido %s, %s, %s", t3.ID, t1.ID, t2.ID, tarefas[0].ID, tarefas[1].ID, tarefas[2].ID)
        }
    })
}

func TestEditarTarefa(t *testing.T) {
	LimparTarefasStore()
	tarefaOriginal, _ := CriarTarefa("Tarefa Original", "2024-05-01", 2, "original")

	t.Run("Edição bem-sucedida", func(t *testing.T) {
		novaDesc := "Descrição Atualizada"
		novoPrazo := "2024-06-15"
		novaPrio := 1
		novoStatus := "Em Andamento"
		novasTags := "atualizada,importante"

		editada, err := EditarTarefa(tarefaOriginal.ID, novaDesc, novoPrazo, novaPrio, novoStatus, novasTags)
		if err != nil {
			t.Fatalf("EditarTarefa falhou: %v", err)
		}
		if editada.Description != novaDesc {
			t.Errorf("Descrição não atualizada corretamente")
		}
        if editada.DueDate.Format("2006-01-02") != novoPrazo {
            t.Errorf("Prazo não atualizado corretamente")
        }
		if editada.Priority != novaPrio {
			t.Errorf("Prioridade não atualizada corretamente")
		}
		if editada.Status != novoStatus {
			t.Errorf("Status não atualizado corretamente")
		}
		if len(editada.Tags) != 2 || editada.Tags[0] != "atualizada" || editada.Tags[1] != "importante" {
			t.Errorf("Tags não atualizadas corretamente")
		}
        if editada.UpdatedAt == tarefaOriginal.UpdatedAt {
            t.Error("UpdatedAt não foi modificado após edição")
        }
	})

	t.Run("Tarefa não encontrada", func(t *testing.T) {
		_, err := EditarTarefa("id-inexistente", "Nova Desc", "", 0, "", "")
		if err == nil {
			t.Error("Esperado erro para ID inexistente, mas não houve erro")
		} else if !strings.Contains(err.Error(), "não encontrada") {
             t.Errorf("Mensagem de erro inesperada para ID inexistente: %v", err)
        }
	})

	t.Run("Nenhuma alteração especificada", func(t *testing.T) {
		_, err := EditarTarefa(tarefaOriginal.ID, "", "", 0, "", "")
		if err == nil {
			t.Error("Esperado erro quando nenhuma alteração é especificada, mas não houve erro")
		} else if !strings.Contains(err.Error(), "nenhuma alteração especificada") {
            t.Errorf("Mensagem de erro inesperada: %v", err)
        }
	})
}

func TestConcluirTarefa(t *testing.T) {
	LimparTarefasStore()
	tarefaPendente, _ := CriarTarefa("Tarefa para concluir", "", 2, "")

	t.Run("Concluir tarefa pendente", func(t *testing.T) {
		concluida, err := ConcluirTarefa(tarefaPendente.ID)
		if err != nil {
			t.Fatalf("ConcluirTarefa falhou: %v", err)
		}
		if concluida.Status != "Concluída" {
			t.Errorf("Esperado status 'Concluída', obtido '%s'", concluida.Status)
		}
        if concluida.UpdatedAt == tarefaPendente.UpdatedAt {
            t.Error("UpdatedAt não foi modificado após conclusão")
        }
	})

	t.Run("Tentar concluir tarefa já concluída", func(t *testing.T) {
		_, err := ConcluirTarefa(tarefaPendente.ID) // Já foi concluída no sub-teste anterior
		if err == nil {
			t.Error("Esperado erro ao concluir tarefa já concluída, mas não houve erro")
		} else if !strings.Contains(err.Error(), "tarefa já está concluída") {
            t.Errorf("Mensagem de erro inesperada: %v", err)
        }
	})
}

func TestRemoverTarefa(t *testing.T) {
	LimparTarefasStore()
	tarefaParaRemover, _ := CriarTarefa("Tarefa para remover", "", 3, "")

	t.Run("Remoção bem-sucedida", func(t *testing.T) {
		err := RemoverTarefa(tarefaParaRemover.ID)
		if err != nil {
			t.Fatalf("RemoverTarefa falhou: %v", err)
		}
		_, errGet := GetTarefaByID(tarefaParaRemover.ID)
		if errGet == nil {
			t.Error("Tarefa ainda encontrada após remoção")
		}
	})

	t.Run("Tentar remover tarefa inexistente", func(t *testing.T) {
		err := RemoverTarefa("id-que-nao-existe")
		if err == nil {
			t.Error("Esperado erro ao remover tarefa inexistente, mas não houve erro")
		}
	})
}

```
