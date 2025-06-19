# Especificação Técnica: Comando `bancoq`

Este documento detalha o comando `bancoq` (Banco de Questões) da Vickgenda CLI, responsável por gerenciar a coleção de questões pedagógicas do professor.

## 1. Visão Geral do Comando

O comando `bancoq` permite aos usuários adicionar, listar, visualizar, editar, pesquisar, remover e importar questões para o banco de dados local.

**Nome do comando:** `vickgenda bancoq`

## 2. Estrutura de Dados de Referência

Este comando opera sobre a struct `models.Question` (definida em `internal/models/question.go`) e utiliza o esquema de importação JSON (definido em `docs/schemas/question_import_schema.md`).

## 3. Subcomandos

### 3.1. `bancoq add`

*   **Propósito:** Adicionar uma nova questão interativamente ou via flags.
*   **Uso:**
    *   `vickgenda bancoq add [flags]`
*   **Flags (para modo não interativo):**
    *   `--subject "Matemática"` (Obrigatório)
    *   `--topic "Álgebra"` (Obrigatório)
    *   `--difficulty "medium"` (Obrigatório: easy, medium, hard)
    *   `--type "multiple_choice"` (Obrigatório: multiple_choice, true_false, essay, short_answer)
    *   `--question "Qual a fórmula de Bhaskara?"` (Obrigatório)
    *   `--option "Opção A"` (Múltiplo, para `multiple_choice`, `true_false`)
    *   `--answer "Opção A"` (Múltiplo, respostas corretas)
    *   `--source "Livro X"` (Opcional)
    *   `--tag "ENEM"` (Múltiplo, opcional)
    *   `--author "Prof. Y"` (Opcional)
*   **Comportamento Interativo:**
    *   Se nenhuma flag obrigatória for fornecida, o comando entra em modo interativo, solicitando cada campo da questão passo a passo.
    *   As opções de múltipla escolha e respostas corretas são solicitadas até que o usuário indique que terminou.
*   **Saída:**
    *   Sucesso: "Questão adicionada com ID: [ID_DA_QUESTAO]"
    *   Erro: Mensagens claras sobre campos faltantes ou inválidos (em pt-BR).
*   **Interação com BD:** Cria um novo registro `Question` na base de dados.

### 3.2. `bancoq list`

*   **Propósito:** Listar as questões existentes com filtros opcionais.
*   **Uso:**
    *   `vickgenda bancoq list [flags]`
*   **Flags:**
    *   `--subject "Matemática"` (Opcional)
    *   `--topic "Álgebra"` (Opcional)
    *   `--difficulty "medium"` (Opcional)
    *   `--type "multiple_choice"` (Opcional)
    *   `--tag "ENEM"` (Opcional)
    *   `--author "Prof. Y"` (Opcional)
    *   `--limit 20` (Opcional, default 20)
    *   `--page 1` (Opcional, default 1, para paginação)
    *   `--sort-by "created_at"` (Opcional, default: created_at. Outras opções: subject, topic, difficulty, last_used_at)
    *   `--order "desc"` (Opcional, default: desc. Opções: asc, desc)
*   **Saída:**
    *   Tabela formatada com colunas: ID (curto), Assunto, Tópico, Tipo, Dificuldade, Início da Questão.
    *   Se nenhuma questão encontrada: "Nenhuma questão encontrada com os filtros aplicados."
*   **Interação com BD:** Lê registros `Question` da base de dados.

### 3.3. `bancoq view <id>`

*   **Propósito:** Visualizar todos os detalhes de uma questão específica.
*   **Uso:**
    *   `vickgenda bancoq view <ID_DA_QUESTAO>`
*   **Argumentos:**
    *   `<ID_DA_QUESTAO>` (Obrigatório): O ID completo da questão.
*   **Saída:**
    *   Exibição formatada de todos os campos da questão, incluindo texto completo, opções (se houver) e respostas.
    *   Se não encontrada: "Questão com ID [ID_DA_QUESTAO] não encontrada."
*   **Interação com BD:** Lê um registro `Question` específico.

### 3.4. `bancoq edit <id>`

*   **Propósito:** Editar uma questão existente interativamente.
*   **Uso:**
    *   `vickgenda bancoq edit <ID_DA_QUESTAO>`
*   **Argumentos:**
    *   `<ID_DA_QUESTAO>` (Obrigatório).
