package models

import "time"

// Task representa uma tarefa a ser realizada.
// Contém todos os detalhes pertinentes a uma tarefa individual.
type Task struct {
	ID          string    // Identificador único da tarefa (ex: "task-1").
	Description string    // Descrição textual da tarefa.
	DueDate     time.Time // Data de vencimento da tarefa. Pode ser zero se não houver prazo.
	Priority    int       // Prioridade da tarefa (ex: 1-Alta, 2-Média, 3-Baixa).
	Status      string    // Status atual da tarefa (ex: "Pendente", "Em Andamento", "Concluída").
	Tags        []string  // Etiquetas ou categorias associadas à tarefa para facilitar a filtragem e organização.
	CreatedAt   time.Time // Timestamp da criação da tarefa.
	UpdatedAt   time.Time // Timestamp da última atualização da tarefa.
}

// Event representa um evento ou compromisso na agenda.
// Difere de uma tarefa por ter horários de início e fim definidos.
type Event struct {
	ID          string    // Identificador único do evento (ex: "event-1").
	Title       string    // Título breve do evento.
	Description string    // Descrição mais detalhada do evento (opcional).
	StartTime   time.Time // Data e hora de início do evento.
	EndTime     time.Time // Data e hora de término do evento.
	Location    string    // Local onde o evento ocorrerá (opcional).
	CreatedAt   time.Time // Timestamp da criação do evento.
	UpdatedAt   time.Time // Timestamp da última atualização do evento.
}

// Routine representa um modelo para a criação de tarefas recorrentes ou em massa.
// Permite definir um padrão para tarefas que precisam ser geradas periodicamente ou sob demanda.
type Routine struct {
	ID              string    // Identificador único do modelo de rotina (ex: "routine-1").
	Name            string    // Nome descritivo do modelo de rotina (ex: "Preparar relatório semanal").
	Description     string    // Descrição detalhada do propósito ou conteúdo do modelo de rotina (opcional).
	Frequency       string    // Define a recorrência da rotina (ex: "diaria", "semanal:seg,qua", "mensal:15", "manual").
	TaskDescription string    // Modelo para a descrição das tarefas que serão geradas por esta rotina. Pode conter placeholders como {data} ou {nome_rotina}.
	TaskPriority    int       // Prioridade padrão para as tarefas geradas por esta rotina.
	TaskTags        []string  // Etiquetas padrão para as tarefas geradas por esta rotina.
	NextRunTime     time.Time // Data e hora da próxima vez que a rotina deve ser processada para gerar tarefas (para rotinas automáticas).
	CreatedAt       time.Time // Timestamp da criação do modelo de rotina.
	UpdatedAt       time.Time // Timestamp da última atualização do modelo de rotina.
}
