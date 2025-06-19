package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// registerCmd representa o comando de registro
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Registra um novo usuário",
	Long:  `Permite que um novo usuário se registre no Vickgenda.`, // Traduzido
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Comando de registro chamado. A lógica de registro de usuário será implementada aqui.")
	},
}

// loginCmd representa o comando de login
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Autentica um usuário existente",
	Long:  `Autentica um usuário existente no Vickgenda.`, // Traduzido
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Comando de login chamado. A lógica de login de usuário será implementada aqui.")
	},
}

// logoutCmd representa o comando de logout
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Desconecta o usuário atual",
	Long:  `Desconecta o usuário atualmente autenticado do Vickgenda.`, // Traduzido
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Comando de logout chamado. A lógica de logout de usuário será implementada aqui.")
	},
}

// A função init adiciona os comandos de autenticação ao rootCmd.
// O Cobra descobre essas funções e as chama.
func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
