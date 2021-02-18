package main

import (
	"log"
	"net/http"
)

func main() {
	// compiler would complain on undeclared name if we won't run all go files
	var m mux
	log.Fatal(http.ListenAndServe("localhost:8080", m))
}

/*
$ go run *.go

$ curl -i -w'\n' localhost:8080/
HTTP/1.1 200 OK
Date: Thu, 18 Feb 2021 10:48:47 GMT
Content-Length: 10
Content-Type: application/octet-stream

We3VAw��
*/
