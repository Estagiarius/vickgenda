# Guia de API e Estruturas de Dados do Squad 2 para Integração

## 1. Visão Geral

Este documento descreve as estruturas de dados e funções públicas (API) fornecidas pelo Squad 2 (Módulos de Produtividade: Tarefa, Agenda, Rotina). O objetivo é auxiliar outros squads, especialmente o Squad 4 (Experiência Principal), na integração e consumo dessas funcionalidades.

Todos os comentários, descrições de campos e mensagens de erro destinadas ao usuário estão em Português do Brasil (pt-BR). Nomes de funções e structs no código Go estão em Inglês.

## 2. Estruturas de Dados Principais (`internal/models`)

As seguintes structs são definidas em `internal/models/productivity.go`:

### 2.1. `Task`

Representa uma tarefa a ser realizada.

```go
type Task struct {
	ID          string    // Identificador único da tarefa
	Description string    // Descrição da tarefa
	DueDate     time.Time // Data de vencimento da tarefa (pode ser zero se não aplicável)
	Priority    int       // Prioridade da tarefa (ex: 1-Alta, 2-Média, 3-Baixa)
	Status      string    // Status da tarefa (ex: "Pendente", "Em Andamento", "Concluída")
	Tags        []string  // Etiquetas ou categorias para a tarefa
	CreatedAt   time.Time // Data de criação da tarefa
	UpdatedAt   time.Time // Data da última atualização da tarefa
}
```

### 2.2. `Event`

Representa um evento na agenda.

```go
type Event struct {
	ID          string    // Identificador único do evento
	Title       string    // Título do evento
	Description string    // Descrição detalhada do evento
	StartTime   time.Time // Data e hora de início do evento
	EndTime     time.Time // Data e hora de término do evento
	Location    string    // Local do evento (opcional)
	CreatedAt   time.Time // Data de criação do evento
	UpdatedAt   time.Time // Data da última atualização do evento
}
```

### 2.3. `Routine`

Representa uma rotina que pode gerar tarefas recorrentes.

```go
type Routine struct {
	ID                string    // Identificador único da rotina
	Name              string    // Nome da rotina (ex: "Preparar aula de Segunda")
	Description       string    // Descrição da rotina (do modelo em si)
	Frequency         string    // Frequência da rotina (ex: "diaria", "semanal:seg,qua", "manual")
	TaskDescription   string    // Modelo para a descrição das tarefas geradas (pode conter placeholders como {data} e {nome_rotina})
	TaskPriority      int       // Prioridade padrão para as tarefas geradas
	TaskTags          []string  // Etiquetas padrão para as tarefas geradas
	NextRunTime       time.Time // Próxima vez que a rotina deve ser executada para gerar tarefas (zero se manual ou não definida)
	CreatedAt         time.Time // Data de criação da rotina
	UpdatedAt         time.Time // Data da última atualização da rotina
}
```

## 3. Funções Públicas (API)

### 3.1. Módulo `tarefa` (`internal/commands/tarefa`)

Fornece funcionalidades para gerenciamento de tarefas.

#### `CriarTarefa(description string, dueDateStr string, priority int, tagsStr string) (models.Task, error)`
*   **Propósito:** Adiciona uma nova tarefa.
*   **Parâmetros:**
    *   `description`: Descrição textual da tarefa (obrigatória).
    *   `dueDateStr`: Data de vencimento no formato "YYYY-MM-DD" (opcional).
    *   `priority`: Prioridade numérica (opcional, padrão 2-Média).
    *   `tagsStr`: String de tags separadas por vírgula (opcional, ex: "urgente,casa").
*   **Retorno:** A `models.Task` criada ou um erro.
*   **Uso (Squad 4):** Permitir ao usuário criar novas tarefas através da UI.

#### `ListarTarefas(statusFilter string, priorityFilter int, dueDateFilterStr string, tagFilter string, sortBy string, sortOrder string) ([]models.Task, error)`
*   **Propósito:** Lista tarefas com base em filtros e ordenação.
*   **Parâmetros (todos opcionais):**
    *   `statusFilter`: Filtrar por status (ex: "Pendente").
    *   `priorityFilter`: Filtrar por prioridade.
    *   `dueDateFilterStr`: Filtrar por tarefas com prazo até "YYYY-MM-DD".
    *   `tagFilter`: Filtrar por uma tag específica.
    *   `sortBy`: Campo para ordenação (ex: "prazo", "prioridade", "descricao", "CreatedAt"). Padrão: "CreatedAt".
    *   `sortOrder`: Ordem ("asc" ou "desc"). Padrão: "asc".
