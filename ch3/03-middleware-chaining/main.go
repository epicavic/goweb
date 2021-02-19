package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// city struct used for request unmarshaling (json body)
type city struct {
	Name string  `json:"name"`
	Area float64 `json:"area"`
}

// filterContentType function used for checking supported media type before main handler processing
func filterContentType(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("filterContentType: before content type processing")
		if r.Header.Get("Content-type") != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("415 - Unsupported Media Type. Please send application/json"))
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// setServerTimeCookie function used for setting time cookie after main handler processing
// cookie must be set before calling original handler (before header map is sent)
// http.SetCookie will silently drop the cookie header if name contains not allowed characters
func setServerTimeCookie(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := http.Cookie{Name: "ServerTimeUTC", Value: strconv.FormatInt(time.Now().Unix(), 10)}
		// fmt.Println(c.String())
		http.SetCookie(w, &c)
		log.Println("setServerTimeCookie: after setting cookie")
		handler.ServeHTTP(w, r)
	})
}

// handle function used as main request processing unit and contains business logic
func handle(w http.ResponseWriter, r *http.Request) {
	log.Println("handle: main request handler")
	if r.Method == "POST" {
		var c city
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			panic(err)
		}
		defer r.Body.Close()

		log.Printf("Got %s city with area of %f sq miles!\n", c.Name, c.Area)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("201 - Created"))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
	}
}

func main() {
	http.Handle("/", filterContentType(setServerTimeCookie(http.HandlerFunc(handle))))
	http.ListenAndServe("localhost:8080", nil)
}

/*
$ go run main.go
2021/02/19 16:28:13 filterContentType: before content type processing

2021/02/19 16:28:42 filterContentType: before content type processing
2021/02/19 16:28:42 setServerTimeCookie: after setting cookie
2021/02/19 16:28:42 handle: main request handler
2021/02/19 16:28:42 Got Korosten city with area of 42.310000 sq miles!

$ curl -i -w'\n' localhost:8080/
HTTP/1.1 415 Unsupported Media Type
Date: Fri, 19 Feb 2021 14:28:13 GMT
Content-Length: 58
Content-Type: text/plain; charset=utf-8

415 - Unsupported Media Type. Please send application/json

$ curl -i -w'\n' localhost:8080/city -d '{"name": "Korosten", "area": 42.31}' -H "Content-Type: application/json"
HTTP/1.1 200 OK
Set-Cookie: ServerTimeUTC=1613744922
Date: Fri, 19 Feb 2021 14:28:42 GMT
Content-Length: 13
Content-Type: text/plain; charset=utf-8

201 - Created
*/
