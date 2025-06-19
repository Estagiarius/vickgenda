# Especificação Técnica: Comando Rotina

O comando `rotina` é usado para gerenciar modelos de rotinas que podem gerar tarefas em massa de forma recorrente ou sob demanda.

## Subcomandos

### 1. `rotina criar-modelo`

*   **Propósito:** Criar um novo modelo de rotina.
*   **Argumentos e Flags:**
    *   `--nome "<nome>"` (obrigatório): Nome descritivo para o modelo da rotina (ex: "Preparativos para aula de segunda").
    *   `--frequencia "<tipo>"` (obrigatório): Define a recorrência.
        *   Valores possíveis: `"diaria"`, `"semanal:<dias>"`, `"mensal:<dia_do_mes>"`, `"manual"`.
        *   Exemplos: `"semanal:seg,qua,sex"`, `"mensal:15"`. Se `"manual"`, as tarefas só são geradas com `rotina gerar-tarefas`.
    *   `--desc-tarefa "<modelo_descricao>"` (obrigatório): Modelo para a descrição das tarefas a serem geradas. Pode incluir placeholders como `{nome_rotina}` ou `{data}`.
    *   `--prioridade-tarefa <numero>` (opcional): Prioridade padrão para as tarefas geradas (1-Alta, 2-Média, 3-Baixa). Padrão: 2.
    *   `--tags-tarefa "<tag1>,<tag2>"` (opcional): Tags padrão para as tarefas geradas.
    *   `--proxima-execucao "YYYY-MM-DD HH:MM"` (opcional, necessário se frequência não for "manual"): Data e hora da primeira execução para gerar tarefas. Se não especificado para rotinas automáticas, pode ser calculado com base na frequência (ex: próximo dia útil, próxima segunda-feira).
*   **Comportamento Esperado:**
    *   Um novo modelo de rotina é criado com um ID único.
    *   `CreatedAt` e `UpdatedAt` são registrados.
    *   `NextRunTime` é calculado se aplicável e não fornecido.
*   **Formato de Saída:**
    *   Sucesso: "Modelo de rotina '<ID do modelo>' criado com sucesso."
*   **Tratamento de Erros:**
    *   Campos obrigatórios não fornecidos.
    *   Formato de frequência inválido: "Erro: Formato de frequência inválido. Exemplos: 'diaria', 'semanal:seg,qua', 'mensal:1', 'manual'."
    *   Formato de data/hora inválido para `--proxima-execucao`.

### 2. `rotina listar-modelos`

*   **Propósito:** Listar todos os modelos de rotina existentes.
*   **Argumentos e Flags:**
    *   `--ordenar-por <campo>` (opcional): Campo para ordenação (ex: "nome", "proxima_execucao"). Padrão: "nome".
*   **Comportamento Esperado:**
    *   Exibe uma lista de todos os modelos de rotina.
*   **Formato de Saída:**
    *   Tabela com colunas: ID, Nome, Frequência, Descrição Tarefa Padrão, Próxima Execução.
    *   Se nenhum modelo: "Nenhum modelo de rotina encontrado."
*   **Tratamento de Erros:** N/A específico, além de falhas gerais do sistema.

### 3. `rotina gerar-tarefas <ID do modelo>`

*   **Propósito:** Gerar manualmente tarefas a partir de um modelo de rotina específico. Útil para rotinas com frequência "manual" ou para adiantar uma execução.
*   **Argumentos e Flags:**
    *   `<ID do modelo>` (obrigatório): O ID do modelo de rotina.
    *   `--data-base "YYYY-MM-DD"` (opcional): Data base para geração das tarefas (ex: se a descrição da tarefa inclui `{data}`). Padrão: data atual.
*   **Comportamento Esperado:**
    *   Cria novas tarefas na lista de tarefas do usuário, baseadas nos campos `TaskDescription`, `TaskPriority`, `TaskTags` do modelo.
    *   Placeholders na `TaskDescription` (como `{data}`) são substituídos.
    *   Se a rotina tem uma frequência automática, o `NextRunTime` do modelo pode ser atualizado.
*   **Formato de Saída:**
    *   Sucesso: "Tarefas geradas com sucesso a partir do modelo '<ID do modelo>'." (Pode listar os IDs das tarefas criadas).
*   **Tratamento de Erros:**
    *   Modelo não encontrado: "Erro: Modelo de rotina com ID '<ID do modelo>' não encontrado."
    *   Falha na criação de alguma tarefa.

### 4. `rotina editar-modelo <ID do modelo>`

*   **Propósito:** Modificar um modelo de rotina existente.
*   **Argumentos e Flags:**
    *   `<ID do modelo>` (obrigatório): ID do modelo a ser editado.
    *   `--nome "<novo_nome>"` (opcional)
    *   `--frequencia "<nova_frequencia>"` (opcional)
    *   `--desc-tarefa "<novo_modelo>"` (opcional)
    *   `--prioridade-tarefa <nova_prioridade>` (opcional)
    *   `--tags-tarefa "<novas_tags>"` (opcional)
    *   `--proxima-execucao "YYYY-MM-DD HH:MM"` (opcional)
*   **Comportamento Esperado:**
    *   O modelo de rotina é atualizado. `UpdatedAt` é registrado.
    *   Pelo menos uma flag de alteração deve ser fornecida.
*   **Formato de Saída:**
    *   Sucesso: "Modelo de rotina '<ID do modelo>' atualizado com sucesso."
*   **Tratamento de Erros:**
    *   Modelo não encontrado.
    *   Nenhuma alteração especificada.
    *   Erros de validação similares ao `criar-modelo`.

### 5. `rotina remover-modelo <ID do modelo>`

*   **Propósito:** Excluir um modelo de rotina. Não exclui tarefas já geradas por ele.
*   **Argumentos e Flags:**
    *   `<ID do modelo>` (obrigatório): ID do modelo a ser removido.
    *   `--force` (opcional): Remove sem pedir confirmação.
*   **Comportamento Esperado:**
    *   O modelo de rotina é permanentemente removido.
*   **Formato de Saída:**
    *   Confirmação: "Tem certeza que deseja remover o modelo de rotina '<Nome do Modelo>' (ID: <ID do modelo>)? (s/N)"
    *   Sucesso: "Modelo de rotina '<ID do modelo>' removido com sucesso."
*   **Tratamento de Erros:**
    *   Modelo não encontrado.

### Lógica de Geração Automática de Tarefas (Background)

*   Para rotinas com frequência diferente de "manual", um processo em segundo plano (ou um comando `vickgenda processar-rotinas` a ser chamado periodicamente pelo sistema do usuário, como cron) verificará os modelos de rotina.
*   Se `NextRunTime` de uma rotina for no passado ou presente, as tarefas são geradas.
*   Após a geração, `NextRunTime` é recalculado com base na frequência.
    *   `diaria`: Próximo dia às HH:MM especificadas (ou padrão, ex: 08:00).
    *   `semanal:dias`: Próximo dia da semana correspondente às HH:MM.
    *   `mensal:dia_do_mes`: Próximo mês no dia especificado às HH:MM.
*   Este processo está fora do escopo direto dos subcomandos `rotina`, mas a estrutura do modelo deve suportá-lo.

```