*   **Retorno:** Slice de `models.Task` ou um erro.
*   **Uso (Squad 4):** Exibir listas de tarefas na UI, com filtros e ordenação definidos pelo usuário.

#### `EditarTarefa(id string, novaDesc, novoPrazoStr string, novaPrioridade int, novoStatus string, novasTagsStr string) (models.Task, error)`
*   **Propósito:** Modifica uma tarefa existente.
*   **Parâmetros:**
    *   `id`: ID da tarefa a ser editada (obrigatório).
    *   Demais parâmetros são opcionais e representam os novos valores. Pelo menos um deve ser fornecido.
*   **Retorno:** A `models.Task` atualizada ou um erro.
*   **Uso (Squad 4):** Permitir edição de tarefas existentes.

#### `ConcluirTarefa(id string) (models.Task, error)`
*   **Propósito:** Marca uma tarefa como "Concluída".
*   **Parâmetros:** `id` da tarefa.
*   **Retorno:** A `models.Task` atualizada ou um erro.
*   **Uso (Squad 4):** Botão/Ação para concluir uma tarefa.

#### `RemoverTarefa(id string) error`
*   **Propósito:** Exclui uma tarefa.
*   **Parâmetros:** `id` da tarefa.
*   **Retorno:** `nil` em sucesso, ou um erro.
*   **Uso (Squad 4):** Ação para remover uma tarefa.

#### `GetTarefaByID(id string) (models.Task, error)`
*   **Propósito:** Busca uma tarefa específica pelo seu ID.
*   **Retorno:** A `models.Task` ou um erro se não encontrada.
*   **Uso (Squad 4):** Obter detalhes de uma tarefa específica para exibição.

#### `ContarTarefas(statusFilter string, priorityFilter int, tagFilter string) (int, error)`
*   **Propósito:** Retorna a contagem de tarefas com base nos filtros.
*   **Parâmetros:** Similares aos de `ListarTarefas` (exceto ordenação e prazo).
*   **Retorno:** Número de tarefas ou um erro.
*   **Uso (Squad 4):** Exibir contagens no dashboard (ex: "3 tarefas pendentes").

### 3.2. Módulo `agenda` (`internal/commands/agenda`)

Fornece funcionalidades para gerenciamento de eventos da agenda.

#### `AdicionarEvento(titulo string, inicioStr string, fimStr string, descricao string, local string) (models.Event, error)`
*   **Propósito:** Adiciona um novo evento.
*   **Parâmetros:**
    *   `titulo`, `inicioStr` ("YYYY-MM-DD HH:MM"), `fimStr` ("YYYY-MM-DD HH:MM") são obrigatórios.
    *   `descricao`, `local` são opcionais.
*   **Retorno:** O `models.Event` criado ou um erro.
*   **Uso (Squad 4):** Permitir criação de novos eventos na agenda.

#### `ListarEventos(periodo string, dataInicioStr string, dataFimStr string, sortBy string, sortOrder string) ([]models.Event, error)`
*   **Propósito:** Lista eventos com base em período ou intervalo de datas.
*   **Parâmetros:**
    *   `periodo`: "dia", "semana", "mes", "proximos", "custom" (opcional, padrão "proximos").
    *   `dataInicioStr`, `dataFimStr`: "YYYY-MM-DD" para período "custom".
    *   `sortBy`, `sortOrder`: Para ordenação (opcional, padrão "inicio" "asc").
*   **Retorno:** Slice de `models.Event` ou um erro.
*   **Uso (Squad 4):** Exibir eventos em visualizações de calendário/agenda.

#### `VerDia(diaStr string) ([]models.Event, error)`
*   **Propósito:** Lista todos os eventos de um dia específico.
*   **Parâmetros:** `diaStr` no formato "YYYY-MM-DD".
*   **Retorno:** Slice de `models.Event` ou um erro.
*   **Uso (Squad 4):** Exibir detalhes de um dia específico na agenda.

#### `EditarEvento(id string, novoTitulo, novoInicioStr, novoFimStr, novaDesc, novoLocal string) (models.Event, error)`
*   **Propósito:** Modifica um evento existente.
*   **Parâmetros:** `id` do evento (obrigatório). Demais são opcionais.
*   **Retorno:** O `models.Event` atualizado ou um erro.
*   **Uso (Squad 4):** Permitir edição de eventos.

