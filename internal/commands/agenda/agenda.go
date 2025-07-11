/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AgendaCmd represents the agenda command
var AgendaCmd = &cobra.Command{
	Use:   "agenda",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("agenda called, logic to be implemented.")
		return nil
	},
}

func init() {
	// rootCmd.AddCommand(AgendaCmd) // This will be done in cmd/cli/cli.go

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// agendaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// agendaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
