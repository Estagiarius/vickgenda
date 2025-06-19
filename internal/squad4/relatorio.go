package squad4

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RelatorioCmd = &cobra.Command{
	Use:   "relatorio",
	Short: "Gera relatórios sobre atividades, produtividade e desempenho.",
	Long:  `Permite gerar diferentes tipos de relatórios para fornecer insights ao professor.`,
}

var relatorioProdutividadeCmd = &cobra.Command{
	Use:   "produtividade [semanal|mensal|bimestral]",
	Short: "Gera um relatório de produtividade.",
	Long:  `Mostra um relatório sobre tarefas concluídas, tempo gasto em eventos, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		periodo := "geral"
		if len(args) > 0 {
			periodo = args[0]
		}
		fmt.Printf("O relatório de produtividade (%s) ainda não foi implementado.\n", periodo)
		fmt.Println("Este relatório mostrará dados como:")
		fmt.Println("- Número de tarefas criadas, concluídas, pendentes.")
		fmt.Println("- Tempo médio para completar tarefas.")
		fmt.Println("- Tempo gasto em reuniões/eventos.")
		fmt.Println("Consulte 'data_requirements_relatorio.md' para mais detalhes sobre os dados necessários.")
	},
}

var relatorioAcademicoCmd = &cobra.Command{
	Use:   "academico [turma <nome_turma>|disciplina <nome_disciplina>] [bimestre <num>]",
	Short: "Gera um relatório de desempenho acadêmico.",
	Long:  `Mostra um relatório sobre notas, progresso de alunos, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("O relatório acadêmico ainda não foi implementado.")
		fmt.Println("Este relatório mostrará dados como:")
		fmt.Println("- Médias de notas por turma/disciplina.")
		fmt.Println("- Distribuição de notas.")
		fmt.Println("- Alunos precisando de atenção.")
		fmt.Println("Consulte 'data_requirements_relatorio.md' para mais detalhes sobre os dados necessários.")
	},
}

var relatorioUsoConteudoCmd = &cobra.Command{
	Use:   "uso-conteudo [disciplina <nome_disciplina>]",
	Short: "Gera um relatório sobre o uso de conteúdo pedagógico.",
	Long:  `Mostra estatísticas sobre o banco de questões, criação de provas, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("O relatório de uso de conteúdo ainda não foi implementado.")
		fmt.Println("Este relatório mostrará dados como:")
		fmt.Println("- Número de questões por disciplina/tópico/dificuldade.")
		fmt.Println("- Frequência de uso de questões em provas.")
		fmt.Println("- Tempo médio para criar testes.")
		fmt.Println("Consulte 'data_requirements_relatorio.md' para mais detalhes sobre os dados necessários.")
	},
}

func init() {
	RelatorioCmd.AddCommand(relatorioProdutividadeCmd)
	RelatorioCmd.AddCommand(relatorioAcademicoCmd)
	RelatorioCmd.AddCommand(relatorioUsoConteudoCmd)
}
