/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TarefaCmd represents the tarefa command
var TarefaCmd = &cobra.Command{
	Use:   "tarefa",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("tarefa called, logic to be implemented.")
		return nil
	},
}

func init() {
	// rootCmd.AddCommand(TarefaCmd) // This will be done in cmd/cli/cli.go

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tarefaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tarefaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