#### `RemoverEvento(id string) error`
*   **Propósito:** Exclui um evento.
*   **Parâmetros:** `id` do evento.
*   **Retorno:** `nil` em sucesso, ou um erro.
*   **Uso (Squad 4):** Ação para remover um evento.

#### `GetEventoByID(id string) (models.Event, error)`
*   **Propósito:** Busca um evento pelo ID.
*   **Retorno:** O `models.Event` ou um erro.
*   **Uso (Squad 4):** Obter detalhes de um evento para exibição.

#### `ListarProximosXEventos(count int) ([]models.Event, error)`
*   **Propósito:** Retorna uma lista dos próximos `count` eventos futuros ou em andamento.
*   **Parâmetros:** `count` (número de eventos a retornar). Se `count <= 0`, retorna todos os futuros/atuais.
*   **Retorno:** Slice de `models.Event` ou um erro.
*   **Uso (Squad 4):** Exibir uma pequena lista de "próximos eventos" no dashboard.

### 3.3. Módulo `rotina` (`internal/commands/rotina`)

Fornece funcionalidades para gerenciar modelos de rotinas e gerar tarefas a partir deles.

#### `CriarModeloRotina(nome, frequencia, descTarefa string, prioridadeTarefa int, tagsTarefaStr string, proximaExecucaoStr string) (models.Routine, error)`
*   **Propósito:** Cria um novo modelo de rotina.
*   **Parâmetros:**
    *   `nome`, `frequencia`, `descTarefa` são obrigatórios.
    *   `proximaExecucaoStr`: "YYYY-MM-DD HH:MM" (opcional, se não for manual e não fornecida, assume `time.Now()`).
*   **Retorno:** O `models.Routine` criado ou um erro.
*   **Uso (Squad 4):** UI para criar e configurar modelos de rotina.

#### `ListarModelosRotina(sortBy string, sortOrder string) ([]models.Routine, error)`
*   **Propósito:** Lista todos os modelos de rotina.
*   **Parâmetros:** `sortBy`, `sortOrder` (opcionais).
*   **Retorno:** Slice de `models.Routine` ou um erro.
*   **Uso (Squad 4):** Exibir lista de modelos de rotina existentes.

#### `EditarModeloRotina(id, novoNome, novaFreq, novaDescTarefa string, novaPrioTarefa int, novasTagsTarefaStr, novaProxExecStr string) (models.Routine, error)`
*   **Propósito:** Modifica um modelo de rotina.
*   **Parâmetros:** `id` do modelo (obrigatório). Demais opcionais.
*   **Retorno:** O `models.Routine` atualizado ou um erro.
*   **Uso (Squad 4):** Permitir edição de modelos de rotina.

#### `RemoverModeloRotina(id string) error`
*   **Propósito:** Exclui um modelo de rotina.
*   **Parâmetros:** `id` do modelo.
*   **Retorno:** `nil` em sucesso, ou um erro.
*   **Uso (Squad 4):** Ação para remover um modelo de rotina.

#### `GetModeloRotinaByID(id string) (models.Routine, error)`
*   **Propósito:** Busca um modelo de rotina pelo ID.
*   **Retorno:** O `models.Routine` ou um erro.
*   **Uso (Squad 4):** Exibir detalhes de um modelo para edição ou visualização.

#### `GerarTarefasFromModelo(modeloID string, dataBaseStr string) ([]models.Task, error)`
*   **Propósito:** Gera tarefas a partir de um modelo de rotina específico.
*   **Parâmetros:**
    *   `modeloID`: ID do modelo de rotina.
    *   `dataBaseStr`: Data base ("YYYY-MM-DD") para placeholders como `{data}` (opcional, padrão `time.Now()`).
*   **Retorno:** Slice de `models.Task` (geralmente uma tarefa) criadas ou um erro.
*   **Uso (Squad 4):** Permitir que o usuário acione manualmente a geração de tarefas de uma rotina, ou para o sistema de agendamento interno.

---
*Este documento deve ser mantido atualizado conforme a API do Squad 2 evolui.*

## 4. Cenários de Uso Mockados para Integração com Squad 4

Esta seção descreve alguns cenários típicos de como o Squad 4 (Experiência Principal/UI) poderia interagir com as APIs do Squad 2.

