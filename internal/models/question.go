package models

import "time"

// Question representa uma única questão no banco de questões.
// Todo o conteúdo de texto para fins de UI deve ser tratado pela camada de apresentação,
// esta struct foca no modelo de dados.
type Question struct {
	ID          string    `json:"id"`           // Identificador único (ex: UUID)
	Subject     string    `json:"subject"`      // Matéria (ex: "Matemática", "História")
	Topic       string    `json:"topic"`        // Tópico específico dentro da matéria (ex: "Álgebra", "Segunda Guerra Mundial")
	Difficulty  string    `json:"difficulty"`   // Nível de dificuldade (ex: "easy", "medium", "hard" - manter em inglês para consistência de código, UI tratará a tradução)
	QuestionText string   `json:"question_text"` // O texto real da questão
	AnswerOptions []string `json:"answer_options,omitempty"` // Respostas possíveis para múltipla escolha, vazio para outros tipos
	CorrectAnswers []string `json:"correct_answers"` // Resposta(s) correta(s)
	QuestionType string   `json:"question_type"`  // Tipo de questão (ex: "multiple_choice", "true_false", "essay" - manter em inglês, UI tratará a tradução)
	Source      string    `json:"source,omitempty"` // Opcional: Fonte da questão (ex: "Livro Didático A, Capítulo 5")
	Tags        []string  `json:"tags,omitempty"` // Opcional: Tags para busca mais refinada
	CreatedAt   time.Time `json:"created_at"`   // Timestamp de quando a questão foi criada
	LastUsedAt  time.Time `json:"last_used_at,omitempty"` // Timestamp de quando a questão foi usada pela última vez em uma prova
	Author      string    `json:"author,omitempty"`   // Opcional: Quem criou/adicionou esta questão
}

// Níveis de dificuldade (constantes de exemplo, poderia ser um enum ou definido em outro lugar)
// Os valores das constantes (e.g., "easy") permanecem em inglês para consistência no código.
// A interface do usuário (UI) será responsável por apresentar esses valores em pt-BR.
const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"
)

// Tipos de questão (constantes de exemplo)
// Os valores das constantes (e.g., "multiple_choice") permanecem em inglês para consistência no código.
// A interface do usuário (UI) será responsável por apresentar esses valores em pt-BR.
const (
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeTrueFalse      = "true_false"
	QuestionTypeEssay          = "essay"
	QuestionTypeShortAnswer    = "short_answer"
)
