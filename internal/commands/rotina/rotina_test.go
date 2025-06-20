package rotina

import (
	"strings"
	"testing"
	"time"

	"vickgenda-cli/internal/models"
	"vickgenda-cli/internal/commands/tarefa" // Needed for checking generated tasks
)

// Helper to check if a slice of routines contains a routine with a specific ID
func containsRoutine(routines []models.Routine, id string) bool {
	for _, r := range routines {
		if r.ID == id {
			return true
		}
	}
	return false
}

func TestCriarModeloRotina(t *testing.T) {
	LimparRotinasStore()

	t.Run("Criação bem-sucedida manual", func(t *testing.T) {
		nome := "Rotina Manual Teste"
		freq := "manual"
		descTarefa := "Tarefa gerada por rotina manual {nome_rotina}"
		prio := 1
		tags := "manual,teste"

		modelo, err := CriarModeloRotina(nome, freq, descTarefa, prio, tags, "") // ProximaExecucao vazia para manual
		if err != nil {
			t.Fatalf("CriarModeloRotina falhou: %v", err)
		}
		if modelo.Name != nome {
			t.Errorf("Esperado nome '%s', obtido '%s'", nome, modelo.Name)
		}
		if modelo.Frequency != freq {
			t.Errorf("Esperado frequência '%s', obtido '%s'", freq, modelo.Frequency)
		}
        if !modelo.NextRunTime.IsZero() {
            t.Errorf("Esperado NextRunTime zero para rotina manual, obtido %v", modelo.NextRunTime)
        }
	})

	t.Run("Criação bem-sucedida diária com próxima execução", func(t *testing.T) {
		nome := "Rotina Diária Teste"
		freq := "diaria"
		descTarefa := "Tarefa diária"
        proxExec := time.Now().Add(24 * time.Hour).Format(dateTimeLayoutRotina)

		modelo, err := CriarModeloRotina(nome, freq, descTarefa, 2, "", proxExec)
		if err != nil {
			t.Fatalf("CriarModeloRotina falhou: %v", err)
		}
		if modelo.Frequency != freq {
			t.Errorf("Esperado frequência '%s', obtido '%s'", freq, modelo.Frequency)
		}
        if modelo.NextRunTime.IsZero() {
            t.Errorf("Esperado NextRunTime não zero para rotina diária com data especificada")
        }
        parsedNextRun, _ := time.Parse(dateTimeLayoutRotina, proxExec)
        if !modelo.NextRunTime.Equal(parsedNextRun) {
             t.Errorf("Esperado NextRunTime %v, obtido %v", parsedNextRun, modelo.NextRunTime)
        }
	})

	t.Run("Criação bem-sucedida diária sem próxima execução (default time.Now())", func(t *testing.T) {
		nome := "Rotina Diária Default"
		freq := "diaria"
		descTarefa := "Tarefa diária default"

        // Chamando sem proxExecStr, NextRunTime deve ser time.Now() (aproximadamente)
		modelo, err := CriarModeloRotina(nome, freq, descTarefa, 2, "", "")
		if err != nil {
			t.Fatalf("CriarModeloRotina falhou: %v", err)
		}
        // Verifica se NextRunTime é recente (dentro de alguns segundos de Now)
        // já que o valor exato de time.Now() no momento da criação é difícil de prever.
		if time.Since(modelo.NextRunTime) > 5*time.Second {
			t.Errorf("Esperado NextRunTime recente para rotina diária sem data especificada, obtido %v", modelo.NextRunTime)
		}
	})

	t.Run("Nome obrigatório", func(t *testing.T) {
		_, err := CriarModeloRotina("", "manual", "Desc", 1, "", "")
		if err == nil || !strings.Contains(err.Error(), "nome do modelo de rotina é obrigatório") {
			t.Errorf("Esperado erro para nome vazio, obtido: %v", err)
		}
	})

	t.Run("Frequência inválida", func(t *testing.T) {
		_, err := CriarModeloRotina("Nome", "anual", "Desc", 1, "", "")
		if err == nil || !strings.Contains(err.Error(), "formato de frequência inválido") {
			t.Errorf("Esperado erro para frequência inválida, obtido: %v", err)
		}
	})

    t.Run("Descrição da tarefa obrigatória", func(t *testing.T) {
		_, err := CriarModeloRotina("Nome Valido", "manual", "", 1, "", "")
		if err == nil || !strings.Contains(err.Error(), "descrição modelo para tarefas é obrigatória") {
			t.Errorf("Esperado erro para descrição da tarefa vazia, obtido: %v", err)
		}
	})

    t.Run("Formato de próxima execução inválido", func(t *testing.T) {
		_, err := CriarModeloRotina("Nome", "diaria", "Desc", 1, "", "data invalida")
		if err == nil || !strings.Contains(err.Error(), "formato de data/hora inválido para próxima execução") {
			t.Errorf("Esperado erro para formato de próxima execução inválido, obtido: %v", err)
		}
	})
}

