# Requisitos de Dados para o Módulo `relatorio`

## Objetivo:
Fornecer insights sobre as atividades do professor, produtividade e desempenho dos alunos.

## Dados Necessários do Squad 2 (Produtividade - `tarefa`, `agenda`, `rotina`):

### `tarefa`:
*   Número de tarefas criadas (filtrável por intervalo de datas, prioridade, projeto/contexto).
*   Número de tarefas concluídas (filtrável por intervalo de datas, prioridade, projeto/contexto).
*   Número de tarefas pendentes (filtrável por intervalo de datas, prioridade, projeto/contexto).
*   Tempo médio para concluir tarefas (geral e por projeto/contexto).
*   Categorias de tarefas mais comuns.

### `agenda`:
*   Tempo gasto em reuniões/eventos (filtrável por tipo de evento, intervalo de datas).
*   Número de eventos agendados (filtrável por tipo de evento, intervalo de datas).

### `rotina`:
*   Frequência de execução de cada rotina definida.
*   Número de tarefas geradas por rotinas (identificar quais rotinas geram mais tarefas).

## Dados Necessários do Squad 3 (Gestão Académica - `aula`, `notas`):

### `aula`:
*   Tempo total dedicado por disciplina (ex: "Matemática: 20 horas no último mês").
*   Tempo total dedicado por turma (ex: "Turma 7B: 15 horas em aulas no último bimestre").
*   Taxa de conclusão do plano de aula (se o sistema permitir registrar conteúdo planejado vs. conteúdo efetivamente ministrado).

### `notas`:
*   Médias de notas por turma (geral e por disciplina, filtrável por bimestre/trimestre/semestre).
*   Médias de notas por disciplina (geral e por turma, filtrável por bimestre/trimestre/semestre).
*   Distribuição de notas (quantitativo de conceitos/notas, ex: Turma 7B - Bimestre 1: 10 Alunos com A, 12 com B, 5 com C).
*   Lista de alunos que necessitam de atenção (abaixo de um certo limiar de nota, configurável).
*   Progressão de desempenho ao longo do tempo (comparativo de médias de uma turma ou aluno entre bimestres/termos).

## Dados Necessários do Squad 5 (Conteúdo Pedagógico - `bancoq`, `prova`):

### `bancoq` (Banco de Questões):
*   Número de questões disponíveis por disciplina.
*   Número de questões disponíveis por tópico/assunto dentro de uma disciplina.
*   Número de questões disponíveis por nível de dificuldade.
*   Frequência de uso de cada questão em provas geradas.
*   Data de criação/última atualização das questões (para identificar questões antigas).

### `prova`:
*   Número de provas criadas (filtrável por disciplina, turma, período).
*   Médias de acertos/erros em provas geradas pelo sistema (geral, por disciplina, por questão).
*   Tempo médio gasto para criar uma prova (se o processo de criação for interativo e cronometrado).
*   Uso de questões do `bancoq` vs. questões inseridas manualmente nas provas.

## Estrutura Conceitual dos Comandos de Relatório:

*   `relatorio produtividade [semanal|mensal|bimestral|anual] [detalhado|sumario]`:
    *   Mostra conclusão de tarefas, alocação de tempo em eventos da agenda, atividades de rotina.
    *   Exemplo: `relatorio produtividade mensal`

*   `relatorio academico [turma <nome_turma>|disciplina <nome_disciplina>] [bimestre <numero_bimestre>|geral]`:
    *   Mostra desempenho dos alunos, distribuições de notas, alunos que precisam de atenção.
    *   Exemplo: `relatorio academico turma 7B bimestre 2`
    *   Exemplo: `relatorio academico disciplina Matematica geral`

*   `relatorio uso-conteudo [disciplina <nome_disciplina>] [mensal|bimestral|anual]`:
    *   Mostra estatísticas do banco de questões, frequência de criação de provas, uso de questões.
    *   Exemplo: `relatorio uso-conteudo disciplina Historia bimestral`
