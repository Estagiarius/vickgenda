package commands_test // Using _test package to avoid import cycles if any, and to test as an external package might

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"vickgenda-cli/internal/commands/agenda"
	"vickgenda-cli/internal/commands/rotina"
	"vickgenda-cli/internal/commands/tarefa"
	"vickgenda-cli/internal/models"
)

const testLayoutDate = "2006-01-02"
const testLayoutDateTime = "2006-01-02 15:04"

// Helper function to clear all stores for Squad 2 modules
func cleanupSquad2Stores() {
	tarefa.LimparTarefasStore()
	agenda.LimparEventosStore()
	rotina.LimparRotinasStore()
}

func TestDashboardDataRetrievalScenario(t *testing.T) {
	cleanupSquad2Stores()

	// Setup: Create some data
	// Tasks
	todayStr := time.Now().Format(testLayoutDate)
	tomorrowStr := time.Now().Add(24 * time.Hour).Format(testLayoutDate)
	_, _ = tarefa.CriarTarefa("Tarefa Pendente 1", todayStr, 1, "dev")
	_, _ = tarefa.CriarTarefa("Tarefa Pendente 2", tomorrowStr, 2, "test")
	taskConcluida, _ := tarefa.CriarTarefa("Tarefa Concluida Hoje", todayStr, 1, "dev")
	_, _ = tarefa.ConcluirTarefa(taskConcluida.ID)
    _, _ = tarefa.CriarTarefa("Tarefa Pendente Antiga", time.Now().Add(-48*time.Hour).Format(testLayoutDate), 1, "old")


	// Events
	now := time.Now()
	eventTime1 := now.Add(1 * time.Hour)
	eventTime2 := now.Add(2 * time.Hour)
    eventTime3 := now.Add(3 * time.Hour) // For ListarProximosXEventos
    eventAmanha := now.Add(25 * time.Hour)


	_, _ = agenda.AdicionarEvento("Evento Hoje 1", eventTime1.Format(testLayoutDateTime), eventTime1.Add(30*time.Minute).Format(testLayoutDateTime), "Desc E1", "Local E1")
	_, _ = agenda.AdicionarEvento("Evento Hoje 2", eventTime2.Format(testLayoutDateTime), eventTime2.Add(1*time.Hour).Format(testLayoutDateTime), "Desc E2", "Local E2")
    _, _ = agenda.AdicionarEvento("Evento Hoje 3 ListarX", eventTime3.Format(testLayoutDateTime), eventTime3.Add(1*time.Hour).Format(testLayoutDateTime), "Desc E3", "Local E3")
    _, _ = agenda.AdicionarEvento("Evento Amanha", eventAmanha.Format(testLayoutDateTime), eventAmanha.Add(1*time.Hour).Format(testLayoutDateTime), "Desc EA", "Local EA")


	t.Run("Contar Tarefas Pendentes", func(t *testing.T) {
		count, err := tarefa.ContarTarefas("Pendente", 0, "")
		if err != nil {
			t.Fatalf("ContarTarefas falhou: %v", err)
		}
		// Esperado: "Tarefa Pendente 1", "Tarefa Pendente 2", "Tarefa Pendente Antiga" == 3
		if count != 3 {
			t.Errorf("Esperado 3 tarefas pendentes, obtido %d", count)
		}
	})

	t.Run("Contar Tarefas Pendentes com Prazo para Hoje", func(t *testing.T) {
		// Nota: A lógica de ListarTarefas com dueDateFilterStr é "ATÉ a data".
        // Para "exatamente hoje", o Squad 4 precisaria de uma lógica de filtragem adicional
        // ou uma função helper mais específica no backend.
        // Aqui, vamos simular o que o Squad 4 faria com a API atual.
		tarefasHoje, err := tarefa.ListarTarefas("Pendente", 0, todayStr, "", "", "")
		if err != nil {
			t.Fatalf("ListarTarefas para hoje falhou: %v", err)
		}

        countHojeExato := 0
        startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0,0,0,0, now.Location())
        endOfToday := startOfToday.Add(24*time.Hour)

        for _, task := range tarefasHoje {
            if !task.DueDate.IsZero() && (task.DueDate.Equal(startOfToday) || (task.DueDate.After(startOfToday) && task.DueDate.Before(endOfToday))) {
                 // E a tarefa é "Tarefa Pendente 1"
                if task.Description == "Tarefa Pendente 1" {
                    countHojeExato++
                }
            }
        }
		// Esperado: "Tarefa Pendente 1" (1 tarefa)
		if countHojeExato != 1 {
			t.Errorf("Esperado 1 tarefa pendente com prazo para hoje, obtido %d após filtragem manual", countHojeExato)
		}
	})

	t.Run("Listar Próximos 3 Eventos de Hoje (usando VerDia)", func(t *testing.T) {
		eventosHoje, err := agenda.VerDia(todayStr) // VerDia já ordena por StartTime
		if err != nil {
			t.Fatalf("agenda.VerDia falhou: %v", err)
		}
        // Esperado: "Evento Hoje 1", "Evento Hoje 2", "Evento Hoje 3 ListarX"
		if len(eventosHoje) != 3 {
			t.Errorf("Esperado 3 eventos para hoje, obtido %d", len(eventosHoje))
		}
        // Checar os nomes (assumindo ordem)
        if len(eventosHoje) > 0 && eventosHoje[0].Title != "Evento Hoje 1" {
            t.Errorf("Primeiro evento esperado 'Evento Hoje 1', obtido '%s'", eventosHoje[0].Title)
        }
        if len(eventosHoje) > 1 && eventosHoje[1].Title != "Evento Hoje 2" {
             t.Errorf("Segundo evento esperado 'Evento Hoje 2', obtido '%s'", eventosHoje[1].Title)
        }
        if len(eventosHoje) > 2 && eventosHoje[2].Title != "Evento Hoje 3 ListarX" {
             t.Errorf("Terceiro evento esperado 'Evento Hoje 3 ListarX', obtido '%s'", eventosHoje[2].Title)
        }
	})

    t.Run("Listar Próximos X Eventos (usando ListarProximosXEventos)", func(t *testing.T) {
        proximosEventos, err := agenda.ListarProximosXEventos(2) // Pegar os próximos 2
        if err != nil {
            t.Fatalf("ListarProximosXEventos falhou: %v", err)
        }
        if len(proximosEventos) != 2 {
            t.Errorf("Esperado 2 próximos eventos, obtido %d", len(proximosEventos))
        }
        if len(proximosEventos) > 0 && proximosEventos[0].Title != "Evento Hoje 1" {
             t.Errorf("Primeiro evento em ListarProximosXEventos incorreto: %s", proximosEventos[0].Title)
        }
    })
}

