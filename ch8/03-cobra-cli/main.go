package main

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Short desc",
	Long:  `Long desc`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name := cmd.PersistentFlags().Lookup("name").Value
		age := cmd.PersistentFlags().Lookup("age").Value
		log.Printf("Hello %s - %s years", name, age)
	},
}

// Execute is Cobra logic startpoint
func Execute() {
	rootCmd.PersistentFlags().StringP("name", "n", "stranger", "your wonderful name")
	rootCmd.PersistentFlags().IntP("age", "a", 0, "your graceful age")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Entrypoint
func main() {
	Execute()
}

/*
$ go run main.go
2021/02/24 16:46:12 Hello stranger - 0 years

$ go run main.go -n noname -a 1
2021/02/24 16:46:09 Hello noname - 1 years

$ go run main.go app -n noname -a 1
2021/02/24 16:48:02 Hello noname - 1 years

$ go run main.go -h
Long desc

Usage:
  app [flags]

Flags:
  -a, --age int       your graceful age
  -h, --help          help for app
  -n, --name string   your wonderful name (default "stranger")
*/
