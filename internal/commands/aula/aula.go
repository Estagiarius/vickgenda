package aula

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"vickgenda-cli/internal/models"
	"vickgenda-cli/internal/store"
)

var ActiveAulaStore store.AulaStore

func InitializeAulaStore(db *sql.DB) {
	ActiveAulaStore = store.NewSQLiteAulaStore(db)
}

// CriarAula adiciona uma nova aula.
func CriarAula(disciplina, topico, dataStr, horaStr, turma, plano, obs string) (models.Lesson, error) {
	if disciplina == "" || topico == "" || dataStr == "" || turma == "" {
		return models.Lesson{}, errors.New("disciplina, tópico, data e turma são obrigatórios")
	}

	layout := "02-01-2006"
	aulaData, err := time.Parse(layout, dataStr)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("formato de data inválido: %s. Use dd-mm-aaaa", dataStr)
	}

	if horaStr != "" {
		horaLayout := "15:04"
		aulaHora, errHora := time.Parse(horaLayout, horaStr)
		if errHora != nil {
			return models.Lesson{}, fmt.Errorf("formato de hora inválido: %s. Use hh:mm", horaStr)
		}
		aulaData = time.Date(aulaData.Year(), aulaData.Month(), aulaData.Day(), aulaHora.Hour(), aulaHora.Minute(), 0, 0, time.Local)
	}

	lesson := models.Lesson{
		// ID é gerado pelo store
		Subject:      disciplina,
		Topic:        topico,
		Date:         aulaData,
		ClassID:      turma,
		Plan:         plano,
		Observations: obs,
	}

	if ActiveAulaStore == nil {
		return models.Lesson{}, errors.New("AulaStore não inicializado")
	}
	savedLesson, err := ActiveAulaStore.SaveLesson(lesson)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("erro ao criar aula: %w", err)
	}
	return savedLesson, nil
}

// ListarAulas retorna uma lista de aulas, com filtros opcionais.
func ListarAulas(disciplina, turma, periodo, mes, ano string) ([]models.Lesson, error) {
	if ActiveAulaStore == nil {
		return nil, errors.New("AulaStore não inicializado")
	}
	// A lógica de parsing de datas para periodo, mes, ano é feita no store
	lessons, err := ActiveAulaStore.ListLessons(disciplina, turma, periodo, mes, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar aulas: %w", err)
	}
	return lessons, nil
}

// VerAula retorna os detalhes de uma aula específica pelo ID.
func VerAula(id string) (models.Lesson, error) {
	if ActiveAulaStore == nil {
		return models.Lesson{}, errors.New("AulaStore não inicializado")
	}
	lesson, err := ActiveAulaStore.GetLessonByID(id)
	if err != nil {
		// O store já formata o erro para sql.ErrNoRows
		return models.Lesson{}, fmt.Errorf("erro ao buscar aula por ID '%s': %w", id, err)
	}
	return lesson, nil
}

// EditarPlanoAula atualiza o plano e/ou observações de uma aula existente.
func EditarPlanoAula(id, novoPlano, novasObservacoes string) (models.Lesson, error) {
	if id == "" {
		return models.Lesson{}, errors.New("ID da aula é obrigatório para edição")
	}
	// Conforme subtask: "if novoPlano and novasObservacoes are both empty ...
	// the command should return an error before calling the store."
	// O comportamento exato de "vazio" pode depender se as flags CLI
	// distinguem entre "não fornecido" e "fornecido como string vazia".
	// Assumindo que se chegam como "" aqui, significa que não foram fornecidos ou foram explicitamente setados para vazio.
	// A lógica original da store (in-memory) permitia setar para vazio.
	// A nova store UpdateLessonPlan também permite. A questão é se o *comando* deve impedir.
	// A descrição do subtask diz "if ... both empty ... return an error before calling the store"
	// No entanto, a lógica original do EditarPlanoAula in-memory tinha `intentToChange`
	// e retornava erro se AMBOS `intentToChange` fossem false. Se um deles fosse `true` (porque a string não era vazia),
	// ele prosseguia, mesmo que a outra string fosse vazia (o que significaria limpar o campo).
	// Para manter a possibilidade de limpar um campo passando string vazia,
	// o erro só deve ocorrer se NENHUMA alteração for de fato passada (e.g. usuário não passou flags).
	// Se o usuário passa `aula editar-plano <id> --plano ""` ele quer limpar o plano.
	// O teste "nenhuma_alteracao" no aula_test.go passa "" para ambos, esperando um erro.
	// Portanto, a condição de erro é se AMBOS são "" e o usuário não passou as flags.
	// Se o usuário passou as flags mas com valor "", isso significa limpar.
	// A CLI normalmente não passaria "" se a flag não fosse usada.
	// Esta função não sabe se as flags foram usadas. Assumimos que "" significa "não alterar este campo específico"
	// a menos que a intenção seja limpar.
	// A store.UpdateLessonPlan já faz GetLessonByID primeiro.
	// A store.UpdateLessonPlan atualiza os campos e depois faz GetLessonByID.
	// A checagem de "nenhuma alteração fornecida" deve ser feita aqui se a intenção é
	// que a store não seja chamada desnecessariamente.
	// O problema é que a store não sabe se "" significa "não alterar" ou "alterar para vazio".
	// O store.UpdateLessonPlan atualiza para o valor passado.
	// A lógica original era: if novoPlano != "" -> intentToChangePlano = true.
	// Erro se !intentToChangePlano && !intentToChangeObs.
	// Isso significa que se as strings são vazias, não há intenção de mudança.
	// Vou manter essa lógica: se as strings recebidas são vazias, não há intenção de mudança.
	if novoPlano == "" && novasObservacoes == "" {
		// Para ser mais preciso, precisaríamos saber se as flags foram de fato passadas como vazias
		// ou não foram passadas de todo. Assumindo que "" aqui significa "não foi passada flag para este campo".
		return models.Lesson{}, errors.New("nenhuma alteração fornecida para plano ou observações")
	}

	if ActiveAulaStore == nil {
		return models.Lesson{}, errors.New("AulaStore não inicializado")
	}

	// O método ActiveAulaStore.UpdateLessonPlan lida com a busca, atualização e retorno da aula atualizada.
	// Ele também lida com o caso de a aula não ser encontrada.
	updatedLesson, err := ActiveAulaStore.UpdateLessonPlan(id, novoPlano, novasObservacoes)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("erro ao editar plano da aula ID '%s': %w", id, err)
	}
	return updatedLesson, nil
}

// ExcluirAula remove uma aula do armazenamento.
func ExcluirAula(id string) error {
	if id == "" {
		return errors.New("ID da aula é obrigatório para exclusão")
	}
	if ActiveAulaStore == nil {
		return errors.New("AulaStore não inicializado")
	}
	err := ActiveAulaStore.DeleteLesson(id)
	if err != nil {
		// O store já pode formatar o erro para sql.ErrNoRows
		return fmt.Errorf("erro ao excluir aula ID '%s': %w", id, err)
	}
	return nil
}
