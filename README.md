# vickgenda
Uma agenda CLI.

## **Plano de Implementação: Vickgenda**

Versão do Documento: Final 1.0  
Data: 19 de junho de 2025

### **Lembrete Essencial para Todas as Equipas**

**Antes de iniciar cada nova fase de implementação, é obrigatório que todos os membros da equipa releiam este documento na sua totalidade.** Esta prática garante o alinhamento contínuo com os objetivos, diretrizes técnicas e dependências entre os squads, sendo fundamental para o sucesso do projeto.

### **1\. Visão Geral da Estratégia**

O projeto será implementado como uma **Aplicação de Linha de Comando (CLI) Interativa e autónoma** para desktops (Windows, Mac, Linux). A arquitetura será modular, mas contida dentro de um único executável, garantindo uma instalação e utilização simples. O foco é entregar uma ferramenta de alta performance e sem distrações, que resolve os problemas centrais do professor no seu ambiente de trabalho principal. Esta abordagem respeita a necessidade de foco do utilizador alvo, oferecendo uma experiência rápida, controlada pelo teclado e livre da sobrecarga visual de interfaces gráficas tradicionais. A filosofia é capacitar o utilizador através da eficiência, em vez de sobrecarregá-lo com opções e elementos visuais desnecessários.

### **2\. Arquitetura da Aplicação**

A aplicação será um único binário Go, autocontido, o que facilita enormemente a distribuição e a instalação – o utilizador simplesmente descarrega e executa um único ficheiro. Internamente, a aplicação utilizará uma base de dados local para armazenar todos os dados do professor, eliminando a necessidade de uma ligação à internet para o seu funcionamento principal. Esta decisão arquitetónica garante máxima privacidade, pois nenhum dado sensível dos alunos ou do planeamento do professor sai da sua máquina, e oferece uma velocidade de resposta instantânea, crucial para manter o fluxo de trabalho sem interrupções. Esta abordagem "local-first" significa que a aplicação é fiável por defeito, funcionando de forma consistente independentemente do estado da rede do utilizador.

### **3\. Estrutura da Equipe e Divisão de Trabalho (5 Squads)**

Com a arquitetura simplificada, os cinco squads podem focar-se em entregar funcionalidades de alto valor em paralelo. Cada squad terá uma missão clara, atuando como "dono" da sua parte do produto.

* **Squad 1: Core & CLI Engine (A Fundação):** Responsável pelo motor da CLI, estrutura da base de dados local, componentes textuais, melhorias de fluxo de trabalho e o empacotamento da aplicação. A sua missão é criar uma plataforma estável e ergonómica para que os outros squads possam construir as suas funcionalidades.  
* **Squad 2: Módulos de Produtividade:** Focado nos comandos do dia a dia do professor: agenda, tarefa e rotina. A sua missão é reduzir o esforço mental necessário para a organização pessoal e temporal.  
* **Squad 3: Gestão Académica:** Focado nos processos de ensino: aula, notas e o sistema de avaliação por bimestres. A sua missão é automatizar e simplificar a burocracia associada à gestão de turmas e avaliações.  
* **Squad 4: Experiência Principal, Foco e Insights:** Responsável pela "alma" da CLI: dashboard, foco, relembrar e relatorio. A sua missão é tornar a aplicação intuitiva, motivadora e fornecer ao professor uma visão clara sobre o seu próprio trabalho.  
* **Squad 5: Conteúdo Pedagógico:** Responsável pelos ativos de ensino do professor: bancoq (banco de questões) e prova. A sua missão é dar ao professor superpoderes para criar e reutilizar materiais pedagógicos de forma eficiente.

### **4\. Diretrizes Técnicas e de Implementação**

* **1\. Padrões de Idioma:**  
  * **UI e Mensagens ao Utilizador:** Todo o texto visível ao utilizador final deve ser em **Português do Brasil (pt-BR)**.  
  * **Comentários de Código e Documentação:** Devem ser escritos em **Português do Brasil (pt-BR)**. Ex: // Esta função calcula a média ponderada do aluno..  
  * **Código-Fonte (Nomenclatura):** Nomes de variáveis, funções, pacotes, etc., devem ser em **Inglês**. Ex: func calculateWeightedAverage(studentID string).  
