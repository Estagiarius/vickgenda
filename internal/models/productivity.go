package models

import "time"

// Task representa uma tarefa a ser realizada.
type Task struct {
	ID          string    // Identificador único da tarefa
	Description string    // Descrição da tarefa
	DueDate     time.Time // Data de vencimento da tarefa
	Priority    int       // Prioridade da tarefa (ex: 1-Alta, 2-Média, 3-Baixa)
	Status      string    // Status da tarefa (ex: Pendente, Em Andamento, Concluída)
	Tags        []string  // Etiquetas ou categorias para a tarefa
	CreatedAt   time.Time // Data de criação da tarefa
	UpdatedAt   time.Time // Data da última atualização da tarefa
}

// Event representa um evento na agenda.
type Event struct {
	ID          string    // Identificador único do evento
	Title       string    // Título do evento
	Description string    // Descrição detalhada do evento
	StartTime   time.Time // Data e hora de início do evento
	EndTime     time.Time // Data e hora de término do evento
	Location    string    // Local do evento (opcional)
	CreatedAt   time.Time // Data de criação do evento
	UpdatedAt   time.Time // Data da última atualização do evento
}

// Routine representa uma rotina que pode gerar tarefas recorrentes.
type Routine struct {
	ID                string   // Identificador único da rotina
	Name              string   // Nome da rotina (ex: "Preparar aula de Segunda")
	Description       string   // Descrição da rotina
	Frequency         string   // Frequência da rotina (ex: "daily", "weekly:Mon,Wed,Fri", "monthly:15", "cron:* * * * *")
	TaskDescription   string   // Modelo para a descrição das tarefas geradas
	TaskPriority      int      // Prioridade padrão para as tarefas geradas
	TaskTags          []string // Etiquetas padrão para as tarefas geradas
	NextRunTime       time.Time // Próxima vez que a rotina deve ser executada para gerar tarefas
	CreatedAt         time.Time // Data de criação da rotina
	UpdatedAt         time.Time // Data da última atualização da rotina
}
