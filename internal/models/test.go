package models

import "time"

// Test representa uma prova no sistema.
// Em contextos de documentação ou comentários, pode ser referida como Prova.
type Test struct {
	// ID é o identificador único da prova.
	ID string
	// Title é o título da prova.
	Title string
	// Subject é a disciplina principal da prova.
	Subject string
	// CreatedAt é a data de criação da prova.
	CreatedAt time.Time
	// Instructions são as instruções gerais para a prova.
	Instructions string
	// QuestionIDs é uma lista ordenada dos IDs das questões incluídas na prova.
	QuestionIDs []string
	// LayoutOptions são opções de formatação para a prova.
	// Exemplo: "columns": "2", "header": "School Name".
	// Utiliza um mapa para flexibilidade.
	LayoutOptions map[string]string
	// RandomizationSeed é a semente utilizada se a ordem das questões ou alternativas foi randomizada.
	// Guardar a semente permite reproduzir a randomização.
	RandomizationSeed int64
}
