package models

import "time"

// Task representa uma tarefa a ser realizada.
// Contém todos os detalhes pertinentes a uma tarefa individual.
type Task struct {
	ID          string    `json:"id"`                       // Identificador único da tarefa (ex: "task-1").
	Description string    `json:"description"`              // Descrição textual da tarefa.
	DueDate     time.Time `json:"due_date,omitempty"`       // Data de vencimento da tarefa. Pode ser zero se não houver prazo.
	Priority    int       `json:"priority"`                 // Prioridade da tarefa (ex: 1-Alta, 2-Média, 3-Baixa).
	Status      string    `json:"status"`                   // Status atual da tarefa (ex: "Pendente", "Em Andamento", "Concluída").
	Tags        []string  `json:"tags,omitempty"`           // Etiquetas ou categorias associadas à tarefa para facilitar a filtragem e organização.
	CreatedAt   time.Time `json:"created_at"`               // Timestamp da criação da tarefa.
	UpdatedAt   time.Time `json:"updated_at"`               // Timestamp da última atualização da tarefa.
}

// TaskStatus constants
const (
	TaskStatusPending    = "Pendente"     // TaskStatusPending indica que a tarefa está pendente.
	TaskStatusInProgress = "Em Andamento" // TaskStatusInProgress indica que a tarefa está em andamento.
	TaskStatusCompleted  = "Concluída"    // TaskStatusCompleted indica que a tarefa foi concluída.
)

// Event representa um evento ou compromisso na agenda.
// Difere de uma tarefa por ter horários de início e fim definidos.
type Event struct {
	ID          string    `json:"id"`                         // Identificador único do evento (ex: "event-1").
	Title       string    `json:"title"`                      // Título breve do evento.
	Description string    `json:"description,omitempty"`      // Descrição mais detalhada do evento (opcional).
	StartTime   time.Time `json:"start_time"`                 // Data e hora de início do evento.
	EndTime     time.Time `json:"end_time"`                   // Data e hora de término do evento.
	Location    string    `json:"location,omitempty"`         // Local onde o evento ocorrerá (opcional).
	CreatedAt   time.Time `json:"created_at"`                 // Timestamp da criação do evento.
	UpdatedAt   time.Time `json:"updated_at"`                 // Timestamp da última atualização do evento.
}

// Routine representa um modelo para a criação de tarefas recorrentes ou em massa.
// Permite definir um padrão para tarefas que precisam ser geradas periodicamente ou sob demanda.
type Routine struct {
	ID              string    `json:"id"`                               // Identificador único do modelo de rotina.
	Name            string    `json:"name"`                             // Nome descritivo do modelo de rotina.
	Description     string    `json:"description,omitempty"`            // Descrição detalhada do propósito ou conteúdo do modelo de rotina.
	Frequency       string    `json:"frequency"`                        // Define a recorrência da rotina (ex: "diaria", "semanal:seg,qua", "mensal:15", "manual").
	TaskDescription string    `json:"task_description"`                 // Modelo para a descrição das tarefas que serão geradas por esta rotina.
	TaskPriority    int       `json:"task_priority"`                    // Prioridade padrão para as tarefas geradas.
	TaskTags        []string  `json:"task_tags,omitempty"`              // Etiquetas padrão para as tarefas geradas.
	NextRunTime     time.Time `json:"next_run_time,omitempty"`          // Data e hora da próxima execução da rotina.
	CreatedAt       time.Time `json:"created_at"`                       // Timestamp da criação do modelo de rotina.
	UpdatedAt       time.Time `json:"updated_at"`                       // Timestamp da última atualização do modelo de rotina.
}
