# Esquema JSON para Importação de Questões

Este documento define o esquema JSON para importar questões para o `bancoq` (banco de questões). O esquema é baseado na struct `models.Question`.

## Definição do Esquema

A importação espera um array JSON de objetos de questão. Cada objeto de questão deve estar em conformidade com a seguinte estrutura:

```json
{
  "id": "string (UUID recomendado, opcional, será gerado se ausente)",
  "subject": "string (obrigatório, ex: \"Matemática\")",
  "topic": "string (obrigatório, ex: \"Álgebra\")",
  "difficulty": "string (obrigatório, ex: \"easy\", \"medium\", \"hard\" - manter em inglês para consistência de código)",
  "question_text": "string (obrigatório, a questão em si)",
  "answer_options": [
    "string (para multiple_choice, true_false)"
  ],
  "correct_answers": [
    "string (obrigatório, uma ou mais respostas corretas)"
  ],
  "question_type": "string (obrigatório, ex: \"multiple_choice\", \"true_false\", \"essay\", \"short_answer\" - manter em inglês para consistência de código)",
  "source": "string (opcional, ex: \"Livro Didático A, Capítulo 5\")",
  "tags": [
    "string (opcional, para categorização)"
  ],
  "created_at": "string (datetime ISO 8601, opcional, padrão para hora da importação)",
  "last_used_at": "string (datetime ISO 8601, opcional)",
  "author": "string (opcional, quem criou/adicionou esta questão)"
}
```

### Descrições dos Campos:

*   **`id`**: (String, Opcional) Um identificador único para a questão. Se não fornecido, o sistema deve gerar um (ex: UUID v4).
*   **`subject`**: (String, Obrigatório) A matéria principal da questão (ex: "História", "Matemática"). O valor deve ser em português.
*   **`topic`**: (String, Obrigatório) Um tópico mais específico dentro da matéria (ex: "Revolução Francesa", "Equações de Primeiro Grau"). O valor deve ser em português.
*   **`difficulty`**: (String, Obrigatório) O nível de dificuldade. Valores sugeridos: `"easy"`, `"medium"`, `"hard"`. Estes valores são chaves internas e devem permanecer em inglês; a UI se encarregará da tradução para o usuário.
*   **`question_text`**: (String, Obrigatório) O texto completo da questão. O valor deve ser em português.
*   **`answer_options`**: (Array de Strings, Opcional) Para tipos de questão como `"multiple_choice"` ou `"true_false"`, este array contém as escolhas possíveis. Para `"essay"` ou `"short_answer"`, pode ser omitido ou ser um array vazio. Os valores devem ser em português.
*   **`correct_answers`**: (Array de Strings, Obrigatório) Um array contendo a(s) resposta(s) correta(s). Para múltipla escolha, seria o texto da(s) opção(ões) correta(s). Para verdadeiro/falso, seria `"Verdadeiro"` ou `"Falso"`. Para dissertativa/resposta curta, poderia ser uma resposta modelo ou pontos chave. Os valores devem ser em português.
*   **`question_type`**: (String, Obrigatório) O tipo de questão. Valores sugeridos: `"multiple_choice"`, `"true_false"`, `"essay"`, `"short_answer"`. Estes valores são chaves internas e devem permanecer em inglês; a UI se encarregará da tradução para o usuário.
*   **`source`**: (String, Opcional) A origem da questão (ex: "Livro Didático X, pg. 52", "Prova Anterior 2022"). O valor deve ser em português.
*   **`tags`**: (Array de Strings, Opcional) Tags para categorização e busca adicionais (ex: `["ENEM", "conceitual"]`). Os valores podem ser em português.
*   **`created_at`**: (String, Opcional) A data e hora em que a questão foi criada, em formato ISO 8601 (ex: `"2023-10-26T10:00:00Z"`). Padrão para a hora da importação se não fornecido.
*   **`last_used_at`**: (String, Opcional) A data e hora em que a questão foi usada pela última vez, em formato ISO 8601.
*   **`author`**: (String, Opcional) A pessoa que criou ou adicionou a questão.

## Exemplo de Conteúdo de Arquivo JSON para Importação

```json
[
  {
    "subject": "Matemática",
    "topic": "Geometria",
    "difficulty": "medium",
    "question_text": "Qual é a fórmula para a área de um círculo?",
    "answer_options": ["A = πr²", "A = 2πr", "A = πd", "A = r²"],
    "correct_answers": ["A = πr²"],
    "question_type": "multiple_choice",
    "tags": ["fórmula", "círculo"],
    "author": "Prof. Alan Turing"
  },
  {
    "subject": "História",
    "topic": "Segunda Guerra Mundial",
    "difficulty": "hard",
    "question_text": "Descreva os principais fatores que levaram ao início da Segunda Guerra Mundial.",
    "correct_answers": ["Expansionismo alemão, falha da Liga das Nações, Tratado de Versalhes, crise de 1929."],
    "question_type": "essay",
    "source": "Documentário XYZ",
    "created_at": "2022-05-10T14:30:00Z"
  },
  {
    "subject": "Ciências",
    "topic": "Biologia Celular",
    "difficulty": "easy",
    "question_text": "A mitocôndria é responsável pela respiração celular. (Verdadeiro/Falso)",
    "answer_options": ["Verdadeiro", "Falso"],
    "correct_answers": ["Verdadeiro"],
    "question_type": "true_false"
  }
]
```
