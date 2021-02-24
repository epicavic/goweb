package main

import (
	"flag"
	"log"
)

var name string
var age int

// use pointers to preserve flag values
func init() {
	flag.StringVar(&name, "name", "stranger", "your wonderful name")
	flag.IntVar(&age, "age", 0, "your graceful age")
}

func main() {
	flag.Parse()
	log.Printf("Hello %s - %d years", name, age)
}

/*
$ go run main.go
2021/02/24 15:21:49 Hello stranger - 0 years

$ go run main.go -name noname -age 1
2021/02/24 15:22:11 Hello noname - 1 years

$ go run main.go -h
Usage of /var/folders/bz/xvk1v1vx7vb99ts2g3jv72gh0000gn/T/go-build272807496/b001/exe/main:
  -age int
        your graceful age
  -name string
        your wonderful name (default "stranger")
*/
