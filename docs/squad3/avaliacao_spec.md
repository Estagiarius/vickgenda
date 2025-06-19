# Especificação Técnica: Sistema de Avaliação por Bimestres

## Visão Geral

Este documento descreve como o sistema de avaliação por períodos (ex: bimestres, trimestres) será estruturado e gerenciado dentro do Vickgenda. Ele se baseia nas funcionalidades dos comandos `notas` e nas estruturas de dados como `Term`, `Grade`, e `Student`. O objetivo é permitir que o professor configure seus períodos letivos, defina como as avaliações contribuem para a nota final de cada período e calcule as médias dos alunos.

## 1. Configuração de Períodos (Bimestres/Trimestres)

A configuração dos períodos letivos é fundamental e serve como base para o lançamento de notas e cálculo de médias.

*   **Entidade Principal:** `Term` (definida em `internal/models/academic.go`)
    ```go
    // Term representa um período de avaliação (ex: Bimestre).
    type Term struct {
        ID        string    // Identificador único do período
        Name      string    // Nome do período (ex: "1º Bimestre")
        StartDate time.Time // Data de início do período
        EndDate   time.Time // Data de término do período
        // Futuramente: Year int, IsActive bool
    }
    ```
*   **Gerenciamento:**
    *   Os períodos são definidos globalmente por ano letivo. Não há, inicialmente, configuração de períodos por turma, simplificando o modelo.
    *   O comando `notas configurar-bimestres` (ou um comando global como `vickgenda configurar termos`) será usado para:
        *   Adicionar novos períodos (`add`): Especificando nome (ex: "1º Bimestre", "Recuperação Trimestral"), data de início e data de fim. O sistema deve gerar um ID único para cada `Term`.
        *   Listar períodos (`listar`): Visualizar os períodos configurados para um ano letivo.
        *   Editar períodos (`editar` - futuro): Modificar nome, datas de um período existente.
        *   Excluir períodos (`excluir` - futuro): Remover um período (com devidas precauções se já houver notas associadas).
*   **Validações:**
    *   Datas de início e fim devem ser válidas e cronológicas.
    *   Não deve haver sobreposição de datas entre períodos do mesmo ano letivo.
    *   O nome do período deve ser único dentro de um mesmo ano letivo.

## 2. Avaliações e Lançamento de Notas

As avaliações são as atividades ou instrumentos que geram notas para os alunos (provas, trabalhos, participação, etc.).

