# Especificação Técnica do Comando `notas`

## Visão Geral

O comando `notas` é utilizado para gerenciar as notas dos alunos, configurar o sistema de avaliação (bimestres/períodos), lançar notas de avaliações e calcular médias.

## Estruturas de Dados de Referência

As principais structs Go associadas a este comando são `Grade`, `Term`, e `Student` (definidas em `internal/models/academic.go`):

```go
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
}

// Term representa um período de avaliação (ex: Bimestre).
type Term struct {
	ID        string    // Identificador único do período
	Name      string    // Nome do período (ex: "1º Bimestre")
	StartDate time.Time // Data de início do período
	EndDate   time.Time // Data de término do período
}

// Student representa um aluno.
type Student struct {
	ID   string // Identificador único do aluno (ex: matrícula)
	Name string // Nome completo do aluno
}
```

## Subcomandos

### 1. `notas configurar-bimestres`

(Este comando pode ser parte de uma configuração mais global, talvez `vickgenda configurar termos` ou similar, gerenciado pelo Squad 1, mas a especificação de como `notas` interage com ele é relevante aqui).

Permite definir ou visualizar os períodos de avaliação (bimestres/trimestres) para um ano letivo.

*   **Uso (Exemplo para definir):** `vickgenda notas configurar-bimestres --ano <ano_letivo> add --nome "1º Bimestre" --inicio <dd-mm-aaaa> --fim <dd-mm-aaaa>`
*   **Uso (Exemplo para listar):** `vickgenda notas configurar-bimestres --ano <ano_letivo> listar`

*   **Argumentos/Flags (para `add`):**
    *   `--ano <ano_letivo>` (Obrigatório): Ano letivo ao qual os bimestres se aplicam.
    *   `add` (Subcomando implícito ou explícito): Indica a ação de adicionar.
    *   `--nome "<nome_bimestre>"` (Obrigatório): Nome do bimestre (ex: "1º Bimestre", "2º Trimestre").
    *   `--inicio <dd-mm-aaaa>` (Obrigatório): Data de início do bimestre.
    *   `--fim <dd-mm-aaaa>` (Obrigatório): Data de fim do bimestre.

*   **Argumentos/Flags (para `listar`):**
    *   `--ano <ano_letivo>` (Obrigatório): Ano letivo para o qual listar os bimestres.
    *   `listar` (Subcomando implícito ou explícito): Indica a ação de listar.

*   **Input:** Dados do bimestre via flags.
*   **Output (`add`):**
    *   Sucesso: "Bimestre '<nome_bimestre>' (<inicio> - <fim>) adicionado para o ano <ano_letivo>. ID: <id_bimestre>"
    *   Erro: Mensagens sobre datas inválidas, sobreposição de datas com bimestres existentes, ou ano não especificado.
*   **Output (`listar`):**
    *   Sucesso: Tabela com ID, Nome, Início, Fim dos bimestres do ano.
        ```
        ID        Nome         Início      Fim
        --------  -----------  ----------  ----------
        term001   1º Bimestre  01-02-2024  15-04-2024
        term002   2º Bimestre  16-04-2024  30-06-2024
        ```
    *   Nenhum resultado: "Nenhum bimestre configurado para o ano <ano_letivo>."

*   **Validação:**
    *   Datas devem ser válidas e em ordem cronológica.
    *   Não deve haver sobreposição de datas entre bimestres do mesmo ano letivo.

### 2. `notas lancar`

Lança uma nova nota para um aluno em uma avaliação específica.

*   **Uso:** `vickgenda notas lancar --aluno <id_aluno> --bimestre <id_bimestre> --disciplina <nome_disciplina> --avaliacao "<desc_avaliacao>" --valor <nota> [--peso <peso>] [--data <dd-mm-aaaa>]`
*   **Alias:** `vickgenda notas add`

