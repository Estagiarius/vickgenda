package models

import "time"

// Test representa uma prova no sistema.
// Em contextos de documentação ou comentários, pode ser referida como Prova.
type Test struct {
	ID                string            `json:"id"`                           // Identificador único da prova.
	Title             string            `json:"title"`                        // Título da prova.
	Subject           string            `json:"subject"`                      // Disciplina principal da prova.
	CreatedAt         time.Time         `json:"created_at"`                   // Data de criação da prova.
	Instructions      string            `json:"instructions,omitempty"`       // Instruções gerais para a prova.
	QuestionIDs       []string          `json:"question_ids,omitempty"`       // Lista ordenada dos IDs das questões incluídas na prova.
	LayoutOptions     map[string]string `json:"layout_options,omitempty"`     // Opções de formatação para a prova (ex: número de colunas).
	RandomizationSeed int64             `json:"randomization_seed,omitempty"` // Semente usada para randomização (se aplicável).
	UpdatedAt         time.Time         `json:"updated_at,omitempty"`         // Timestamp da última atualização da prova.
	PublishedAt       time.Time         `json:"published_at,omitempty"`       // Timestamp de quando a prova foi publicada/aplicada (pode ser zero).
	TermID            string            `json:"term_id,omitempty"`            // ID do período letivo (bimestre/semestre) ao qual esta prova está associada.
	AuthorID          string            `json:"author_id,omitempty"`          // ID do autor/professor que criou a prova.
}
