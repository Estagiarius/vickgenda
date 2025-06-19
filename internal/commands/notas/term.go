package notas

import (
	"errors"
	"fmt"
	// "sort" // No longer needed for in-memory sorting
	"time"

	"vickgenda/internal/models"
	"vickgenda/internal/store" // Import the new store package
)

var nextTermID = 1 // Keep for now for generating IDs; consider UUIDs later.

func generateTermID() string {
	id := fmt.Sprintf("term%03d", nextTermID)
	nextTermID++
	return id
}

// ConfigurarBimestreAdicionar adiciona um novo bimestre (Term).
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

	// Basic validation for year (can be enhanced)
	if dataInicio.Year() != anoLetivo {
		// This could be a warning or an error based on strictness for terms crossing year boundaries
		// For now, let's treat it as a point of attention, not a hard error, to align with previous code.
		// Consider logging: log.Printf("Aviso: O ano da data de início (%d) não corresponde ao ano letivo fornecido (%d) para o bimestre '%s'", dataInicio.Year(), anoLetivo, nome)
	}

	// Validação de sobreposição e nome único é agora primariamente tratada pela CONSTRAINT
	// UNIQUE(name, strftime('%Y', start_date)) no banco de dados.
	// A store.CreateTerm retornará um erro se a constraint for violada.
	// Se uma verificação prévia explícita ainda for desejada aqui (antes de tentar inserir),
	// seria preciso consultar o store.ListTermsByYear(dataInicio.Year()) e iterar.
	// Ex:
	// existingTerms, errList := store.ListTermsByYear(dataInicio.Year())
	// if errList != nil {
	//    return models.Term{}, fmt.Errorf("erro ao verificar bimestres existentes: %w", errList)
	// }
	// for _, et := range existingTerms {
	//    if et.Name == nome {
	//        return models.Term{}, fmt.Errorf("bimestre com nome '%s' já existe para o ano %d", nome, dataInicio.Year())
	//    }
	//    // Checagem de sobreposição de datas: (StartA <= EndB) and (EndA >= StartB)
	//    if dataInicio.Before(et.EndDate.Add(time.Nanosecond)) && dataFim.After(et.StartDate.Add(-time.Nanosecond)) {
	//         return models.Term{}, fmt.Errorf("o período de %s a %s se sobrepõe com o bimestre existente '%s' (%s a %s) no ano %d", inicioStr, fimStr, et.Name, et.StartDate.Format(layout), et.EndDate.Format(layout), dataInicio.Year())
	//    }
	// }
	// Por simplicidade e para confiar na DB constraint, essa checagem manual é omitida aqui,
	// assumindo que o erro do store.CreateTerm será suficiente.

	newID := generateTermID() // Continuamos usando o gerador incremental por enquanto.
	term := models.Term{
		ID:        newID,
		Name:      nome,
		StartDate: dataInicio,
		EndDate:   dataFim,
	}

	err = store.CreateTerm(term)
	if err != nil {
		// O erro pode ser por violação da constraint UNIQUE ou outro problema de DB.
		return models.Term{}, fmt.Errorf("erro ao salvar bimestre: %w", err)
	}

	return term, nil
}

// ConfigurarBimestreListar lista os bimestres (Terms) para um ano letivo.
func ConfigurarBimestreListar(anoLetivo int) ([]models.Term, error) {
	if anoLetivo == 0 {
		return nil, errors.New("ano letivo é obrigatório para listar bimestres")
	}
	terms, err := store.ListTermsByYear(anoLetivo)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar bimestres: %w", err)
	}
	// A ordenação já é feita em store.ListTermsByYear "ORDER BY start_date"
	return terms, nil
}

// GetTermByID busca um Term pelo ID usando o store.
// Esta função é importante se outras partes do pacote 'notas' ou outros pacotes
// precisarem buscar um Term através da lógica de comandos/notas e não diretamente do store.
func GetTermByID(id string) (models.Term, error) {
	if id == "" {
		return models.Term{}, errors.New("ID do termo não pode ser vazio")
	}
	term, err := store.GetTermByID(id)
	if err != nil {
		// O store.GetTermByID já formata o erro para "term with ID '%s' not found"
		// ou outros erros de acesso ao DB.
		return models.Term{}, err
	}
	return term, nil
}

// NOTA: A função LimparStoreTermos() foi REMOVIDA daqui.
// Os testes que dependiam dela devem agora:
// 1. Chamar store.InitTermsTable() para garantir que a tabela exista.
// 2. Chamar store.ClearTermsTableForTesting() antes de cada teste ou conjunto de testes
//    para limpar os dados da tabela 'terms'.
// Isso requer que o pacote de store e db estejam corretamente configurados para o ambiente de teste
// (ex: usando um banco de dados em memória ou um banco de teste dedicado).

// Outras funções do pacote 'notas' (como as em grade.go) que dependiam de
// termStore local ou termList local para validar TermID (ex: buscando um bimestre por ID)
// precisarão ser atualizadas para usar GetTermByID(id) desta package, que agora chama o store.
// Exemplo: Se grade.go tinha algo como `term, ok := termsStore[termID]`,
// deverá ser mudado para `term, err := GetTermByID(termID)`.
// É importante garantir que InitTermsTable seja chamado uma vez no início da aplicação/testes.
package notas