*   **Argumentos/Flags:**
    *   `--aluno <id_aluno>` (Obrigatório): ID do aluno.
    *   `--bimestre <id_bimestre>` (Obrigatório): ID do bimestre/período.
    *   `--disciplina <nome_disciplina>` (Obrigatório): Nome da disciplina.
    *   `--avaliacao "<desc_avaliacao>"` (Obrigatório): Descrição da avaliação (ex: "Prova 1", "Trabalho de História").
    *   `--valor <nota>` (Obrigatório): Valor da nota (numérico, ex: 8.5, 75).
    *   `--peso <peso>` (Opcional, Default: 1): Peso da nota para cálculo da média ponderada (ex: 0.4 para uma prova com 40% do peso). Se não informado, assume peso 1.
    *   `--data <dd-mm-aaaa>` (Opcional, Default: data atual): Data da avaliação ou lançamento da nota.

*   **Input:** Dados da nota via flags.
*   **Output:**
    *   Sucesso: "Nota <valor> para <desc_avaliacao> de <disciplina> lançada para o aluno <id_aluno> no bimestre <id_bimestre>. ID da nota: <id_nota>"
    *   Erro: Mensagens sobre campos inválidos, aluno/bimestre/disciplina não encontrados, valor da nota fora do intervalo permitido (ex: 0-10 ou 0-100, a ser definido).

*   **Validação:**
    *   `<id_aluno>`, `<id_bimestre>`, `<nome_disciplina>` devem existir.
    *   `<valor>` deve ser numérico e dentro da escala definida (ex: 0 a 10 ou 0 a 100).
    *   `<peso>` (se fornecido) deve ser numérico positivo.
    *   `<data>` (se fornecida) deve ser válida.

### 3. `notas ver`

Visualiza as notas de um aluno, podendo filtrar por bimestre ou disciplina.

*   **Uso:** `vickgenda notas ver --aluno <id_aluno> [--bimestre <id_bimestre>] [--disciplina <nome_disciplina>]`
*   **Alias:** `vickgenda notas show`, `vickgenda notas ls` (se diferenciado de listar alunos/bimestres)

*   **Argumentos/Flags:**
    *   `--aluno <id_aluno>` (Obrigatório): ID do aluno.
    *   `--bimestre <id_bimestre>` (Opcional): Filtrar por um bimestre específico.
    *   `--disciplina <nome_disciplina>` (Opcional): Filtrar por uma disciplina específica.

*   **Input:** ID do aluno e filtros opcionais.
*   **Output:**
    *   Sucesso: Tabela com as notas do aluno, mostrando Data, Disciplina, Avaliação, Valor, Peso.
        ```
        Aluno: [Nome do Aluno] (ID: <id_aluno>)
        Bimestre: [Nome do Bimestre] (ID: <id_bimestre>) (se filtrado)
        Disciplina: [Nome da Disciplina] (se filtrado)

        Data        Disciplina    Avaliação      Valor  Peso
        ----------  ------------  -------------  -----  ----
        10-03-2024  Matemática    Prova 1        8.5    0.4
        15-03-2024  Matemática    Trabalho       7.0    0.3
        20-04-2024  Português     Redação        9.0    0.5
        ...
        ```
    *   Nenhum resultado: "Nenhuma nota encontrada para o aluno <id_aluno> com os filtros especificados."
    *   Erro: "Aluno com ID '<id_aluno>' não encontrado."

*   **Validação:**
    *   `<id_aluno>` deve existir.
    *   `<id_bimestre>` (se fornecido) deve existir.
    *   `<nome_disciplina>` (se fornecida) deve existir.

### 4. `notas calcular-media`

Calcula e exibe a média de um aluno para uma disciplina em um determinado bimestre, ou a média geral do bimestre.

*   **Uso:** `vickgenda notas calcular-media --aluno <id_aluno> --bimestre <id_bimestre> [--disciplina <nome_disciplina>]`
*   **Alias:** `vickgenda notas media`

*   **Argumentos/Flags:**
    *   `--aluno <id_aluno>` (Obrigatório): ID do aluno.
    *   `--bimestre <id_bimestre>` (Obrigatório): ID do bimestre.
    *   `--disciplina <nome_disciplina>` (Opcional): Se fornecido, calcula a média da disciplina. Se omitido, pode calcular uma média geral do bimestre (se aplicável e definido como funcionalidade).

