package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Processing request!")
	w.Write([]byte("OK"))
	log.Println("Finished processing request")
}

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/", handle)
	ml := handlers.LoggingHandler(os.Stdout, m)
	http.ListenAndServe("localhost:8080", ml)
}

/*
http://httpd.apache.org/docs/current/logs.html#common

$ go run main.go
2021/02/20 10:49:57 Processing request!
2021/02/20 10:49:57 Finished processing request
127.0.0.1 - - [20/Feb/2021:10:49:57 +0200] "GET / HTTP/1.1" 200 2

$ curl -i -w'\n' localhost:8080/
HTTP/1.1 200 OK
Date: Sat, 20 Feb 2021 08:49:57 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

OK
*/