import (
	"errors"
	"fmt"
	// "sort" // No longer needed for in-memory sorting
	"time"

	"vickgenda/internal/models"
	"vickgenda/internal/store" // Import the new store package
)

var nextTermID = 1 // Keep for now for generating IDs; consider UUIDs later.

func generateTermID() string {
	id := fmt.Sprintf("term%03d", nextTermID)
	nextTermID++
	return id
}

// ConfigurarBimestreAdicionar adiciona um novo bimestre (Term).
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

	// Basic validation for year (can be enhanced)
	if dataInicio.Year() != anoLetivo {
		// This could be a warning or an error based on strictness for terms crossing year boundaries
		// For now, let's treat it as a point of attention, not a hard error, to align with previous code.
		// Consider logging: log.Printf("Aviso: O ano da data de início (%d) não corresponde ao ano letivo fornecido (%d) para o bimestre '%s'", dataInicio.Year(), anoLetivo, nome)
	}

	// Validação de sobreposição e nome único é agora primariamente tratada pela CONSTRAINT
	// UNIQUE(name, strftime('%Y', start_date)) no banco de dados.
	// A store.CreateTerm retornará um erro se a constraint for violada.
	// Se uma verificação prévia explícita ainda for desejada aqui (antes de tentar inserir),
	// seria preciso consultar o store.ListTermsByYear(dataInicio.Year()) e iterar.
	// Ex:
	// existingTerms, errList := store.ListTermsByYear(dataInicio.Year())
	// if errList != nil {
	//    return models.Term{}, fmt.Errorf("erro ao verificar bimestres existentes: %w", errList)
	// }
	// for _, et := range existingTerms {
	//    if et.Name == nome {
	//        return models.Term{}, fmt.Errorf("bimestre com nome '%s' já existe para o ano %d", nome, dataInicio.Year())
	//    }
	//    // Checagem de sobreposição de datas: (StartA <= EndB) and (EndA >= StartB)
	//    if dataInicio.Before(et.EndDate.Add(time.Nanosecond)) && dataFim.After(et.StartDate.Add(-time.Nanosecond)) {
	//         return models.Term{}, fmt.Errorf("o período de %s a %s se sobrepõe com o bimestre existente '%s' (%s a %s) no ano %d", inicioStr, fimStr, et.Name, et.StartDate.Format(layout), et.EndDate.Format(layout), dataInicio.Year())
	//    }
	// }
	// Por simplicidade e para confiar na DB constraint, essa checagem manual é omitida aqui,
	// assumindo que o erro do store.CreateTerm será suficiente.

	newID := generateTermID() // Continuamos usando o gerador incremental por enquanto.
	term := models.Term{
		ID:        newID,
		Name:      nome,
		StartDate: dataInicio,
		EndDate:   dataFim,
	}

	err = store.CreateTerm(term)
	if err != nil {
		// O erro pode ser por violação da constraint UNIQUE ou outro problema de DB.
		return models.Term{}, fmt.Errorf("erro ao salvar bimestre: %w", err)
	}

	return term, nil
}

// ConfigurarBimestreListar lista os bimestres (Terms) para um ano letivo.
func ConfigurarBimestreListar(anoLetivo int) ([]models.Term, error) {
	if anoLetivo == 0 {
		return nil, errors.New("ano letivo é obrigatório para listar bimestres")
	}
	terms, err := store.ListTermsByYear(anoLetivo)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar bimestres: %w", err)
	}
	// A ordenação já é feita em store.ListTermsByYear "ORDER BY start_date"
	return terms, nil
}

// GetTermByID busca um Term pelo ID usando o store.
// Esta função é importante se outras partes do pacote 'notas' ou outros pacotes
// precisarem buscar um Term através da lógica de comandos/notas e não diretamente do store.
func GetTermByID(id string) (models.Term, error) {
	if id == "" {
		return models.Term{}, errors.New("ID do termo não pode ser vazio")
	}
	term, err := store.GetTermByID(id)
	if err != nil {
		// O store.GetTermByID já formata o erro para "term with ID '%s' not found"
		// ou outros erros de acesso ao DB.
		return models.Term{}, err
	}
	return term, nil
}

// NOTA: A função LimparStoreTermos() foi REMOVIDA daqui.
// Os testes que dependiam dela devem agora:
// 1. Chamar store.InitTermsTable() para garantir que a tabela exista.
// 2. Chamar store.ClearTermsTableForTesting() antes de cada teste ou conjunto de testes
//    para limpar os dados da tabela 'terms'.
// Isso requer que o pacote de store e db estejam corretamente configurados para o ambiente de teste
// (ex: usando um banco de dados em memória ou um banco de teste dedicado).

// Outras funções do pacote 'notas' (como as em grade.go) que dependiam de
// termStore local ou termList local para validar TermID (ex: buscando um bimestre por ID)
// precisarão ser atualizadas para usar GetTermByID(id) desta package, que agora chama o store.
// Exemplo: Se grade.go tinha algo como `term, ok := termsStore[termID]`,
// deverá ser mudado para `term, err := GetTermByID(termID)`.
// É importante garantir que InitTermsTable seja chamado uma vez no início da aplicação/testes.
