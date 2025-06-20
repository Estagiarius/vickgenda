package models

import "time"

// Term representa um período de avaliação (ex: Bimestre).
type Term struct {
	ID            string    `json:"id"`                         // Identificador único do período
	Name          string    `json:"name"`                       // Nome do período (ex: "1º Bimestre")
	AcademicYear  string    `json:"academic_year"`              // Ano letivo do período (ex: "2023", "2024")
	StartDate     time.Time `json:"start_date"`                 // Data de início do período
	EndDate       time.Time `json:"end_date"`                   // Data de término do período
	CreatedAt     time.Time `json:"created_at"`                 // Timestamp da criação do período
	UpdatedAt     time.Time `json:"updated_at"`                 // Timestamp da última atualização do período
	// Outros campos relevantes podem ser adicionados aqui.
}

// Student representa um aluno.
type Student struct {
	ID          string    `json:"id"`                           // Identificador único do aluno (ex: matrícula)
	Name        string    `json:"name"`                         // Nome completo do aluno
	ClassID     string    `json:"class_id,omitempty"`           // ID da turma em que o aluno está matriculado
	Email       string    `json:"email,omitempty"`              // Email de contato do aluno (opcional)
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`      // Data de nascimento do aluno (opcional)
	CreatedAt   time.Time `json:"created_at"`                   // Timestamp da criação do registro do aluno
	UpdatedAt   time.Time `json:"updated_at"`                   // Timestamp da última atualização do registro do aluno
	// Outros campos como contato dos pais, etc.
	// podem ser adicionados conforme a necessidade.
}

// Lesson representa uma aula.
type Lesson struct {
	ID           string    `json:"id"`                        // Identificador único da aula
	Subject      string    `json:"subject"`                   // Disciplina da aula (ex: "Matemática")
	Topic        string    `json:"topic"`                     // Tópico da aula (ex: "Equações de 2º Grau")
	Date         time.Time `json:"date"`                      // Data e hora da aula
	ClassID      string    `json:"class_id"`                  // Identificador da turma para a qual a aula foi dada
	Plan         string    `json:"plan,omitempty"`            // Plano de aula detalhado
	Observations string    `json:"observations,omitempty"`    // Observações ou anotações sobre a aula
	CreatedAt    time.Time `json:"created_at"`                // Timestamp da criação do registro da aula
	UpdatedAt    time.Time `json:"updated_at"`                // Timestamp da última atualização do registro da aula
	// Pode-se adicionar um slice de StudentIDs se for necessário
	// rastrear a presença por aula, ou um campo para materiais didáticos.
}

// Grade representa uma nota atribuída a um aluno em uma avaliação específica.
type Grade struct {
	ID          string    `json:"id"`                         // Identificador único da nota
	StudentID   string    `json:"student_id"`                 // ID do aluno que recebeu a nota
	TermID      string    `json:"term_id"`                    // ID do período (bimestre) em que a nota foi atribuída
	Subject     string    `json:"subject"`                    // Disciplina referente à nota
	Description string    `json:"description,omitempty"`      // Descrição da avaliação (ex: "Prova Mensal", "Trabalho em Grupo")
	Value       float64   `json:"value"`                      // O valor da nota
	Weight      float64   `json:"weight,omitempty"`           // Peso da nota para cálculo da média ponderada (ex: 0.4 para 40%)
	Date        time.Time `json:"date"`                       // Data em que a nota foi atribuída
	CreatedAt   time.Time `json:"created_at"`                 // Timestamp da criação do registro de nota
	UpdatedAt   time.Time `json:"updated_at"`                 // Timestamp da última atualização do registro de nota
	// Poderia ter um campo para EvaluationID se as avaliações
	// fossem entidades separadas.
}

// Class representa uma turma.
type Class struct {
	ID           string    `json:"id"`                         // Identificador único da turma (ex: "class-1")
	Name         string    `json:"name"`                       // Nome da turma (ex: "3º Ano A - Manhã", "Extensivo Noturno")
	Level        string    `json:"level"`                      // Nível de ensino (ex: "Ensino Fundamental II", "Ensino Médio", "Cursinho")
	AcademicYear string    `json:"academic_year"`              // Ano letivo da turma (ex: "2024")
	TermIDs      []string  `json:"term_ids,omitempty"`         // IDs dos períodos letivos (Term) associados a esta turma
	SubjectIDs   []string  `json:"subject_ids,omitempty"`      // IDs das disciplinas (Subject) lecionadas para esta turma
	StudentIDs   []string  `json:"student_ids,omitempty"`      // IDs dos alunos matriculados nesta turma
	CreatedAt    time.Time `json:"created_at"`                 // Timestamp da criação da turma
	UpdatedAt    time.Time `json:"updated_at"`                 // Timestamp da última atualização da turma
}

// Subject representa uma disciplina ou matéria.
type Subject struct {
	ID          string    `json:"id"`                         // Identificador único da disciplina (ex: "subj-math")
	Name        string    `json:"name"`                       // Nome da disciplina (ex: "Matemática", "História do Brasil")
	Description string    `json:"description,omitempty"`      // Descrição/Ementa da disciplina (opcional)
	TeacherIDs  []string  `json:"teacher_ids,omitempty"`      // IDs dos professores que lecionam esta disciplina (pode ser um ou mais)
	CreatedAt   time.Time `json:"created_at"`                 // Timestamp da criação da disciplina
	UpdatedAt   time.Time `json:"updated_at"`                 // Timestamp da última atualização da disciplina
}

// Outras structs relacionadas à gestão acadêmica podem ser adicionadas aqui.
