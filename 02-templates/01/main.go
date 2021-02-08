// print main.go to stdout
package main

import (
	"log"
	"os"
	"text/template"
)

var t *template.Template

func init() {
	// template.Must - does errorchecking. panics if the error is non-nil
	// template.ParseFiles - creates a new Template and parses the template definitions from the named files
	t = template.Must(template.ParseFiles("main.go"))
}

func main() {
	if err := t.Execute(os.Stdout, nil); err != nil {
		// fmt.Println("msg") and os.Exit(1)
		log.Fatalln("Failed to execute template")
	}
}
