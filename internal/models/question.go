package models

import "time"

// Question representa uma única questão no banco de questões.
// Todo o conteúdo de texto para fins de UI deve ser tratado pela camada de apresentação,
// esta struct foca no modelo de dados.
type Question struct {
	ID             string    `json:"id"`           // Identificador único da questão.
	Subject        string    `json:"subject"`      // Matéria (disciplina) a que a questão pertence.
	Topic          string    `json:"topic"`        // Tópico específico dentro da matéria.
	Difficulty     string    `json:"difficulty"`   // Nível de dificuldade da questão.
	QuestionText   string    `json:"question_text"` // O texto integral da questão.
	AnswerOptions  []string  `json:"answer_options,omitempty"` // Opções de resposta para questões de múltipla escolha; pode ser vazio para outros tipos.
	CorrectAnswers []string  `json:"correct_answers"`          // Lista de respostas corretas. Para múltipla escolha, geralmente uma; para "selecione todas as aplicáveis", pode haver várias.
	QuestionType   string    `json:"question_type"`            // Tipo da questão (ex: múltipla escolha, verdadeiro/falso).
	Source         string    `json:"source,omitempty"`       // Fonte de onde a questão foi retirada (ex: livro, exame anterior).
	Tags           []string  `json:"tags,omitempty"`         // Etiquetas para categorização e busca.
	CreatedAt      time.Time `json:"created_at"`             // Timestamp da criação da questão.
	LastUsedAt     time.Time `json:"last_used_at,omitempty"` // Timestamp da última vez que a questão foi utilizada em uma prova.
	Author         string    `json:"author,omitempty"`       // Autor ou quem adicionou a questão ao banco.
}

// Níveis de dificuldade (constantes de exemplo, poderia ser um enum ou definido em outro lugar)
// Os valores das constantes (e.g., "easy") permanecem em inglês para consistência no código.
// A interface do usuário (UI) será responsável por apresentar esses valores em pt-BR.
const (
	DifficultyEasy   = "easy"   // DifficultyEasy representa o nível de dificuldade fácil.
	DifficultyMedium = "medium" // DifficultyMedium representa o nível de dificuldade médio.
	DifficultyHard   = "hard"   // DifficultyHard representa o nível de dificuldade difícil.
)

// Tipos de questão (constantes de exemplo)
// Os valores das constantes (e.g., "multiple_choice") permanecem em inglês para consistência no código.
// A interface do usuário (UI) será responsável por apresentar esses valores em pt-BR.
const (
	QuestionTypeMultipleChoice = "multiple_choice" // QuestionTypeMultipleChoice representa questões de múltipla escolha.
	QuestionTypeTrueFalse      = "true_false"      // QuestionTypeTrueFalse representa questões de verdadeiro ou falso.
	QuestionTypeEssay          = "essay"           // QuestionTypeEssay representa questões dissertativas.
	QuestionTypeShortAnswer    = "short_answer"    // QuestionTypeShortAnswer representa questões de resposta curta.
)

// FormatDifficultyToPtBR converte o valor de dificuldade para sua representação em pt-BR.
func FormatDifficultyToPtBR(difficulty string) string {
	switch difficulty {
	case DifficultyEasy:
		return "Fácil"
	case DifficultyMedium:
		return "Média"
	case DifficultyHard:
		return "Difícil"
	default:
		return difficulty // Retorna o valor original se não houver mapeamento.
	}
}

// FormatQuestionTypeToPtBR converte o tipo de questão para sua representação em pt-BR.
func FormatQuestionTypeToPtBR(qType string) string {
	switch qType {
	case QuestionTypeMultipleChoice:
		return "Múltipla Escolha"
	case QuestionTypeTrueFalse:
		return "Verdadeiro/Falso"
	case QuestionTypeEssay:
		return "Dissertativa"
	case QuestionTypeShortAnswer:
		return "Resposta Curta"
	default:
		return qType // Retorna o valor original se não houver mapeamento.
	}
}

// FormatLastUsedAt formata o timestamp LastUsedAt para exibição amigável.
// Retorna "Nunca utilizada" se a data for zero (não definida).
func FormatLastUsedAt(t time.Time) string {
	if t.IsZero() {
		return "Nunca utilizada"
	}
	return t.Format("02/01/2006 15:04") // Formato de data e hora comum em pt-BR.
}
