package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func timestamp(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, time.Now().String())
}

func main() {
	http.HandleFunc("/", timestamp)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

/*
curl -i -w'\n' localhost:8080/
HTTP/1.1 200 OK
Date: Thu, 18 Feb 2021 10:02:40 GMT
Content-Length: 51
Content-Type: text/plain; charset=utf-8

2021-02-18 12:02:40.129618 +0200 EET m=+3.721037417
*/
