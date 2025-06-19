# Especificação Técnica do Comando `aula`

## Visão Geral

O comando `aula` é responsável por gerenciar todas as informações relacionadas às aulas, como planejamento, registro de aulas dadas e conteúdo.

## Estrutura de Dados de Referência

A principal struct Go associada a este comando é `Lesson` (definida em `internal/models/academic.go`):

```go
// Lesson representa uma aula.
type Lesson struct {
	ID          string    // Identificador único da aula
	Subject     string    // Disciplina da aula (ex: "Matemática")
	Topic       string    // Tópico da aula (ex: "Equações de 2º Grau")
	Date        time.Time // Data e hora da aula
	ClassID     string    // Identificador da turma para a qual a aula foi dada
	Plan        string    // Plano de aula detalhado
	Observations string    // Observações ou anotações sobre a aula
}
```

## Subcomandos

### 1. `aula criar`

Registra uma nova aula que foi ou será ministrada.

*   **Uso:** `vickgenda aula criar --disciplina <nome_disciplina> --topico "<topico_aula>" --data <dd-mm-aaaa> [--hora <hh:mm>] --turma <id_turma> [--plano "<plano_detalhado>"] [--obs "<observacoes>"]`
*   **Alias:** `vickgenda aula add`

*   **Argumentos/Flags:**
    *   `--disciplina <nome_disciplina>` (Obrigatório): Nome da disciplina. Deve corresponder a uma disciplina previamente cadastrada.
    *   `--topico "<topico_aula>"` (Obrigatório): Tópico principal da aula.
    *   `--data <dd-mm-aaaa>` (Obrigatório): Data da aula.
    *   `--hora <hh:mm>` (Opcional): Hora da aula. Se não fornecido, pode assumir um horário padrão ou ser apenas data.
    *   `--turma <id_turma>` (Obrigatório): Identificador da turma. Deve corresponder a uma turma previamente cadastrada.
    *   `--plano "<plano_detalhado>"` (Opcional): Texto com o plano de aula. Pode ser um texto longo.
    *   `--obs "<observacoes>"` (Opcional): Observações adicionais sobre a aula.

*   **Input:** Valores fornecidos via flags.
*   **Output:**
    *   Sucesso: "Aula de <disciplina> sobre '<topico>' para a turma <id_turma> em <data> criada com sucesso. ID: <id_aula>"
    *   Erro: Mensagens claras sobre campos ausentes, inválidos ou se a disciplina/turma não existir.

*   **Validação:**
    *   Todos os campos obrigatórios devem ser fornecidos.
    *   `<nome_disciplina>` deve existir no sistema.
    *   `<id_turma>` deve existir no sistema.
    *   `<data>` deve ser uma data válida no formato `dd-mm-aaaa`.
    *   `<hora>` (se fornecido) deve ser uma hora válida no formato `hh:mm`.

### 2. `aula listar`

Lista as aulas registradas, com filtros opcionais.

*   **Uso:** `vickgenda aula listar [--disciplina <nome_disciplina>] [--turma <id_turma>] [--periodo <data_inicio>:<data_fim>] [--mes <mm-aaaa>] [--ano <aaaa>]`
*   **Alias:** `vickgenda aula ls`

*   **Argumentos/Flags:**
    *   `--disciplina <nome_disciplina>` (Opcional): Filtrar por disciplina.
    *   `--turma <id_turma>` (Opcional): Filtrar por turma.
    *   `--periodo <data_inicio>:<data_fim>` (Opcional): Filtrar por um período específico (ex: `01-03-2024:15-03-2024`).
    *   `--mes <mm-aaaa>` (Opcional): Filtrar por mês/ano (ex: `03-2024`).
    *   `--ano <aaaa>` (Opcional): Filtrar por ano.

*   **Input:** Valores fornecidos via flags.
*   **Output:**
    *   Sucesso: Uma tabela formatada com as aulas encontradas, mostrando ID, Data, Disciplina, Tópico, Turma.
        ```
        ID        Data        Disciplina    Tópico                Turma
        --------  ----------  ------------  --------------------  -----
        aula001   10-03-2024  Matemática    Equações              T301
        aula002   12-03-2024  Português     Concordância Verbal   T302
        ```
    *   Nenhum resultado: "Nenhuma aula encontrada com os filtros especificados."
    *   Erro: Mensagens sobre formato de data inválido nos filtros.

*   **Validação:**
    *   Datas em filtros de período ou mês devem ser válidas.

