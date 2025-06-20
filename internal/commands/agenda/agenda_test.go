package agenda

import (
	"strings"
	"testing"
	"time"
	"vickgenda-cli/internal/models"
)

// Helper to check if a slice of events contains an event with a specific ID
func containsEvent(events []models.Event, id string) bool {
	for _, event := range events {
		if event.ID == id {
			return true
		}
	}
	return false
}

func TestAdicionarEvento(t *testing.T) {
	LimparEventosStore()

	t.Run("Adição bem-sucedida", func(t *testing.T) {
		titulo := "Reunião de Planejamento"
		inicio := "2024-07-01 10:00"
		fim := "2024-07-01 11:00"
		desc := "Discutir próximos passos"
		local := "Sala 3"

		evento, err := AdicionarEvento(titulo, inicio, fim, desc, local)
		if err != nil {
			t.Fatalf("AdicionarEvento falhou: %v", err)
		}
		if evento.Title != titulo {
			t.Errorf("Esperado título '%s', obtido '%s'", titulo, evento.Title)
		}
		if evento.StartTime.Format(dateTimeLayout) != inicio {
			t.Errorf("Esperado início '%s', obtido '%s'", inicio, evento.StartTime.Format(dateTimeLayout))
		}
        if evento.EndTime.Format(dateTimeLayout) != fim {
			t.Errorf("Esperado fim '%s', obtido '%s'", fim, evento.EndTime.Format(dateTimeLayout))
		}
		if evento.Description != desc {
			t.Errorf("Esperado descrição '%s', obtido '%s'", desc, evento.Description)
		}
		if evento.Location != local {
			t.Errorf("Esperado local '%s', obtido '%s'", local, evento.Location)
		}
	})

	t.Run("Título obrigatório", func(t *testing.T) {
		_, err := AdicionarEvento("", "2024-07-01 10:00", "2024-07-01 11:00", "", "")
		if err == nil || !strings.Contains(err.Error(), "título do evento é obrigatório") {
			t.Errorf("Esperado erro para título vazio, obtido: %v", err)
		}
	})

	t.Run("Formato de início inválido", func(t *testing.T) {
		_, err := AdicionarEvento("Título", "01/07/2024 10:00", "2024-07-01 11:00", "", "")
		if err == nil || !strings.Contains(err.Error(), "formato de data/hora inválido para início") {
			t.Errorf("Esperado erro para formato de início inválido, obtido: %v", err)
		}
	})

	t.Run("Formato de fim inválido", func(t *testing.T) {
		_, err := AdicionarEvento("Título", "2024-07-01 10:00", "01/07/2024 11:00", "", "")
		if err == nil || !strings.Contains(err.Error(), "formato de data/hora inválido para término") {
			t.Errorf("Esperado erro para formato de fim inválido, obtido: %v", err)
		}
	})

	t.Run("Fim antes ou igual ao início", func(t *testing.T) {
		_, err := AdicionarEvento("Título", "2024-07-01 11:00", "2024-07-01 10:00", "", "")
		if err == nil || !strings.Contains(err.Error(), "hora de término deve ser posterior") {
			t.Errorf("Esperado erro para fim antes do início, obtido: %v", err)
		}
		_, err = AdicionarEvento("Título", "2024-07-01 10:00", "2024-07-01 10:00", "", "")
        if err == nil || !strings.Contains(err.Error(), "hora de término deve ser posterior") {
			t.Errorf("Esperado erro para fim igual ao início, obtido: %v", err)
		}
	})
}

func TestListarEventos(t *testing.T) {
	LimparEventosStore()
	now := time.Now()
	e1, _ := AdicionarEvento("Evento Futuro 1", now.Add(2*time.Hour).Format(dateTimeLayout), now.Add(3*time.Hour).Format(dateTimeLayout), "", "")
	e2, _ := AdicionarEvento("Evento Futuro 2", now.Add(24*time.Hour).Format(dateTimeLayout), now.Add(25*time.Hour).Format(dateTimeLayout), "", "")
	e3, _ := AdicionarEvento("Evento Passado", now.Add(-2*time.Hour).Format(dateTimeLayout), now.Add(-1*time.Hour).Format(dateTimeLayout), "", "")
    // Evento que abrange hoje
    todayStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
    todayEnd := todayStart.Add(1 * time.Hour)
    eToday, _ := AdicionarEvento("Evento de Hoje", todayStart.Format(dateTimeLayout), todayEnd.Format(dateTimeLayout), "", "")


	t.Run("Listar próximos (default)", func(t *testing.T) {
		eventos, err := ListarEventos("", "", "", "", "")
		if err != nil {
			t.Fatalf("ListarEventos falhou: %v", err)
		}
		// Esperado e1, e2, eToday (3 eventos)
		if len(eventos) != 3 {
			t.Errorf("Esperado 3 eventos próximos/atuais, obtido %d. Eventos: %+v", len(eventos), eventos)
		}
        if !containsEvent(eventos, e1.ID) || !containsEvent(eventos, e2.ID) || !containsEvent(eventos, eToday.ID) {
            t.Error("Lista de próximos não contém todos os eventos esperados")
        }
        if containsEvent(eventos, e3.ID) {
            t.Error("Lista de próximos não deveria conter evento passado e3")
        }
	})

	t.Run("Listar dia (hoje)", func(t *testing.T) {
		eventos, err := ListarEventos("dia", "", "", "", "")
		if err != nil {
			t.Fatalf("ListarEventos falhou: %v", err)
		}
        // Esperado eToday (e1 se a hora atual for próxima do evento)
        // Para ser mais preciso, o evento eToday deve estar aqui.
        // e1 pode ou não estar dependendo da hora exata do teste.
        // Vamos focar no eToday
		if !containsEvent(eventos, eToday.ID) {
			t.Errorf("Esperado evento de hoje (eToday) na lista do dia, mas não encontrado. Eventos: %+v", eventos)
		}
	})

    t.Run("Listar customizado", func(t *testing.T) {
        customStart := now.Add(-3 * time.Hour).Format(dateLayout) // Inclui e3 e eToday
        customEnd := now.Format(dateLayout)

		eventos, err := ListarEventos("custom", customStart, customEnd, "", "")
		if err != nil {
			t.Fatalf("ListarEventos falhou: %v", err)
		}
        // Esperado e3, eToday. e1 e e2 são no futuro.
		if len(eventos) != 2 {
			t.Errorf("Esperado 2 eventos no período customizado, obtido %d. Eventos: %+v", len(eventos), eventos)
		}
        if !containsEvent(eventos, e3.ID) || !containsEvent(eventos, eToday.ID) {
             t.Error("Período customizado não encontrou e3 ou eToday")
        }
	})
}

