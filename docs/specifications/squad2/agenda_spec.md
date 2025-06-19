# Especificação Técnica: Comando Agenda

O comando `agenda` é usado para gerenciar os eventos e compromissos do usuário.

## Subcomandos

### 1. `agenda adicionar-evento`

*   **Propósito:** Adicionar um novo evento à agenda.
*   **Argumentos e Flags:**
    *   `--titulo "<texto>"` (obrigatório): Título do evento.
    *   `--inicio "YYYY-MM-DD HH:MM"` (obrigatório): Data e hora de início do evento.
    *   `--fim "YYYY-MM-DD HH:MM"` (obrigatório): Data e hora de término do evento.
    *   `--descricao "<texto>"` (opcional): Descrição detalhada do evento.
    *   `--local "<texto>"` (opcional): Local do evento.
*   **Comportamento Esperado:**
    *   Um novo evento é criado com um ID único.
    *   A data de criação (`CreatedAt`) e atualização (`UpdatedAt`) são registradas automaticamente.
    *   Valida se a hora de término é posterior à hora de início.
*   **Formato de Saída:**
    *   Sucesso: "Evento '<ID do evento>' adicionado com sucesso."
*   **Tratamento de Erros:**
    *   Campos obrigatórios não fornecidos: "Erro: Os campos --titulo, --inicio e --fim são obrigatórios."
    *   Formato de data/hora inválido: "Erro: Formato de data/hora inválido. Use YYYY-MM-DD HH:MM."
    *   Hora de término anterior ou igual à de início: "Erro: A hora de término deve ser posterior à hora de início."

### 2. `agenda listar-eventos`

*   **Propósito:** Listar eventos futuros ou dentro de um período específico.
*   **Argumentos e Flags:**
    *   `--periodo <dia|semana|mes|proximos>` (opcional): Período para listar eventos (ex: "dia" para hoje, "semana" para os próximos 7 dias, "mes" para os próximos 30 dias, "proximos" para todos os futuros). Padrão: "proximos".
    *   `--data-inicio "YYYY-MM-DD"` (opcional): Data de início para um período customizado. Requer `--data-fim`.
    *   `--data-fim "YYYY-MM-DD"` (opcional): Data de fim para um período customizado. Requer `--data-inicio`.
    *   `--ordenar-por <campo>` (opcional): Campo para ordenação (ex: "inicio", "titulo"). Padrão: "inicio".
    *   `--ordem <asc|desc>` (opcional): Ordem de classificação. Padrão: "asc".
*   **Comportamento Esperado:**
    *   Exibe uma lista de eventos que correspondem aos filtros.
*   **Formato de Saída:**
    *   Tabela com colunas: ID, Título, Início, Fim, Local, Descrição.
    *   Se nenhum evento for encontrado: "Nenhum evento encontrado para o período especificado."
*   **Tratamento de Erros:**
    *   Período inválido: "Erro: Período '<periodo>' inválido."
    *   Datas de período customizado ausentes ou incompletas: "Erro: Para período customizado, forneça --data-inicio e --data-fim."

### 3. `agenda ver-dia [YYYY-MM-DD]`

*   **Propósito:** Mostrar todos os eventos de um dia específico.
*   **Argumentos e Flags:**
    *   `[YYYY-MM-DD]` (opcional): A data para visualização. Se não fornecida, usa a data atual (hoje).
*   **Comportamento Esperado:**
    *   Lista todos os eventos agendados para a data especificada.
*   **Formato de Saída:**
    *   Lista formatada dos eventos do dia, mostrando Título, Horário (Início - Fim), Local.
    *   Ex: "Eventos para YYYY-MM-DD:"
        *   "09:00 - 10:00: Reunião de Equipe (Sala A)"
        *   "14:00 - 15:30: Consulta Médica (Clínica Central)"
    *   Se nenhum evento: "Nenhum evento agendado para YYYY-MM-DD."
*   **Tratamento de Erros:**
    *   Formato de data inválido: "Erro: Formato de data inválido. Use YYYY-MM-DD."

### 4. `agenda editar-evento <ID do evento>`

*   **Propósito:** Modificar um evento existente.
*   **Argumentos e Flags:**
    *   `<ID do evento>` (obrigatório): O ID do evento a ser editado.
    *   `--titulo "<novo_texto>"` (opcional)
    *   `--inicio "YYYY-MM-DD HH:MM"` (opcional)
    *   `--fim "YYYY-MM-DD HH:MM"` (opcional)
    *   `--descricao "<novo_texto>"` (opcional)
    *   `--local "<novo_texto>"` (opcional)
*   **Comportamento Esperado:**
    *   O evento especificado é atualizado.
    *   A data de atualização (`UpdatedAt`) é registrada.
    *   Pelo menos uma flag de alteração deve ser fornecida.
*   **Formato de Saída:**
    *   Sucesso: "Evento '<ID do evento>' atualizado com sucesso."
*   **Tratamento de Erros:**
    *   Evento não encontrado: "Erro: Evento com ID '<ID do evento>' não encontrado."
    *   Nenhuma alteração especificada: "Erro: Nenhuma alteração especificada."
    *   Erros de validação de data/hora (similar ao `adicionar-evento`).

### 5. `agenda remover-evento <ID do evento>`

*   **Propósito:** Excluir um evento da agenda.
*   **Argumentos e Flags:**
    *   `<ID do evento>` (obrigatório): O ID do evento a ser removido.
    *   `--force` (opcional): Remove sem pedir confirmação.
*   **Comportamento Esperado:**
    *   O evento é permanentemente removido.
    *   Pede confirmação por padrão.
*   **Formato de Saída:**
    *   Confirmação: "Tem certeza que deseja remover o evento '<Título do Evento>' (ID: <ID do evento>)? (s/N)"
    *   Sucesso: "Evento '<ID do evento>' removido com sucesso."
    *   Cancelado: "Remoção cancelada."
*   **Tratamento de Erros:**
    *   Evento não encontrado: "Erro: Evento com ID '<ID do evento>' não encontrado."

```
