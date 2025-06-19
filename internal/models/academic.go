package models

import "time"

// Term representa um período de avaliação (ex: Bimestre).
type Term struct {
	ID        string    // Identificador único do período
	Name      string    // Nome do período (ex: "1º Bimestre")
	StartDate time.Time // Data de início do período
	EndDate   time.Time // Data de término do período
	// Outros campos relevantes podem ser adicionados aqui,
	// como o ano letivo a que este período pertence.
}

// Student representa um aluno.
type Student struct {
	ID   string // Identificador único do aluno (ex: matrícula)
	Name string // Nome completo do aluno
	// Outros campos como data de nascimento, contato dos pais, etc.
	// podem ser adicionados conforme a necessidade.
}

// Lesson representa uma aula.
type Lesson struct {
	ID          string    // Identificador único da aula
	Subject     string    // Disciplina da aula (ex: "Matemática")
	Topic       string    // Tópico da aula (ex: "Equações de 2º Grau")
	Date        time.Time // Data e hora da aula
	ClassID     string    // Identificador da turma para a qual a aula foi dada
	Plan        string    // Plano de aula detalhado
	Observations string    // Observações ou anotações sobre a aula
	// Pode-se adicionar um slice de StudentIDs se for necessário
	// rastrear a presença por aula, ou um campo para materiais didáticos.
}

// Grade representa uma nota atribuída a um aluno em uma avaliação específica.
type Grade struct {
	ID          string  // Identificador único da nota
	StudentID   string  // ID do aluno que recebeu a nota
	TermID      string  // ID do período (bimestre) em que a nota foi atribuída
	Subject     string  // Disciplina referente à nota
	Description string  // Descrição da avaliação (ex: "Prova Mensal", "Trabalho em Grupo")
	Value       float64 // O valor da nota
	Weight      float64 // Peso da nota para cálculo da média ponderada (ex: 0.4 para 40%)
	Date        time.Time // Data em que a nota foi atribuída
	// Poderia ter um campo para EvaluationID se as avaliações
	// fossem entidades separadas.
}

// Outras structs relacionadas à gestão acadêmica podem ser adicionadas aqui,
// como Class (Turma), Subject (Disciplina), etc.
