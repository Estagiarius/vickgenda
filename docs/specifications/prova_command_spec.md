# Especificação Técnica: Comando `prova`

Este documento detalha o comando `prova` da Vickgenda CLI, responsável por gerar e gerenciar provas (avaliações) a partir das questões armazenadas no `bancoq` (Banco de Questões).

## 1. Visão Geral do Comando

O comando `prova` permite aos usuários gerar novas provas selecionando questões do banco, listar provas existentes, visualizar detalhes de uma prova, e potencialmente exportá-las.

**Nome do comando:** `vickgenda prova`

## 2. Estrutura de Dados de Referência

*   Este comando utiliza as questões (`models.Question`) do `bancoq`.
*   Será necessário definir uma nova struct, por exemplo `models.Test` (ou `models.Prova`), para armazenar informações sobre a prova gerada. Esta struct conteria:
    *   `ID`: Identificador único da prova.
    *   `Title`: Título da prova (e.g., "Prova Mensal de Matemática - Turma A").
    *   `Subject`: Disciplina principal da prova.
    *   `CreatedAt`: Data de criação.
    *   `Instructions`: Instruções gerais para a prova.
    *   `QuestionIDs`: Uma lista ordenada dos IDs das questões incluídas na prova.
    *   `LayoutOptions`: Opções de formatação (e.g., número de colunas, cabeçalho).
    *   `RandomizationSeed`: Se a ordem das questões ou alternativas foi randomizada, guardar a semente.

## 3. Subcomandos

### 3.1. `prova generate`

*   **Propósito:** Gerar uma nova prova interativamente ou via flags, selecionando questões do `bancoq`.
*   **Uso:**
    *   `vickgenda prova generate [flags]`
*   **Flags:**
    *   `--title "Prova Bimestral de História"` (Obrigatório)
    *   `--subject "História"` (Obrigatório, para filtrar questões)
    *   `--topic "Revolução Francesa"` (Múltiplo, opcional, para filtrar questões)
    *   `--difficulty "medium"` (Múltiplo, opcional: easy, medium, hard, para filtrar questões)
    *   `--type "multiple_choice"` (Múltiplo, opcional, para filtrar tipos de questão)
    *   `--tag " ENEM"` (Múltiplo, opcional, para filtrar por tags)
    *   `--num-questions 10` (Opcional, número total de questões desejadas)
    *   `--num-easy 3` (Opcional, número específico de questões fáceis)
    *   `--num-medium 4` (Opcional, número específico de questões médias)
    *   `--num-hard 3` (Opcional, número específico de questões difíceis)
    *   `--allow-duplicates false` (Opcional, default: false. Permite usar a mesma questão mais de uma vez na prova)
    *   `--randomize-order true` (Opcional, default: false. Randomiza a ordem das questões na prova)
    *   `--output-file "prova_historia.txt"` (Opcional. Se não fornecido, exibe no console)
    *   `--output-format "txt"` (Opcional, default: txt. Futuramente: md, pdf)
    *   `--instructions "Leia atentamente cada questão."` (Opcional)
*   **Comportamento:**
    1.  Filtra questões do `bancoq` com base nos critérios fornecidos (subject, topic, difficulty, type, tags).
    2.  Seleciona o número de questões especificado. Se `num-questions` for usado junto com `num-easy/medium/hard`, o sistema tenta atender às especificações de dificuldade dentro do total. Se houver conflito, prioriza `num-questions`.
    3.  Se não houver questões suficientes no banco para atender à solicitação, informa o usuário.
    4.  Permite ao usuário revisar as questões selecionadas e, opcionalmente, substituí-las ou adicioná-las manualmente (modo interativo avançado).
    5.  Gera a prova no formato especificado.
    6.  Salva os metadados da prova (struct `models.Test`) na base de dados.
*   **Saída:**
    *   Sucesso: "Prova '[ID_DA_PROVA] - Título da Prova' gerada com sucesso."
    *   Se `--output-file` especificado: "Prova salva em [CAMINHO_DO_ARQUIVO]."
    *   Se não: Exibe a prova formatada no console.
    *   Erro: Mensagens claras (em pt-BR) se não for possível gerar a prova (e.g., "Não há questões suficientes no banco para os critérios especificados.").
