package main

import (
	"fmt"
	"math/rand"
	"net/http"
)

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/f", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, rand.Float64()) })
	m.HandleFunc("/i", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, rand.Intn(1000)) })
	http.ListenAndServe("localhost:8080", m)
}

/*
~ $ curl -i -w'\n' localhost:8080/f
HTTP/1.1 200 OK
Date: Thu, 18 Feb 2021 11:12:09 GMT
Content-Length: 18
Content-Type: text/plain; charset=utf-8

0.4377141871869802
~ $ curl -i -w'\n' localhost:8080/i
HTTP/1.1 200 OK
Date: Thu, 18 Feb 2021 11:12:12 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

81
*/