### 4.1. Exibição de Informações no Dashboard

*   **Objetivo do Dashboard:** Mostrar um resumo rápido para o usuário.
    *   Número de tarefas pendentes.
    *   Número de tarefas pendentes com prazo para hoje.
    *   Lista dos próximos 3 eventos de hoje.

*   **Ações do Squad 4 e Chamadas à API do Squad 2:**

    1.  **Contar Tarefas Pendentes:**
        *   Squad 4 chama: `tarefasPendentesCount, err := tarefa.ContarTarefas(statusFilter: "Pendente", priorityFilter: 0, tagFilter: "")`
        *   UI exibe `tarefasPendentesCount`.

    2.  **Contar Tarefas Pendentes com Prazo para Hoje:**
        *   Squad 4 obtém a data de hoje como string: `hojeStr := time.Now().Format("2006-01-02")`
        *   Squad 4 chama: `tarefasHoje, err := tarefa.ListarTarefas(statusFilter: "Pendente", priorityFilter: 0, dueDateFilterStr: hojeStr, tagFilter: "", sortBy: "", sortOrder: "")`
        *   UI exibe `len(tarefasHoje)`.
        *   *Nota: `dueDateFilterStr` em `ListarTarefas` lista tarefas COM prazo ATÉ a data especificada. Para "exatamente hoje", Squad 4 pode precisar filtrar adicionalmente o resultado se a tarefa tiver um campo de data exata de vencimento e não apenas um "até". A semântica atual de `ListarTarefas` com `dueDateFilterStr` pode precisar de clarificação ou um helper mais específico se "exatamente hoje" for um requisito comum.* (Para a struct `Task`, `DueDate` é uma `time.Time`, então a filtragem precisa ser `DueDate >= startOfToday && DueDate < startOfTomorrow`). A função `ListarTarefas` atual com `dueDateFilterStr` pode não ser precisa para "exatamente hoje" e pode precisar de um novo filtro ou ajuste no Squad 4.

    3.  **Listar Próximos 3 Eventos de Hoje:**
        *   Squad 4 obtém a data de hoje: `hojeStr := time.Now().Format("2006-01-02")`
        *   Squad 4 chama: `eventosHoje, err := agenda.VerDia(diaStr: hojeStr)`
        *   Squad 4 pega os primeiros 3 eventos da lista `eventosHoje` (se houver). A função `VerDia` já retorna os eventos ordenados por `StartTime`.
        *   UI exibe os detalhes dos eventos (Título, Horário).

### 4.2. Usuário Gera Tarefas Manualmente de uma Rotina

*   **Objetivo:** Usuário seleciona um modelo de rotina e clica em "Gerar Tarefas Agora". A UI deve atualizar a lista de tarefas.

*   **Ações do Squad 4 e Chamadas à API do Squad 2:**

    1.  **Contexto:** A UI já exibiu uma lista de modelos de rotina (obtida via `modelos, err := rotina.ListarModelosRotina("", "")`).
    2.  Usuário clica no botão "Gerar Tarefas" para um `ModeloX` com ID `routine-xyz`.
    3.  Squad 4 obtém a data atual: `hojeStr := time.Now().Format("2006-01-02")`
    4.  Squad 4 chama: `novasTarefas, err := rotina.GerarTarefasFromModelo(modeloID: "routine-xyz", dataBaseStr: hojeStr)`
    5.  **Em caso de sucesso (`err == nil`):**
        *   `novasTarefas` contém um slice das tarefas recém-criadas (geralmente uma).
        *   Squad 4 pode adicionar essas `novasTarefas` diretamente ao seu modelo de visualização da lista de tarefas ou recarregar a lista inteira chamando `tarefa.ListarTarefas(...)`.
        *   UI exibe uma mensagem de sucesso e a lista de tarefas é atualizada.
    6.  **Em caso de erro:**
        *   UI exibe a mensagem de erro retornada.

### 4.3. Visualizar e Concluir uma Tarefa

*   **Objetivo:** Usuário clica em uma tarefa para ver detalhes e depois a marca como concluída.

