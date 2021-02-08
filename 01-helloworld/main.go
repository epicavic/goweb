package main

import (
	"fmt"
	"net/http"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! from '%s' location\n", r.URL.Path)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe("localhost:8080", nil)
}