func TestManualRoutineTaskGenerationFlow(t *testing.T) {
	cleanupSquad2Stores()

    // 1. Squad 4 lista modelos de rotina (ou já tem o ID)
	modeloNome := "Minha Rotina Diária"
	descTemplate := "Revisar {nome_rotina} em {data}"
	modelo, err := rotina.CriarModeloRotina(modeloNome, "manual", descTemplate, 1, "trabalho", "")
	if err != nil {
		t.Fatalf("Falha ao criar modelo de rotina para teste: %v", err)
	}

    // 2. Usuário/Squad 4 aciona a geração de tarefas
    dataBaseGeracao := time.Now().Format(testLayoutDate)
	tarefasGeradas, err := rotina.GerarTarefasFromModelo(modelo.ID, dataBaseGeracao)
	if err != nil {
		t.Fatalf("GerarTarefasFromModelo falhou: %v", err)
	}
	if len(tarefasGeradas) != 1 {
		t.Fatalf("Esperado 1 tarefa gerada, obtido %d", len(tarefasGeradas))
	}
	novaTarefaGerada := tarefasGeradas[0]

    // 3. Squad 4 verifica/exibe a nova tarefa
    expectedDesc := fmt.Sprintf("Revisar %s em %s", modeloNome, dataBaseGeracao)
    if novaTarefaGerada.Description != expectedDesc {
        t.Errorf("Descrição da tarefa gerada incorreta. Esperado '%s', obtido '%s'", expectedDesc, novaTarefaGerada.Description)
    }

    // 4. Squad 4 pode querer listar todas as tarefas para atualizar a UI
    todasAsTarefas, err := tarefa.ListarTarefas("",0,"","","","")
    if err != nil {
        t.Fatalf("ListarTarefas após geração falhou: %v", err)
    }
    if len(todasAsTarefas) != 1 {
        t.Errorf("Esperado 1 tarefa no total após geração, obtido %d", len(todasAsTarefas))
    }
    if todasAsTarefas[0].ID != novaTarefaGerada.ID {
        t.Errorf("A tarefa listada não é a tarefa gerada.")
    }
}

func TestViewAndCompleteTaskFlow(t *testing.T) {
    cleanupSquad2Stores()

    // 1. Setup: Criar uma tarefa
    tarefaInicial, err := tarefa.CriarTarefa("Tarefa a ser concluída", time.Now().Format(testLayoutDate), 1, "teste")
    if err != nil {
        t.Fatalf("Falha ao criar tarefa inicial: %v", err)
    }

    // 2. Squad 4 busca a tarefa pelo ID (simulando clique do usuário)
    tarefaParaVer, err := tarefa.GetTarefaByID(tarefaInicial.ID)
    if err != nil {
        t.Fatalf("GetTarefaByID falhou: %v", err)
    }
    if tarefaParaVer.Status != "Pendente" {
        t.Errorf("Status inicial da tarefa deveria ser 'Pendente', obtido '%s'", tarefaParaVer.Status)
    }

    // 3. Squad 4 aciona a conclusão da tarefa
    tarefaConcluida, err := tarefa.ConcluirTarefa(tarefaParaVer.ID)
    if err != nil {
        t.Fatalf("ConcluirTarefa falhou: %v", err)
    }
    if tarefaConcluida.Status != "Concluída" {
        t.Errorf("Status da tarefa não mudou para 'Concluída'. Obtido: '%s'", tarefaConcluida.Status)
    }
    if tarefaConcluida.ID != tarefaInicial.ID {
        t.Errorf("ID da tarefa concluída não corresponde ao ID original.")
    }

    // 4. Squad 4 pode re-buscar ou usar o objeto retornado para atualizar a UI.
    // Verificar se a contagem de tarefas pendentes diminuiu.
    tarefasPendentesCount, err := tarefa.ContarTarefas("Pendente", 0, "")
    if err != nil {
        t.Fatalf("ContarTarefas (pendentes) falhou: %v", err)
    }
    if tarefasPendentesCount != 0 {
        t.Errorf("Esperado 0 tarefas pendentes após conclusão, obtido %d", tarefasPendentesCount)
    }
}
