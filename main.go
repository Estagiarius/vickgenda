package main

import (
	"vickgenda-cli/cmd" // Updated to use the new cmd package
)

func main() {
	// db.InitDB() // Initialization will be handled by individual commands or a PersistentPreRun later
	cmd.Execute()
}