*   **Entidade Principal:** `Grade` (definida em `internal/models/academic.go`)
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
    ```
*   **Criação e Associação:**
    *   Uma "Avaliação" não é uma entidade separada no banco de dados inicialmente. Ela é representada pela `Description` na struct `Grade`. Por exemplo, "Prova 1 - Unidade 1" é uma avaliação específica dentro de uma disciplina e bimestre.
    *   Ao lançar uma nota com o comando `notas lancar`, o professor implicitamente define uma avaliação através do campo `--avaliacao "<desc_avaliacao>"`.
    *   Cada `Grade` (nota) é obrigatoriamente associada a:
        *   Um `StudentID` (aluno).
        *   Um `TermID` (bimestre/período).
        *   Um `Subject` (disciplina).
*   **Pesos das Avaliações:**
    *   Cada `Grade` possui um campo `Weight`. Este campo é crucial para o cálculo da média ponderada.
    *   O professor define o peso no momento do lançamento da nota (`--peso <valor>` no comando `notas lancar`).
    *   Exemplo:
        *   Prova 1: Nota 8.0, Peso 0.4 (ou 4)
        *   Trabalho em Grupo: Nota 7.0, Peso 0.3 (ou 3)
        *   Participação: Nota 9.0, Peso 0.3 (ou 3)
    *   A soma dos pesos para todas as avaliações de uma disciplina dentro de um bimestre não precisa necessariamente somar 1 (ou 10, ou 100), pois o cálculo da média ponderada normalizará isso (ver seção 3).

## 3. Cálculo de Médias Bimestrais/Por Período

O sistema deve ser capaz de calcular a média final de um aluno em uma disciplina específica dentro de um bimestre/período.

*   **Comando Envolvido:** `notas calcular-media`
*   **Lógica de Cálculo (Média Ponderada):**
    Para um dado aluno, bimestre e disciplina:
    1.  Selecionar todas as `Grade` (notas) que correspondem a esses três critérios.
    2.  Para cada nota selecionada, multiplicar seu `Value` pelo seu `Weight`.
        *   `ValorPonderadoDaNota = Nota.Value * Nota.Weight`
    3.  Somar todos os `ValorPonderadoDaNota` obtidos.
        *   `SomaDosValoresPonderados = Σ (Nota.Value * Nota.Weight)`
    4.  Somar todos os `Weight` das notas selecionadas.
        *   `SomaDosPesos = Σ Nota.Weight`
    5.  A média final é calculada como:
        *   `MediaFinal = SomaDosValoresPonderados / SomaDosPesos`
*   **Tratamento de Casos Especiais:**
    *   **Soma dos Pesos é Zero:** Se `SomaDosPesos` for 0 (ex: nenhuma nota lançada, ou todas as notas com peso zero), a média não pode ser calculada por divisão por zero. O sistema deve:
        *   Informar que a média não pode ser calculada.
        *   OU, considerar a média como 0 (a ser definido qual comportamento é mais apropriado).
    *   **Nenhuma Nota Lançada:** Similar ao caso acima. Se não houver notas para os critérios, informar que não há dados para o cálculo.
*   **Escala de Notas e Arredondamento:**
    *   A escala das notas (`Value` na struct `Grade`, ex: 0-10, 0-100) deve ser consistente. Idealmente, configurável ou claramente documentada.
    *   As regras de arredondamento para a `MediaFinal` devem ser definidas (ex: duas casas decimais, arredondamento padrão).

## 4. Visualização e Relatórios

*   O comando `notas ver` permite visualizar as notas lançadas, que são a base para qualquer cálculo de média.
*   O comando `notas calcular-media` exibe a média calculada e, opcionalmente, as notas e pesos que contribuíram para ela.
*   Futuramente, o Squad 4 (Experiência Principal) poderá usar esses dados para gerar relatórios mais complexos de desempenho de alunos e turmas.

## Fluxo de Exemplo para o Professor

1.  **Início do Ano Letivo:**
    *   Professor usa `notas configurar-bimestres` para definir os 4 bimestres do ano, com suas datas de início e fim.
    *   Ex: `vickgenda notas configurar-bimestres --ano 2024 add --nome "1º Bimestre" --inicio 01-02-2024 --fim 15-04-2024` (repete para os demais).

2.  **Durante o 1º Bimestre (Disciplina: Matemática):**
    *   Professor aplica a "Prova Mensal". Ele decide que esta prova terá peso 4.
    *   Para o aluno João (ID: `aluno001`), que tirou 7.5:
        `vickgenda notas lancar --aluno aluno001 --bimestre <id_1ºbim> --disciplina Matematica --avaliacao "Prova Mensal" --valor 7.5 --peso 4`
    *   Professor passa um "Trabalho Prático". Ele decide que este trabalho terá peso 3.
    *   Para o aluno João, que tirou 8.0:
        `vickgenda notas lancar --aluno aluno001 --bimestre <id_1ºbim> --disciplina Matematica --avaliacao "Trabalho Prático" --valor 8.0 --peso 3`
    *   Professor avalia a "Participação". Ele decide que terá peso 3.
    *   Para o aluno João, que obteve 9.0:
        `vickgenda notas lancar --aluno aluno001 --bimestre <id_1ºbim> --disciplina Matematica --avaliacao "Participação" --valor 9.0 --peso 3`

3.  **Final do 1º Bimestre (Calcular Média de Matemática para João):**
    *   Professor executa: `vickgenda notas calcular-media --aluno aluno001 --bimestre <id_1ºbim> --disciplina Matematica`
    *   O sistema calcula:
        *   SomaDosValoresPonderados = (7.5 * 4) + (8.0 * 3) + (9.0 * 3) = 30 + 24 + 27 = 81
        *   SomaDosPesos = 4 + 3 + 3 = 10
        *   MediaFinal = 81 / 10 = 8.1
    *   O sistema exibe a média 8.1 para João em Matemática no 1º Bimestre.

## Considerações Futuras (Fora do Escopo Inicial da Fase 1/2)

*   **Tipos de Avaliação Configuráveis:** Permitir que o professor defina "tipos" de avaliação (Prova, Trabalho, Participação) com pesos padrão por disciplina.
*   **Recuperação:** Sistema para lançamento de notas de recuperação e como elas afetam a média final do bimestre ou ano.
*   **Fechamento de Notas:** Um processo formal para "fechar" as notas de um bimestre, impedindo alterações posteriores sem permissão especial.
*   **Média Anual:** Cálculo da média final anual baseada nas médias bimestrais.
*   **Critérios de Aprovação/Reprovação:** Configuração de regras para determinar se um aluno foi aprovado ou reprovado.