### 3. `aula ver <id_aula>`

Exibe detalhes de uma aula específica.

*   **Uso:** `vickgenda aula ver <id_aula>`

*   **Argumentos/Flags:**
    *   `<id_aula>` (Obrigatório): O ID da aula a ser visualizada.

*   **Input:** ID da aula.
*   **Output:**
    *   Sucesso: Detalhes completos da aula:
        ```
        ID: aula001
        Disciplina: Matemática
        Tópico: Equações de 2º Grau
        Data: 10-03-2024 09:00
        Turma: T301
        Plano de Aula:
        -------------
        1. Introdução ao conceito de equações quadráticas.
        2. Demonstração da fórmula de Bhaskara.
        3. Exemplos práticos.
        4. Exercícios.
        Observações:
        -----------
        Alunos participativos. João teve dificuldade inicial mas compreendeu após explicação individual.
        ```
    *   Erro: "Aula com ID '<id_aula>' não encontrada."

*   **Validação:**
    *   `<id_aula>` deve ser fornecido.

### 4. `aula editar-plano <id_aula>`

Permite editar o plano de aula e as observações de uma aula existente. Outros campos como data, disciplina, turma e tópico não são editáveis por este comando para manter a integridade do registro; para tal, seria necessário excluir e criar novamente ou um comando `aula editar <id_aula>` mais completo (escopo para futuras discussões).

*   **Uso:** `vickgenda aula editar-plano <id_aula> [--plano "<novo_plano>"] [--obs "<novas_observacoes>"]`
*   **Alias:** `vickgenda aula plan`

*   **Argumentos/Flags:**
    *   `<id_aula>` (Obrigatório): O ID da aula a ser editada.
    *   `--plano "<novo_plano>"` (Opcional): O novo texto para o plano de aula. Se não fornecido, o plano atual é mantido.
    *   `--obs "<novas_observacoes>"` (Opcional): O novo texto para as observações. Se não fornecido, as observações atuais são mantidas.
    *   Se nem `--plano` nem `--obs` forem fornecidos, o comando pode abrir um editor de texto interativo (como Nano ou Vim) pré-preenchido com o plano e observações atuais, ou simplesmente informar que nada foi alterado.

*   **Input:** ID da aula e os novos textos para plano e/ou observações.
*   **Output:**
    *   Sucesso: "Plano de aula/observações da aula <id_aula> atualizados com sucesso."
    *   Nada a fazer: "Nenhuma alteração fornecida para o plano ou observações da aula <id_aula>."
    *   Erro: "Aula com ID '<id_aula>' não encontrada."

*   **Validação:**
    *   `<id_aula>` deve ser fornecido.
    *   Pelo menos uma das opções `--plano` ou `--obs` deve ser fornecida, ou o modo interativo (se implementado) deve ser acionado.

### 5. `aula excluir <id_aula>`

Remove o registro de uma aula.

*   **Uso:** `vickgenda aula excluir <id_aula> [--confirmar]`
*   **Alias:** `vickgenda aula rm`, `vickgenda aula del`

*   **Argumentos/Flags:**
    *   `<id_aula>` (Obrigatório): O ID da aula a ser excluída.
    *   `--confirmar` (Opcional): Confirma a exclusão sem pedir confirmação interativa. Útil para scripts.

*   **Input:** ID da aula.
*   **Output:**
    *   Sucesso: "Aula <id_aula> excluída com sucesso."
    *   Confirmação: Se `--confirmar` não for usado, perguntar: "Tem certeza que deseja excluir a aula <id_aula> (<disciplina> - <topico>)? (s/N)".
    *   Cancelado: "Exclusão da aula <id_aula> cancelada."
    *   Erro: "Aula com ID '<id_aula>' não encontrada."

*   **Validação:**
    *   `<id_aula>` deve ser fornecido.

## Considerações Adicionais

*   **IDs:** Os IDs (`<id_aula>`, `<id_turma>`) podem ser UUIDs ou um formato serial mais simples (ex: `aula001`, `aula002`). A definir pelo Squad 1 (Core).
*   **Persistência:** Os dados das aulas serão armazenados no banco de dados SQLite local.
*   **Interação com outros módulos:**
    *   O comando `aula` pode interagir com um futuro módulo de `disciplina` e `turma` para validar existências.
    *   Pode ser usado pelo Squad 4 (Experiência Principal) para exibir informações no dashboard ou relatórios.