* **2\. Gestão de Código-Fonte (Git):**  
  * A branch main deve ser sempre estável e refletir o estado de uma potencial versão de lançamento.  
  * O trabalho é feito em feature-branches e integrado via Pull Requests (PRs) revistos. Um PR só pode ser aprovado se os testes automatizados passarem e se tiver a aprovação de pelo menos um outro colega.  
* **3\. Base de Dados:**  
  * **Tecnologia:** **SQLite**. O ficheiro da base de dados será armazenado localmente num diretório padrão do sistema operativo, garantindo que não polui o diretório pessoal do utilizador. As funções os.UserConfigDir() de Go serão usadas para encontrar os caminhos corretos de forma multiplataforma.  
* **4\. Gestão de Configuração:**  
  * **Formato:** Ficheiro **TOML** para configurações do utilizador, armazenado localmente no diretório de configuração padrão do sistema operativo.  
* **5\. Contratos de Dados (APIs Internas):**  
  * As structs Go partilhadas (ex: Task, Lesson, Question) residirão num pacote comum (/internal/models). Isto é crucial para desacoplar os módulos; por exemplo, o Squad 2 pode trabalhar na lógica das tarefas sem precisar de saber os detalhes da implementação da base de dados, apenas que interage com a struct Task.  
* **6\. Tratamento de Erros e Logging:**  
  * Erros previsíveis são mostrados ao utilizador de forma clara. Ex: Erro: A disciplina "Ciências" não foi encontrada. Use 'disciplina listar' para ver as disciplinas disponíveis..  
  * Erros inesperados são guardados num ficheiro de log local com detalhes técnicos (stack trace). Ex: FATAL: Falha ao escrever na base de dados: o disco está cheio. Stack: ....

### **5\. Ordem de Implementação, Fases e Instruções por Squad**

Esta secção detalha a linha temporal do projeto e as tarefas específicas de cada squad em cada fase.

#### **FASE 1: Fundação e Planeamento**

* **Foco Geral:** Estabelecer a arquitetura e as bases técnicas.  
* **Atividade Principal:** **Squad 1** implementa o núcleo da aplicação. **Squads 2, 3, 4 e 5** realizam o design técnico e a especificação dos seus módulos.  
* **Milestone de Conclusão:** **"Plataforma Pronta para Desenvolvimento"**.

**Instruções por Squad \- Fase 1:**

* **Squad 1 (Core):**  
  * Finalizar a escolha da stack tecnológica (bibliotecas Go como Cobra e Bubble Tea).  
  * Criar a estrutura de diretórios do projeto e o parser de comandos inicial.  
  * Implementar os comandos de autenticação (login, logout, registar).  
  * Desenvolver o wizard de configuração inicial (setup).  
  * Criar os primeiros componentes do Kit Textual (ex: uma função para renderizar tabelas de forma consistente).  
* **Squad 2 (Produtividade):**  
  * Definir as structs Go para Task, Event e Routine.  
  * Escrever a especificação técnica completa para os comandos tarefa, agenda e rotina.  
* **Squad 3 (Gestão Académica):**  
  * Definir as structs Go para Lesson, Grade, Term (Bimestre), Student, etc.  
  * Escrever a especificação técnica completa para os comandos aula, notas e o sistema de avaliação.  
* **Squad 4 (Experiência Principal):**  
  * Criar protótipos em texto (mockups) para o dashboard, relembrar e os ecrãs do modo foco.  
  * Definir as métricas e os dados necessários dos outros módulos para construir os relatórios.  
* **Squad 5 (Conteúdo Pedagógico):**  
  * Definir a struct Go para Question e o esquema do ficheiro JSON de importação.  
  * Escrever a especificação técnica completa para os comandos bancoq e prova.

#### **FASE 2: Desenvolvimento dos Módulos Principais**

* **Foco Geral:** Implementar a lógica de negócio de todas as funcionalidades centrais.  
* **Atividade Principal:** **Squads 2, 3 e 5** trabalham em paralelo no desenvolvimento dos seus comandos. **Squad 4** constrói a UI com dados fictícios. **Squad 1** fornece suporte e implementa as melhorias de workflow.  
* **Milestone de Conclusão:** **"Funcionalidades Core Implementadas"**.

**Instruções por Squad \- Fase 2:**

* **Squad 1 (Core):**  
  * Implementar os sistemas de IDs Contextuais, Autocompletar para o shell e a Barra de Status persistente.  
  * Dar suporte contínuo aos outros squads, refinando o Kit Textual com novos componentes conforme a necessidade.  
