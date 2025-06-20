package notas

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"vickgenda/internal/models"
	"vickgenda/internal/store"
)

var ActiveTermStore store.TermStore

func InitializeTermStore(db *sql.DB) {
	ActiveTermStore = store.NewSQLiteTermStore(db)
}

// ConfigurarBimestreAdicionar adiciona um novo bimestre (Term).
// Uso (Exemplo para definir): vickgenda notas configurar-bimestres --ano <ano_letivo> add --nome "1º Bimestre" --inicio <dd-mm-aaaa> --fim <dd-mm-aaaa>
func ConfigurarBimestreAdicionar(anoLetivo int, nome, inicioStr, fimStr string) (models.Term, error) {
	if nome == "" || inicioStr == "" || fimStr == "" || anoLetivo == 0 {
		return models.Term{}, errors.New("nome, data de início, data de fim e ano letivo são obrigatórios")
	}

	layout := "02-01-2006"
	dataInicio, err := time.Parse(layout, inicioStr)
	if err != nil {
		return models.Term{}, fmt.Errorf("formato de data de início inválido: %s. Use dd-mm-aaaa", inicioStr)
	}
	dataFim, err := time.Parse(layout, fimStr)
	if err != nil {
		return models.Term{}, fmt.Errorf("formato de data de fim inválido: %s. Use dd-mm-aaaa", fimStr)
	}

	if dataFim.Before(dataInicio) {
		return models.Term{}, errors.New("data de fim não pode ser anterior à data de início")
	}

	// Validação de ano letivo agora é tratada implicitamente pelo store ao usar StartDate.Year()
	// A validação de sobreposição de datas e unicidade de nome também é feita pelo store.

	term := models.Term{
		// ID é gerado pelo store
		Name:      nome,
		StartDate: dataInicio,
		EndDate:   dataFim,
	}

	savedTerm, err := ActiveTermStore.SaveTerm(term)
	if err != nil {
		return models.Term{}, fmt.Errorf("erro ao salvar bimestre: %w", err)
	}

	return savedTerm, nil
}

// ConfigurarBimestreListar lista os bimestres (Terms) para um ano letivo.
// Uso (Exemplo para listar): vickgenda notas configurar-bimestres --ano <ano_letivo> listar
func ConfigurarBimestreListar(anoLetivo int) ([]models.Term, error) {
	if anoLetivo == 0 {
		return nil, errors.New("ano letivo é obrigatório para listar bimestres")
	}
	if ActiveTermStore == nil {
		return nil, errors.New("TermStore não inicializado")
	}
	terms, err := ActiveTermStore.ListTermsByYear(anoLetivo)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar bimestres para o ano %d: %w", anoLetivo, err)
	}
	return terms, nil
}

// GetTermByID busca um Term pelo ID. Auxiliar para outros pacotes.
func GetTermByID(id string) (models.Term, error) {
	if ActiveTermStore == nil {
		return models.Term{}, errors.New("TermStore não inicializado")
	}
	term, err := ActiveTermStore.GetTermByID(id)
	if err != nil {
		// O store já retorna um erro formatado para sql.ErrNoRows
		return models.Term{}, fmt.Errorf("erro ao buscar bimestre por ID '%s': %w", id, err)
	}
	return term, nil
}
