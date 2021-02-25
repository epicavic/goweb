package main

import (
	"fmt"
	"log"
	"os"

	"github.com/levigross/grequests"
	"github.com/urfave/cli"
)

// Repo struct for holding response of repositories fetch API
type repo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Forks    int    `json:"forks"`
	Private  bool   `json:"private"`
}

// getStats fetches the repos for the given Github user
func getStats(url string) *grequests.Response {
	// request can be modified by passing an optional RequestOptions struct
	resp, err := grequests.Get(url, nil)
	if err != nil {
		log.Fatal("Unable to make request: ", err)
	}
	return resp
}

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name:    "fetch",
			Aliases: []string{"f"},
			Usage:   "[Usage]: appname fetch user",
			Action: func(c *cli.Context) error {
				if c.NArg() > 0 {
					// Github API Logic
					var repos []repo
					user := c.Args().Get(0)
					var repoURL = fmt.Sprintf("https://api.github.com/users/%s/repos", user)
					resp := getStats(repoURL)
					resp.JSON(&repos)
					log.Println(repos)
				} else {
					log.Println("Please give a username. See -h to see help")
				}
				return nil
			},
		},
	}

	app.Version = "1.0"
	app.Run(os.Args)
}

/*
$ go run main.go
NAME:
   main - A new cli application

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   1.0

COMMANDS:
   fetch, f  [Usage]: appname fetch user
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)

$ go run main.go f
2021/02/25 13:18:23 Please give a username. See -h to see help

$ go run main.go f epicavic
2021/02/25 13:12:24 [{339015075 golang-web-dev epicavic/golang-web-dev 0 false} {268479418 goplay epicavic/goplay 0 false} {328367958 goweb epicavic/goweb 0 false} {339108695 Hands-On-Restful-Web-services-with-Go epicavic/Hands-On-Restful-Web-services-with-Go 0 false}]
*/
