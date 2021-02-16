package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type fastest struct {
	FastestMirror string        `json:"fastest_mirror"`
	Latency       time.Duration `json:"latency"`
}

func main() {
	mirrors, err := readList("mirrors.list")
	if err != nil {
		log.Fatalf("readList: %s", err)
	}

	// for _, mirror := range *mirrors {
	// 	fmt.Println(mirror)
	// }

	// fmt.Println(findFastest(mirrors))

	fmt.Println("Starting server")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := findFastest(mirrors)
		respJSON, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
	})
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// findFastest reads the list of mirrors and returns fastest
func findFastest(mirrors *[]string) fastest {
	mirrorChan := make(chan string)
	latencyChan := make(chan time.Duration)

	for _, mirror := range *mirrors {
		go func(mirror string) {
			// log.Println("Started probing: ", mirror)
			start := time.Now()
			_, err := http.Get(mirror)
			if err != nil {
				log.Println(err)
				return
			}
			latency := time.Now().Sub(start)
			// log.Printf("Got the best mirror: %s with latency: %s", mirror, latency)
			mirrorChan <- mirror
			latencyChan <- latency
		}(mirror)
	}

	return fastest{<-mirrorChan, <-latencyChan}
}

// readList reads the file and returns pointer to a list of strings and error
func readList(path string) (*[]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var list []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}

	return &list, scanner.Err()
}

/* USAGE:
$ curl -i -w'\n' localhost:8080/
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 16 Feb 2021 12:56:36 GMT
Content-Length: 72

{"fastest_mirror":"http://ftp.by.debian.org/debian/","latency":81780824}
*/

/* TO-DO
how to init and share mirrors list ?
when called from http server goroutines are not cancelled after fastest has been found (is there a way to cancel them ?)
there is no limit on number of goroutines (when multiple calls are done server will use a lot of outgoing connections)
*/
