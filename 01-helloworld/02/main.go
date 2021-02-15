package main

import (
	"fmt"
	"net/http"
	"time"
)

func dynamicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello from dynamic ", time.Now().String())
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "s.html")
}

func main() {
	fmt.Println("Started server")
	http.HandleFunc("/static", staticHandler)
	http.HandleFunc("/", dynamicHandler)
	http.ListenAndServe("localhost:8080", nil)
}