*   **Interação com BD:** Lê registros `Question`, cria um novo registro `Test`.

### 3.2. `prova list`

*   **Propósito:** Listar as provas já geradas.
*   **Uso:**
    *   `vickgenda prova list [flags]`
*   **Flags:**
    *   `--subject "História"` (Opcional)
    *   `--limit 10` (Opcional, default 10)
    *   `--page 1` (Opcional, default 1)
    *   `--sort-by "created_at"` (Opcional, default: created_at. Outras opções: title, subject)
    *   `--order "desc"` (Opcional, default: desc. Opções: asc, desc)
*   **Saída:**
    *   Tabela formatada com colunas: ID da Prova, Título, Assunto, Data de Criação, Nº de Questões.
    *   Se nenhuma prova encontrada: "Nenhuma prova encontrada."
*   **Interação com BD:** Lê registros `Test`.

### 3.3. `prova view <id_prova>`

*   **Propósito:** Visualizar uma prova específica, incluindo todas as suas questões.
*   **Uso:**
    *   `vickgenda prova view <ID_DA_PROVA> [--show-answers]`
*   **Argumentos:**
    *   `<ID_DA_PROVA>` (Obrigatório): O ID da prova.
*   **Flags:**
    *   `--show-answers` ou `-a`: Exibe as respostas corretas junto com as questões.
    *   `--output-format "txt"` (Opcional, default: txt. Futuramente: md, pdf)
*   **Saída:**
    *   Exibição formatada da prova, incluindo cabeçalho, instruções e todas as questões.
    *   Se `--show-answers` for usado, as respostas são mostradas.
    *   Se não encontrada: "Prova com ID [ID_DA_PROVA] não encontrada."
*   **Interação com BD:** Lê um registro `Test` e os registros `Question` associados.

### 3.4. `prova delete <id_prova>`

*   **Propósito:** Remover o registro de uma prova gerada (não remove as questões do banco).
*   **Uso:**
    *   `vickgenda prova delete <ID_DA_PROVA> [--force]`
*   **Argumentos:**
    *   `<ID_DA_PROVA>` (Obrigatório).
*   **Flags:**
    *   `--force` ou `-f`: Pula a confirmação.
*   **Comportamento:**
    *   Solicita confirmação antes de deletar, a menos que `--force` seja usado.
*   **Saída:**
    *   Sucesso: "Prova [ID_DA_PROVA] removida com sucesso."
    *   Cancelado: "Remoção cancelada pelo usuário."
    *   Erro: "Prova com ID [ID_DA_PROVA] não encontrada."
*   **Interação com BD:** Remove um registro `Test`.

### 3.5. `prova export <id_prova> <filepath>`

*   **Propósito:** Exportar uma prova gerada para um arquivo em um formato específico. (Similar a `prova generate --output-file` mas para provas já existentes).
*   **Uso:**
    *   `vickgenda prova export <ID_DA_PROVA> <CAMINHO_DO_ARQUIVO> [flags]`
*   **Argumentos:**
    *   `<ID_DA_PROVA>` (Obrigatório).
    *   `<CAMINHO_DO_ARQUIVO>` (Obrigatório).
*   **Flags:**
    *   `--format "txt"` (Opcional, default: txt. Futuramente: md, pdf).
    *   `--show-answers` (Opcional, default: false).
*   **Saída:**
    *   Sucesso: "Prova [ID_DA_PROVA] exportada para [CAMINHO_DO_ARQUIVO]."
    *   Erro: "Não foi possível exportar a prova. Verifique o ID e o caminho do arquivo."
*   **Interação com BD:** Lê um registro `Test` e os `Question` associados.

## 4. Considerações de Implementação

*   **Seleção de Questões:** A lógica para selecionar questões deve ser robusta, lidando com casos onde não há questões suficientes que atendam aos critérios.
*   **Randomização:** Se a randomização for implementada (ordem das questões ou alternativas), garantir que seja possível reproduzir uma prova gerada (armazenando a semente de randomização).
*   **Formatos de Saída:** Começar com texto simples (`.txt`). Markdown (`.md`) seria um bom próximo passo. PDF é mais complexo e pode ser um objetivo futuro.
*   **Definição da Struct `Test`:** A struct `models.Test` precisa ser definida no pacote `internal/models` antes da implementação deste comando.
