package main

import (
	"fmt"
	"log"
	"net/http"
)

// A Handler responds to an HTTP request. https://golang.org/pkg/net/http/#Handler
// type Handler interface { ServeHTTP(ResponseWriter, *Request) }
func middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Execute middleware before request phase")
		handler.ServeHTTP(w, r) // Execute handle func
		fmt.Println("Execute middleware after request phase")
	})
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Execute main handle")
	w.Write([]byte("OK"))
}

// Handle registers the handler for the given pattern in the DefaultServeMux. https://golang.org/pkg/net/http/#Handle
// func Handle(pattern string, handler Handler)
// HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers(type conversion). https://golang.org/pkg/net/http/#HandlerFunc
// type HandlerFunc func(ResponseWriter, *Request)
// func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
func main() {
	http.Handle("/", middleware(http.HandlerFunc(handle)))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

/*
In an application with no middleware, a request reaches the API server and gets handled by a function handler directly.
The response is immediately sent back from the server, and the client receives it.
But in applications with middleware configured to a function handler, it can pass through a set of stages, such as
logging, authentication, session validation, and so on, and then proceeds to the business logic.
*/

/*
$ go run main.go
Execute middleware before request phase
Execute main handle
Execute middleware after request phase

~ $ curl -i -w'\n' localhost:8080/
HTTP/1.1 200 OK
Date: Fri, 19 Feb 2021 07:54:15 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

OK
*/