func TestVerDia(t *testing.T) {
	LimparEventosStore()
	now := time.Now()
	todayDateStr := now.Format(dateLayout)
	otherDateStr := now.AddDate(0,0,1).Format(dateLayout) // Amanhã

	ev1Today, _ := AdicionarEvento("Evento 1 Hoje", now.Format(dateTimeLayout), now.Add(1*time.Hour).Format(dateTimeLayout), "", "")
	AdicionarEvento("Evento Amanhã", now.AddDate(0,0,1).Format(dateTimeLayout), now.AddDate(0,0,1).Add(1*time.Hour).Format(dateTimeLayout), "", "")

	t.Run("Ver dia de hoje com evento", func(t *testing.T) {
		eventos, err := VerDia(todayDateStr)
		if err != nil {
			t.Fatalf("VerDia falhou: %v", err)
		}
		if len(eventos) != 1 {
			t.Errorf("Esperado 1 evento para hoje, obtido %d", len(eventos))
		}
        if !containsEvent(eventos, ev1Today.ID) {
            t.Error("Evento de hoje não encontrado")
        }
	})

	t.Run("Ver dia de amanhã sem evento (no store atual)", func(t *testing.T) {
        // Nota: o evento "Evento Amanhã" foi adicionado, então este teste espera 1 evento.
		eventos, err := VerDia(otherDateStr)
		if err != nil {
			t.Fatalf("VerDia falhou: %v", err)
		}
		if len(eventos) != 1 { // Deveria encontrar "Evento Amanhã"
			t.Errorf("Esperado 1 evento para amanhã, obtido %d", len(eventos))
		}
	})

	t.Run("Ver dia sem eventos (data distante)", func(t *testing.T) {
		eventos, err := VerDia("2099-01-01")
		if err != nil {
			t.Fatalf("VerDia falhou: %v", err)
		}
		if len(eventos) != 0 {
			t.Errorf("Esperado 0 eventos para data distante, obtido %d", len(eventos))
		}
	})
}


func TestEditarEvento(t *testing.T) {
	LimparEventosStore()
	original, _ := AdicionarEvento("Original", "2024-08-01 10:00", "2024-08-01 11:00", "Desc Original", "Local Original")

	t.Run("Edição bem-sucedida", func(t *testing.T) {
		novoTitulo := "Título Editado"
		novoInicio := "2024-08-01 14:00"
		novoFim := "2024-08-01 15:30"

		editado, err := EditarEvento(original.ID, novoTitulo, novoInicio, novoFim, "", "")
		if err != nil {
			t.Fatalf("EditarEvento falhou: %v", err)
		}
		if editado.Title != novoTitulo {
			t.Errorf("Título não foi atualizado")
		}
        if editado.StartTime.Format(dateTimeLayout) != novoInicio {
             t.Errorf("Hora de início não foi atualizada")
        }
        if editado.EndTime.Format(dateTimeLayout) != novoFim {
             t.Errorf("Hora de fim não foi atualizada")
        }
		if editado.UpdatedAt == original.UpdatedAt {
			t.Error("UpdatedAt não foi modificado")
		}
	})

	t.Run("Editar com fim antes do início", func(t *testing.T) {
		_, err := EditarEvento(original.ID, "", "2024-08-01 10:00", "2024-08-01 09:00", "", "")
		if err == nil || !strings.Contains(err.Error(), "hora de término deve ser posterior") {
			t.Errorf("Esperado erro para fim antes do início na edição, obtido: %v", err)
		}
	})

    t.Run("Evento não encontrado", func(t *testing.T) {
		_, err := EditarEvento("id-inexistente", "Novo Titulo", "", "", "", "")
		if err == nil || !strings.Contains(err.Error(), "não encontrado") {
			t.Errorf("Esperado erro para ID inexistente, obtido: %v", err)
		}
	})
}

func TestRemoverEvento(t *testing.T) {
	LimparEventosStore()
	eventoParaRemover, _ := AdicionarEvento("Para Remover", "2024-09-01 10:00", "2024-09-01 11:00", "", "")

	t.Run("Remoção bem-sucedida", func(t *testing.T) {
		err := RemoverEvento(eventoParaRemover.ID)
		if err != nil {
			t.Fatalf("RemoverEvento falhou: %v", err)
		}
		_, errGet := GetEventoByID(eventoParaRemover.ID)
		if errGet == nil {
			t.Error("Evento ainda encontrado após remoção")
		}
	})

	t.Run("Tentar remover evento inexistente", func(t *testing.T) {
		err := RemoverEvento("id-que-nao-existe")
		if err == nil {
			t.Error("Esperado erro ao remover evento inexistente, mas não houve erro")
		}
	})
}
