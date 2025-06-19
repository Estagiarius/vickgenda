# Especificação Técnica: Comando Tarefa

O comando `tarefa` é usado para gerenciar as tarefas do usuário.

## Subcomandos

### 1. `tarefa criar`

*   **Propósito:** Adicionar uma nova tarefa à lista.
*   **Argumentos e Flags:**
    *   `--descricao "<texto>"` (obrigatório): O texto descritivo da tarefa.
    *   `--prazo "YYYY-MM-DD"` (opcional): Data de vencimento da tarefa. Se não fornecido, a tarefa não tem prazo.
    *   `--prioridade <numero>` (opcional): Nível de prioridade (ex: 1 para Alta, 2 para Média, 3 para Baixa). Padrão: 2 (Média).
    *   `--tags "<tag1>,<tag2>"` (opcional): Lista de tags separadas por vírgula.
*   **Comportamento Esperado:**
    *   Uma nova tarefa é criada com um ID único.
    *   A data de criação (`CreatedAt`) e atualização (`UpdatedAt`) são registradas automaticamente.
    *   O status inicial é "Pendente".
*   **Formato de Saída:**
    *   Sucesso: "Tarefa '<ID da tarefa>' criada com sucesso."
*   **Tratamento de Erros:**
    *   Descrição não fornecida: "Erro: A descrição da tarefa é obrigatória. Use --descricao "<texto>"."
    *   Formato de data inválido: "Erro: Formato de data inválido para --prazo. Use YYYY-MM-DD."
    *   Prioridade inválida: "Erro: Nível de prioridade inválido. Use um número (ex: 1, 2, 3)."

### 2. `tarefa listar`

*   **Propósito:** Listar todas as tarefas ou filtrar por critérios.
*   **Argumentos e Flags:**
    *   `--status <status>` (opcional): Filtrar por status (ex: "Pendente", "Em Andamento", "Concluída").
    *   `--prioridade <numero>` (opcional): Filtrar por prioridade.
    *   `--prazo-ate "YYYY-MM-DD"` (opcional): Listar tarefas com prazo até a data especificada.
    *   `--tag "<tag>"` (opcional): Filtrar por uma tag específica.
    *   `--ordenar-por <campo>` (opcional): Campo para ordenação (ex: "prazo", "prioridade", "descricao"). Padrão: "CreatedAt".
    *   `--ordem <asc|desc>` (opcional): Ordem de classificação ("asc" para ascendente, "desc" para descendente). Padrão: "asc".
*   **Comportamento Esperado:**
    *   Exibe uma lista de tarefas que correspondem aos filtros.
    *   Se nenhum filtro for fornecido, lista todas as tarefas.
*   **Formato de Saída:**
    *   Tabela com colunas: ID, Descrição, Prazo, Prioridade, Status, Tags.
    *   Se nenhuma tarefa for encontrada: "Nenhuma tarefa encontrada."
*   **Tratamento de Erros:**
    *   Critério de filtro inválido: "Erro: Critério de filtro '<criterio>' inválido."

### 3. `tarefa editar <ID da tarefa>`

*   **Propósito:** Modificar uma tarefa existente.
*   **Argumentos e Flags:**
    *   `<ID da tarefa>` (obrigatório): O ID da tarefa a ser editada.
    *   `--descricao "<novo_texto>"` (opcional): Novo texto descritivo.
    *   `--prazo "YYYY-MM-DD"` (opcional): Nova data de vencimento.
    *   `--prioridade <novo_numero>` (opcional): Novo nível de prioridade.
    *   `--status "<novo_status>"` (opcional): Novo status.
    *   `--tags "<tag1>,<tag2>"` (opcional): Nova lista de tags (substitui as existentes).
*   **Comportamento Esperado:**
    *   A tarefa especificada é atualizada com os novos valores.
    *   A data de atualização (`UpdatedAt`) é registrada automaticamente.
    *   Pelo menos uma flag de alteração deve ser fornecida.
*   **Formato de Saída:**
    *   Sucesso: "Tarefa '<ID da tarefa>' atualizada com sucesso."
*   **Tratamento de Erros:**
    *   Tarefa não encontrada: "Erro: Tarefa com ID '<ID da tarefa>' não encontrada."
    *   Nenhuma alteração especificada: "Erro: Nenhuma alteração especificada. Forneça pelo menos uma flag para modificar."
    *   Formato de data/prioridade inválido (similar ao `criar`).

### 4. `tarefa concluir <ID da tarefa>`

*   **Propósito:** Marcar uma tarefa como concluída.
*   **Argumentos e Flags:**
    *   `<ID da tarefa>` (obrigatório): O ID da tarefa a ser concluída.
*   **Comportamento Esperado:**
    *   O status da tarefa é alterado para "Concluída".
    *   A data de atualização (`UpdatedAt`) é registrada.
*   **Formato de Saída:**
    *   Sucesso: "Tarefa '<ID da tarefa>' marcada como concluída."
*   **Tratamento de Erros:**
    *   Tarefa não encontrada: "Erro: Tarefa com ID '<ID da tarefa>' não encontrada."
    *   Tarefa já concluída: "Info: Tarefa '<ID da tarefa>' já está concluída."

### 5. `tarefa remover <ID da tarefa>`

*   **Propósito:** Excluir uma tarefa.
*   **Argumentos e Flags:**
    *   `<ID da tarefa>` (obrigatório): O ID da tarefa a ser removida.
    *   `--force` (opcional): Remove sem pedir confirmação.
*   **Comportamento Esperado:**
    *   A tarefa especificada é permanentemente removida.
    *   Por padrão, pede confirmação antes de remover.
*   **Formato de Saída:**
    *   Confirmação (se `--force` não usado): "Tem certeza que deseja remover a tarefa '<ID da tarefa>'? (s/N)"
    *   Sucesso: "Tarefa '<ID da tarefa>' removida com sucesso."
    *   Cancelado: "Remoção cancelada."
*   **Tratamento de Erros:**
    *   Tarefa não encontrada: "Erro: Tarefa com ID '<ID da tarefa>' não encontrada."

```