*   **Input:** IDs do aluno, bimestre e, opcionalmente, disciplina.
*   **Output:**
    *   Sucesso (Média da Disciplina):
        "Aluno: [Nome do Aluno] (<id_aluno>)
"
        "Bimestre: [Nome do Bimestre] (<id_bimestre>)
"
        "Disciplina: <nome_disciplina>
"
        "Notas Lançadas:
"
        "  - <desc_avaliacao1>: <valor1> (Peso: <peso1>)
"
        "  - <desc_avaliacao2>: <valor2> (Peso: <peso2>)
"
        "Soma dos Pesos: <soma_total_pesos>
"
        "Média Ponderada: <media_final>"
    *   Sucesso (Média Geral do Bimestre - se implementado): Similar, mas agregando de todas as disciplinas.
    *   Erro: "Não há notas suficientes ou pesos configurados para calcular a média de <disciplina> para o aluno <id_aluno> no bimestre <id_bimestre>."
    *   Erro: "Aluno <id_aluno>, bimestre <id_bimestre> ou disciplina <nome_disciplina> não encontrados."

*   **Cálculo da Média Ponderada:**
    *   Média = Σ (Nota * Peso) / Σ Pesos
    *   Se a soma dos pesos for zero, tratar para evitar divisão por zero (ex: média é zero ou erro).

*   **Validação:**
    *   Todos os IDs fornecidos devem existir.
    *   Deve haver notas lançadas com pesos para a combinação aluno/bimestre/disciplina.

### 5. `notas editar <id_nota>`

Permite editar uma nota lançada (valor, peso, descrição, data).

*   **Uso:** `vickgenda notas editar <id_nota> [--valor <novo_valor>] [--peso <novo_peso>] [--avaliacao "<nova_desc>"] [--data <nova_data>]`

*   **Argumentos/Flags:**
    *   `<id_nota>` (Obrigatório): ID da nota a ser editada.
    *   `--valor <novo_valor>` (Opcional): Novo valor da nota.
    *   `--peso <novo_peso>` (Opcional): Novo peso da nota.
    *   `--avaliacao "<nova_desc>"` (Opcional): Nova descrição da avaliação.
    *   `--data <nova_data>` (Opcional): Nova data da avaliação/lançamento.

*   **Input:** ID da nota e os campos a serem alterados.
*   **Output:**
    *   Sucesso: "Nota <id_nota> atualizada com sucesso."
    *   Erro: "Nota com ID '<id_nota>' não encontrada."
    *   Erro: Mensagens sobre valores inválidos para os campos.

*   **Validação:**
    *   `<id_nota>` deve existir.
    *   Novos valores devem seguir as mesmas regras de validação do `notas lancar`.

### 6. `notas excluir <id_nota>`

Remove uma nota lançada.

*   **Uso:** `vickgenda notas excluir <id_nota> [--confirmar]`
*   **Alias:** `vickgenda notas rm`, `vickgenda notas del`

*   **Argumentos/Flags:**
    *   `<id_nota>` (Obrigatório): ID da nota a ser excluída.
    *   `--confirmar` (Opcional): Confirma a exclusão sem prompt.

*   **Input:** ID da nota.
*   **Output:**
    *   Sucesso: "Nota <id_nota> excluída com sucesso."
    *   Confirmação: Se não usar `--confirmar`, perguntar: "Tem certeza que deseja excluir a nota <id_nota> (<desc_avaliacao> - <valor>)? (s/N)".
    *   Cancelado: "Exclusão da nota <id_nota> cancelada."
    *   Erro: "Nota com ID '<id_nota>' não encontrada."

*   **Validação:**
    *   `<id_nota>` deve ser fornecido.

## Considerações Adicionais

*   **Escala de Notas:** Definir claramente a escala de notas (ex: 0-10, 0-100). Isso pode ser uma configuração global.
*   **Arredondamento:** Definir regras de arredondamento para as médias.
*   **Consistência de IDs:** `id_aluno`, `id_bimestre`, `id_disciplina` (se esta se tornar uma entidade) devem ser consistentes e validados.
*   **Feedback ao Usuário:** As mensagens de erro e sucesso devem ser claras e informativas.