func TestListarModelosRotina(t *testing.T) {
	LimparRotinasStore()
	r1, _ := CriarModeloRotina("Rotina ZZZ", "manual", "Desc Z", 1, "", "")
	r2, _ := CriarModeloRotina("Rotina AAA", "diaria", "Desc A", 2, "", time.Now().Format(dateTimeLayoutRotina))

	t.Run("Listar todos", func(t *testing.T) {
		modelos, err := ListarModelosRotina("", "")
		if err != nil {
			t.Fatalf("ListarModelosRotina falhou: %v", err)
		}
		if len(modelos) != 2 {
			t.Errorf("Esperado 2 modelos, obtido %d", len(modelos))
		}
	})

	t.Run("Ordenar por nome ascendente (padrão)", func(t *testing.T) {
		modelos, err := ListarModelosRotina("nome", "asc")
		if err != nil {
			t.Fatalf("ListarModelosRotina falhou: %v", err)
		}
		if len(modelos) != 2 || modelos[0].ID != r2.ID || modelos[1].ID != r1.ID { // AAA antes de ZZZ
			t.Errorf("Ordenação por nome falhou. Esperado IDs %s, %s. Obtido %s, %s", r2.ID, r1.ID, modelos[0].ID, modelos[1].ID)
		}
	})
}

func TestEditarModeloRotina(t *testing.T) {
	LimparRotinasStore()
	original, _ := CriarModeloRotina("Original", "manual", "Desc Orig", 1, "tag1", "")

	t.Run("Edição bem-sucedida", func(t *testing.T) {
		novoNome := "Nome Editado"
		novaFreq := "diaria"
        novaProxExec := time.Now().Add(5 * time.Minute).Format(dateTimeLayoutRotina) // Precisa de prox exec para diaria

		editado, err := EditarModeloRotina(original.ID, novoNome, novaFreq, "", 0, "", novaProxExec)
		if err != nil {
			t.Fatalf("EditarModeloRotina falhou: %v", err)
		}
		if editado.Name != novoNome {
			t.Errorf("Nome não foi atualizado")
		}
		if editado.Frequency != novaFreq {
			t.Errorf("Frequência não foi atualizada")
		}
        if editado.NextRunTime.IsZero() && novaFreq != "manual" {
            t.Errorf("NextRunTime não deveria ser zero para frequência '%s'", novaFreq)
        }
	})

    t.Run("Mudar para manual zera NextRunTime", func(t *testing.T) {
        // Criar uma com NextRunTime
        comTempo, _ := CriarModeloRotina("Com Tempo", "diaria", "Desc", 1, "", time.Now().Format(dateTimeLayoutRotina))

        editado, err := EditarModeloRotina(comTempo.ID, "", "manual", "", 0, "", "")
        if err != nil {
            t.Fatalf("EditarModeloRotina falhou: %v", err)
        }
        if !editado.NextRunTime.IsZero() {
            t.Errorf("NextRunTime deveria ser zero após mudar para manual, obtido %v", editado.NextRunTime)
        }
    })

	t.Run("Modelo não encontrado", func(t *testing.T) {
		_, err := EditarModeloRotina("id-inexistente", "Novo Nome", "", "", 0, "", "")
		if err == nil || !strings.Contains(err.Error(), "não encontrado") {
			t.Errorf("Esperado erro para ID inexistente, obtido: %v", err)
		}
	})
}

