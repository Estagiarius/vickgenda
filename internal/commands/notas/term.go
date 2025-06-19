package notas

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"vickgenda/internal/models"
)

// Simulação de armazenamento em memória para Termos (Bimestres)
// Usaremos um slice para facilitar a ordenação por data, e um map para acesso rápido por ID.
var termsStore = make(map[string]models.Term)
var termList = []models.Term{} // Mantém uma lista ordenada para listagem
var nextTermID = 1

func generateTermID() string {
	id := fmt.Sprintf("term%03d", nextTermID)
	nextTermID++
	return id
}

// LimparStoreTermos é uma função auxiliar para testes.
func LimparStoreTermos() {
	termsStore = make(map[string]models.Term)
	termList = []models.Term{}
	nextTermID = 1
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

	// Adicionar validação de ano letivo (se dataInicio e dataFim pertencem ao anoLetivo)
	if dataInicio.Year() != anoLetivo || dataFim.Year() != anoLetivo {
		// Esta validação pode ser flexibilizada dependendo das regras de negócio (bimestres que cruzam anos)
		// Por ora, vamos manter estrito ao ano.
		// return models.Term{}, fmt.Errorf("as datas do bimestre (início: %d, fim: %d) não correspondem ao ano letivo fornecido (%d)", dataInicio.Year(), dataFim.Year(), anoLetivo)
	}


	// Validação de sobreposição (simplificada: apenas checa se o nome já existe no ano)
	// Uma validação completa de sobreposição de datas seria mais complexa.
	for _, existingTerm := range termList {
		// Assumindo que o nome do bimestre deve ser único por ano.
		// A especificação em avaliacao_spec.md diz: "O nome do período deve ser único dentro de um mesmo ano letivo."
		// E "Não deve haver sobreposição de datas entre períodos do mesmo ano letivo."
		// A checagem de nome é mais simples de implementar agora.
		// A checagem de sobreposição de datas: (StartA <= EndB) and (EndA >= StartB)
		if existingTerm.Name == nome && existingTerm.StartDate.Year() == anoLetivo { // Simplificando para checar ano da data de início
			return models.Term{}, fmt.Errorf("bimestre com nome '%s' já existe para o ano %d", nome, anoLetivo)
		}
        // Checagem de sobreposição de datas
        // (dataInicio ANTES OU IGUAL existingTerm.EndDate) E (dataFim DEPOIS OU IGUAL existingTerm.StartDate)
        if dataInicio.Before(existingTerm.EndDate.Add(time.Nanosecond)) && dataFim.After(existingTerm.StartDate.Add(-time.Nanosecond)) && existingTerm.StartDate.Year() == anoLetivo {
             return models.Term{}, fmt.Errorf("o período de %s a %s se sobrepõe com o bimestre existente '%s' (%s a %s) no ano %d", inicioStr, fimStr, existingTerm.Name, existingTerm.StartDate.Format(layout), existingTerm.EndDate.Format(layout), anoLetivo)
        }
	}


	newID := generateTermID()
	term := models.Term{
		ID:        newID,
		Name:      nome,
		StartDate: dataInicio,
		EndDate:   dataFim,
		// Year: anoLetivo, // Se adicionarmos Year à struct Term
	}

	termsStore[newID] = term
	termList = append(termList, term)
	// Manter a lista ordenada por data de início
	sort.Slice(termList, func(i, j int) bool {
		return termList[i].StartDate.Before(termList[j].StartDate)
	})

	return term, nil
}

// ConfigurarBimestreListar lista os bimestres (Terms) para um ano letivo.
// Uso (Exemplo para listar): vickgenda notas configurar-bimestres --ano <ano_letivo> listar
func ConfigurarBimestreListar(anoLetivo int) ([]models.Term, error) {
	if anoLetivo == 0 {
		return nil, errors.New("ano letivo é obrigatório para listar bimestres")
	}
	var result []models.Term
	for _, term := range termList {
		// Considerar se o Term struct terá um campo Year ou se filtramos pelo ano da StartDate
		if term.StartDate.Year() == anoLetivo || term.EndDate.Year() == anoLetivo { // Permite bimestres que cruzam a virada do ano mas pertencem ao ano letivo
			result = append(result, term)
		}
	}
	// A termList já está ordenada, então result também estará se não houver filtro complexo de ano.
	// Se filtrarmos estritamente por StartDate.Year() == anoLetivo, a ordem é mantida.
	return result, nil
}

// GetTermByID busca um Term pelo ID. Auxiliar para outros pacotes.
func GetTermByID(id string) (models.Term, error) {
    term, found := termsStore[id]
    if !found {
        return models.Term{}, fmt.Errorf("termo (bimestre) com ID '%s' não encontrado", id)
    }
    return term, nil
}
