package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
)

// mux is a custom multiplexer
type mux struct{}

func (m mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		random(w, r)
		return
	}
	http.NotFound(w, r)
	return
}

func random(w http.ResponseWriter, r *http.Request) {
	bs := make([]byte, 10)
	_, err := rand.Read(bs)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprint(w, string(bs))
}