*   **Ações do Squad 4 e Chamadas à API do Squad 2:**

    1.  **Contexto:** UI exibe uma lista de tarefas. Usuário clica na `TarefaY` com ID `task-abc`.
    2.  **Opcional (se detalhes completos não estiverem na lista):**
        *   Squad 4 chama: `detalheTarefa, err := tarefa.GetTarefaByID(id: "task-abc")`
        *   UI exibe os `detalheTarefa`.
    3.  Usuário clica no botão "Concluir Tarefa".
    4.  Squad 4 chama: `tarefaAtualizada, err := tarefa.ConcluirTarefa(id: "task-abc")`
    5.  **Em caso de sucesso (`err == nil`):**
        *   `tarefaAtualizada` contém a tarefa com o status "Concluída".
        *   UI atualiza a aparência da `TarefaY` na lista (ex: tachado, cor diferente).
        *   UI pode atualizar contadores de tarefas pendentes (ex: chamando `tarefa.ContarTarefas(statusFilter: "Pendente", ...)`).
    6.  **Em caso de erro (ex: tarefa já concluída, não encontrada):**
        *   UI exibe a mensagem de erro.

## 5. Considerações Adicionais para Integração

Esta seção aborda alguns aspectos transversais importantes para a integração eficaz com os módulos do Squad 2.

### 5.1. Tratamento de Erros

*   **Verificação de Erros:** Todas as funções públicas da API do Squad 2 que podem falhar retornam um valor do tipo `error` como último parâmetro. **É crucial que o Squad 4 (e qualquer consumidor da API) verifique este valor.** Se `err != nil`, a operação não foi bem-sucedida.
*   **Mensagens de Erro:** As mensagens de erro retornadas são geralmente em português e tentam ser descritivas o suficiente para serem exibidas ao usuário final (ex: "Tarefa com ID 'X' não encontrada", "Formato de data inválido. Use YYYY-MM-DD").
    *   O Squad 4 pode optar por exibir essas mensagens diretamente ou usar um sistema de notificação/alerta da UI.
*   **Tipos Comuns de Erro:**
    *   **Validação de Entrada:** Campos obrigatórios ausentes, formatos de dados incorretos (especialmente datas/horas), valores lógicos inconsistentes (ex: data de término de evento antes da data de início).
    *   **Não Encontrado:** Tentativa de operar sobre um item (tarefa, evento, modelo de rotina) que não existe (ex: editar uma tarefa com ID inválido).
    *   **Estado Inválido:** Tentar realizar uma operação que não é permitida no estado atual do objeto (ex: tentar concluir uma tarefa que já está concluída).

### 5.2. Formatação de Dados (Datas e Horas)

*   **Entrada de Dados (Input para API do Squad 2):**
    *   Quando as funções da API do Squad 2 esperam strings para datas ou data/horas, os formatos exatos são especificados na documentação da função (e nas GoDoc comments). Geralmente são:
        *   Datas: `"YYYY-MM-DD"` (ex: `"2024-07-28"`)
        *   Data/Horas: `"YYYY-MM-DD HH:MM"` (ex: `"2024-07-28 14:30"`)
    *   É responsabilidade do Squad 4 garantir que as strings de entrada estejam nesses formatos.

*   **Saída de Dados (Output da API do Squad 2):**
    *   Campos de data e hora nas structs (`models.Task.DueDate`, `models.Event.StartTime`, etc.) são do tipo `time.Time` do Go.
    *   Isso dá ao Squad 4 total flexibilidade para formatar essas datas e horas para exibição na UI da maneira que for mais apropriada para o usuário (ex: "28/07/2024", "14:30h", "domingo, 28 de julho de 2024"). Utilize o método `Format()` do objeto `time.Time` com o layout de formatação desejado.

### 5.3. Concorrência

*   As funções da API do Squad 2 são projetadas para serem seguras para chamadas concorrentes (thread-safe). Os armazenamentos de dados em memória utilizados internamente são protegidos por mutexes.
*   O Squad 4 não precisa implementar mecanismos de bloqueio externos ao chamar funções individuais da API do Squad 2.

### 5.4. Idempotência

*   As operações de criação (ex: `CriarTarefa`) não são idempotentes; chamá-las múltiplas vezes resultará em múltiplos objetos criados.
*   Operações de edição e remoção são geralmente idempotentes no sentido de que tentar aplicar a mesma edição várias vezes terá o mesmo efeito final, e tentar remover um item já removido resultará em um erro de "não encontrado" (que é um resultado consistente).
*   O Squad 4 deve considerar a lógica da UI para evitar, por exemplo, submissões duplas de formulários de criação, se esse não for o comportamento desejado.
