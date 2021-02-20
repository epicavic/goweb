package main

import (
	"fmt"
	"net/http"
	"time"

	rest "github.com/emicklei/go-restful"
)

func timeNow(r *rest.Request, w *rest.Response) {
	fmt.Fprint(w, time.Now().String())
}

func main() {
	svc := new(rest.WebService)
	svc.Route(svc.GET("/").To(timeNow))
	rest.Add(svc)
	http.ListenAndServe("localhost:8080", nil)
}

/*
$ curl -i -w'\n' localhost:8080/
HTTP/1.1 200 OK
Date: Sat, 20 Feb 2021 09:44:03 GMT
Content-Length: 51
Content-Type: text/plain; charset=utf-8

2021-02-20 11:44:03.437913 +0200 EET m=+1.493379499
*/
