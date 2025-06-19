package components

import (
	"fmt"
	"strings"
)

// RenderTable recebe cabeçalhos e linhas e retorna uma string representando uma tabela formatada.
// Calcula a largura das colunas com base na string mais longa em cada coluna (incluindo cabeçalhos).
// Usa espaços para preenchimento, '|' para separadores verticais e '-' para linhas horizontais.
func RenderTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	numCols := len(headers)
	colWidths := make([]int, numCols)

	// Calcula a largura inicial das colunas com base nos cabeçalhos
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Ajusta a largura das colunas com base no conteúdo das linhas
	for _, row := range rows {
		if len(row) != numCols {
			// Ou lida com o erro apropriadamente
			fmt.Println("Aviso: O comprimento da linha não corresponde ao comprimento do cabeçalho. Pulando linha:", row)
			continue
		}
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var builder strings.Builder

	// Função auxiliar para construir uma string de linha
	buildRowString := func(cells []string, isHeader bool) string {
		var rowBuilder strings.Builder
		rowBuilder.WriteString("|")
		for i, cell := range cells {
			padding := strings.Repeat(" ", colWidths[i]-len(cell))
			if isHeader {
				rowBuilder.WriteString(fmt.Sprintf(" %s%s |", cell, padding))
			} else {
				rowBuilder.WriteString(fmt.Sprintf(" %s%s |", cell, padding))
			}
		}
		return rowBuilder.String()
	}

	// Constrói a linha de cabeçalho
	builder.WriteString(buildRowString(headers, true))
	builder.WriteString("\n")

	// Constrói a linha separadora
	builder.WriteString("|")
	for _, width := range colWidths {
		builder.WriteString(strings.Repeat("-", width+2)) // +2 para espaços ao redor do conteúdo
		builder.WriteString("|")
	}
	builder.WriteString("\n")

	// Constrói as linhas de dados
	for _, row := range rows {
		if len(row) == numCols { // Garante que a linha tenha o número correto de colunas
			builder.WriteString(buildRowString(row, false))
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// Função main temporária para testar RenderTable.
// Para executar: go run internal/tui/components/table.go
/*
func main() {
	headers := []string{"ID", "Name", "Role"}
	rows := [][]string{
		{"1", "Alice", "Developer"},
		{"2", "Bob", "Designer"},
		{"10", "Charles", "Project Manager"},
		{"111", "Diana", "QA Engineer Long Name"},
	}
	table := RenderTable(headers, rows)
	fmt.Print(table)

	headers2 := []string{"Task", "Status"}
	rows2 := [][]string{
		{"Implement feature X", "In Progress"},
		{"Fix bug Y", "Done"},
		{"Write documentation", "To Do"},
	}
	table2 := RenderTable(headers2, rows2)
	fmt.Print(table2)

	// Test with mismatched row length
    headers3 := []string{"Fruit", "Color"}
    rows3 := [][]string{
        {"Apple", "Red"},
        {"Banana"}, // Mismatched
        {"Cherry", "Red"},
    }
    table3 := RenderTable(headers3, rows3)
    fmt.Print(table3)

    // Test with empty rows
    table4 := RenderTable(headers, [][]string{})
    fmt.Print(table4)

	// Test with empty headers (should return empty string)
	table5 := RenderTable([]string{}, [][]string{})
	fmt.Print(table5)
	fmt.Println("Test empty headers done.")
}
*/
