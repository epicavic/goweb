package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	// create new app
	app := cli.NewApp()

	// add app flags
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "name",
			Aliases: []string{"n"},
			EnvVars: []string{"APP_NAME"},
			Value:   "stranger",
			Usage:   "your wonderful name",
		},
		&cli.IntFlag{
			Name:    "age",
			Aliases: []string{"a"},
			EnvVars: []string{"APP_AGE"},
			Value:   0,
			Usage:   "your graceful age",
		},
	}
	// parse and bring data in cli.Context struct
	app.Action = func(c *cli.Context) error {
		// c.String, c.Int looks for value of given flag
		log.Printf("Hello %s - %d years", c.String("name"), c.Int("age"))
		return nil
	}
	// pass os.Args to cli app to parse content
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

/*
$ go run main.go
2021/02/24 15:45:59 Hello stranger - 0 years

$ go run main.go -name noname -age 1
2021/02/24 15:46:10 Hello noname - 1 years

$ go run main.go -n noname -a 1
2021/02/24 15:46:17 Hello noname - 1 years

$ APP_NAME=noname APP_AGE=1 go run main.go
2021/02/24 15:46:37 Hello noname - 1 years

$ go run main.go -h
NAME:
   main - A new cli application

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --name value, -n value  your wonderful name (default: "stranger") [$APP_NAME]
   --age value, -a value   your graceful age (default: 0) [$APP_AGE]
   --help, -h              show help (default: false)
*/