*   **Comportamento:**
    *   Carrega os dados da questão.
    *   Permite ao usuário modificar cada campo interativamente, mostrando o valor atual.
    *   O usuário pode pular campos que não deseja alterar.
*   **Saída:**
    *   Sucesso: "Questão [ID_DA_QUESTAO] atualizada com sucesso."
    *   Erro: "Questão com ID [ID_DA_QUESTAO] não encontrada."
*   **Interação com BD:** Atualiza um registro `Question` existente.

### 3.5. `bancoq delete <id>`

*   **Propósito:** Remover uma questão do banco de dados.
*   **Uso:**
    *   `vickgenda bancoq delete <ID_DA_QUESTAO> [--force]`
*   **Argumentos:**
    *   `<ID_DA_QUESTAO>` (Obrigatório).
*   **Flags:**
    *   `--force` ou `-f`: Pula a confirmação.
*   **Comportamento:**
    *   Solicita confirmação antes de deletar, a menos que `--force` seja usado.
*   **Saída:**
    *   Sucesso: "Questão [ID_DA_QUESTAO] removida com sucesso."
    *   Cancelado: "Remoção cancelada pelo usuário."
    *   Erro: "Questão com ID [ID_DA_QUESTAO] não encontrada."
*   **Interação com BD:** Remove um registro `Question`.

### 3.6. `bancoq import <filepath>`

*   **Propósito:** Importar questões de um arquivo JSON.
*   **Uso:**
    *   `vickgenda bancoq import <CAMINHO_DO_ARQUIVO_JSON>`
*   **Argumentos:**
    *   `<CAMINHO_DO_ARQUIVO_JSON>` (Obrigatório): Caminho para o arquivo JSON contendo um array de questões (ver `docs/schemas/question_import_schema.md`).
*   **Comportamento:**
    *   Processa o arquivo JSON.
    *   Para cada questão no arquivo:
        *   Valida os dados contra o schema.
        *   Se um ID for fornecido e já existir, pode pular, atualizar (requer flag --update-existing) ou falhar (default).
        *   Se nenhum ID for fornecido, gera um novo.
        *   Adiciona a questão ao banco de dados.
*   **Flags:**
    *   `--on-conflict "skip|update|fail"` (Default: `fail`): O que fazer se uma questão com o mesmo ID já existir. `update` necessitaria de lógica adicional para mesclar.
    *   `--dry-run`: Simula a importação sem gravar no banco, apenas reportando o que seria feito.
*   **Saída:**
    *   Progresso da importação (e.g., "Importando questão X de Y...").
    *   Resumo: "Importação concluída. X questões importadas com sucesso. Y questões falharam."
    *   Relatório de erros detalhado para questões que falharam na importação (e.g., "Erro na questão Z: Campo 'subject' obrigatório não fornecido.").
*   **Interação com BD:** Cria múltiplos registros `Question`.

### 3.7. `bancoq search "<query>"`

*   **Propósito:** Procurar questões por palavras-chave no texto da questão, assunto, tópico ou tags.
*   **Uso:**
    *   `vickgenda bancoq search "<TERMO_DE_BUSCA>" [flags]`
*   **Argumentos:**
    *   `<TERMO_DE_BUSCA>` (Obrigatório).
*   **Flags:**
    *   Mesmas flags de filtro e paginação de `bancoq list` (e.g., `--subject`, `--limit`, etc.).
    *   `--field "all|text|subject|topic|tags"` (Default: `text`): Em quais campos procurar. `all` busca em todos os campos textuais relevantes.
*   **Saída:**
    *   Similar a `bancoq list`, mostrando as questões que correspondem à busca.
*   **Interação com BD:** Lê registros `Question` usando consultas de pesquisa (e.g., LIKE ou Full-Text Search se o SQLite estiver configurado para isso).

## 4. Considerações Gerais

*   **IDs:** IDs de questões devem ser únicos (preferencialmente UUIDs). IDs curtos podem ser usados para exibição e entrada do usuário onde não houver ambiguidade, mas o sistema deve sempre resolver para o ID completo internamente.
*   **Validação:** Validação robusta de entradas e dados importados.
*   **Feedback ao Usuário:** Mensagens claras e úteis em Português do Brasil (pt-BR).
*   **Ajuda:** Cada subcomando deve ter uma tela de ajuda detalhada (`--help`).