func TestRemoverModeloRotina(t *testing.T) {
	LimparRotinasStore()
	modeloParaRemover, _ := CriarModeloRotina("Para Remover", "manual", "Desc", 1, "", "")

	t.Run("Remoção bem-sucedida", func(t *testing.T) {
		err := RemoverModeloRotina(modeloParaRemover.ID)
		if err != nil {
			t.Fatalf("RemoverModeloRotina falhou: %v", err)
		}
		_, errGet := GetModeloRotinaByID(modeloParaRemover.ID)
		if errGet == nil {
			t.Error("Modelo ainda encontrado após remoção")
		}
	})
}

func TestGerarTarefasFromModelo(t *testing.T) {
	LimparRotinasStore()
	tarefa.LimparTarefasStore() // Limpar também o store de tarefas

	modeloNome := "Rotina Geradora"
	dataBaseStr := "2024-03-15"
	taskDescTemplate := "Tarefa de {nome_rotina} para {data}"
	expectedTaskDesc := "Tarefa de Rotina Geradora para 2024-03-15"
	taskPrio := 1
	taskTags := "gerada,auto"

	modelo, _ := CriarModeloRotina(modeloNome, "manual", taskDescTemplate, taskPrio, taskTags, "")

	t.Run("Geração de tarefa bem-sucedida", func(t *testing.T) {
		tarefasGeradas, err := GerarTarefasFromModelo(modelo.ID, dataBaseStr)
		if err != nil {
			t.Fatalf("GerarTarefasFromModelo falhou: %v", err)
		}
		if len(tarefasGeradas) != 1 {
			t.Fatalf("Esperado 1 tarefa gerada, obtido %d", len(tarefasGeradas))
		}

		tarefaGerada := tarefasGeradas[0]
		if tarefaGerada.Description != expectedTaskDesc {
			t.Errorf("Descrição da tarefa gerada incorreta. Esperado '%s', obtido '%s'", expectedTaskDesc, tarefaGerada.Description)
		}
		if tarefaGerada.Priority != taskPrio {
			t.Errorf("Prioridade da tarefa gerada incorreta. Esperado %d, obtido %d", taskPrio, tarefaGerada.Priority)
		}
		if len(tarefaGerada.Tags) != 2 || tarefaGerada.Tags[0] != "gerada" || tarefaGerada.Tags[1] != "auto" {
			t.Errorf("Tags da tarefa gerada incorretas. Esperado '[gerada auto]', obtido '%v'", tarefaGerada.Tags)
		}

		// Verificar se a tarefa foi realmente adicionada ao store de tarefas
		_, errGet := tarefa.GetTarefaByID(tarefaGerada.ID)
		if errGet != nil {
			t.Errorf("Tarefa gerada não encontrada no store de tarefas: %v", errGet)
		}
	})

	t.Run("Modelo não encontrado para geração", func(t *testing.T) {
		_, err := GerarTarefasFromModelo("id-inexistente", dataBaseStr)
		if err == nil || !strings.Contains(err.Error(), "não encontrado") {
			t.Errorf("Esperado erro para modelo inexistente na geração, obtido: %v", err)
		}
	})

    t.Run("Formato de data base inválido para geração", func(t *testing.T) {
		_, err := GerarTarefasFromModelo(modelo.ID, "15/03/2024")
		if err == nil || !strings.Contains(err.Error(), "formato de data inválido para data base") {
			t.Errorf("Esperado erro para formato de data base inválido, obtido: %v", err)
		}
	})
}