* **Squad 2 (Produtividade):**  
  * Desenvolver a lógica completa de criação, listagem, edição e conclusão para os comandos tarefa e agenda.  
  * Implementar o sistema de rotina para gerar tarefas em massa a partir de modelos.  
* **Squad 3 (Gestão Académica):**  
  * Desenvolver a lógica para gestão de aulas, incluindo o comando aula editar-plano.  
  * Implementar o sistema de notas, incluindo a configuração de bimestres, pesos e o cálculo da média ponderada.  
* **Squad 4 (Experiência Principal):**  
  * Desenvolver a interface de utilizador para os comandos dashboard, relembrar e foco, utilizando dados "mockados".  
  * Iniciar a prototipagem do módulo relatorio.  
* **Squad 5 (Conteúdo Pedagógico):**  
  * Desenvolver a lógica completa do bancoq, incluindo a importação de questões via JSON.  
  * Implementar o gerador de prova, com todos os seus filtros por disciplina, tópico e dificuldade.

#### **FASE 3: Integração e Testes de Fluxo de Trabalho**

* **Foco Geral:** Conectar todos os módulos e garantir que a aplicação funciona como um todo.  
* **Atividade Principal:** **Squad 4** lidera a integração, conectando a sua UI com os dados reais dos outros módulos. **Todos os squads** participam nos testes de ponta a ponta e na resolução de bugs.  
* **Milestone de Conclusão:** **"Versão Alpha Interna Pronta"**.

**Instruções por Squad \- Fase 3:**

* **Squad 1 (Core):**  
  * Liderar a otimização de performance, analisando gargalos.  
  * Gerar os binários da aplicação completa para testes internos multiplataforma (Windows, Mac, Linux).  
* **Squads 2, 3, 5 (Módulos Funcionais):**  
  * Participar ativamente nos testes de integração, assegurando que os vossos dados são corretamente consumidos e apresentados pelo Squad 4 e que os fluxos de trabalho que envolvem múltiplos módulos funcionam como esperado.  
* **Squad 4 (Experiência Principal):**  
  * Substituir todos os dados fictícios pelos dados reais, integrando com as APIs internas fornecidas pelos Squads 2, 3 e 5\. Esta é a vossa tarefa mais crítica nesta fase.

#### **FASE 4: Polimento, Documentação e Lançamento Beta**

* **Foco Geral:** Preparar a aplicação para o utilizador final, com foco na estabilidade, usabilidade e documentação.  
* **Atividade Principal:** **Todos os squads** focam-se em refinar a experiência do utilizador e completar a documentação. **Squad 1** finaliza o empacotamento para distribuição.  
* **Milestone de Conclusão:** **"Versão Beta Pronta para Lançamento"**.

**Instruções por Squad \- Fase 4:**

* **Squad 1 (Core):**  
  * Finalizar o empacotamento, criando instaladores ou instruções claras.  
  * Concluir a implementação das funcionalidades de personalização (temas e aliases).  
* **Squads 2, 3, 5 (Módulos Funcionais):**  
  * Rever e melhorar todas as mensagens de utilizador e a formatação textual dos vossos comandos.  
  * Garantir que a documentação de ajuda (--help) é clara, útil e completa para todos os vossos comandos e subcomandos.  
* **Squad 4 (Experiência Principal):**  
  * Realizar os ajustes finais no dashboard e nos relatórios com base no feedback dos testes internos.  
  * Garantir que a experiência de utilização geral é coesa, polida e intuitiva.

### **6\. Qualidade e Boas Práticas**

* **Revisão de Código (Code Review), Testes Automatizados, Documentação Contínua e Definição de "Pronto" (DoD)** são práticas obrigatórias em todo o projeto.

### **7\. Visão de Longo Prazo (Fora do Escopo Inicial)**

* **Sincronização na Nuvem:** Introduzir um backend opcional para sincronizar a base de dados SQLite entre diferentes instalações da CLI.  
* **Aplicações Móveis:** Criar interfaces de utilizador nativas para iOS e Android que consumam a API do serviço de sincronização.  
* **Colaboração:** Funcionalidades para partilhar planos de aula ou bancos de questões com outros professores.  
* **Sistema de Plugins:** Permitir que utilizadores avançados escrevam os seus próprios scripts para criar comandos personalizados.  
* **Integrações Externas:** Conectar com outras ferramentas (Google Calendar, sistemas de gestão escolar).
