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

// fastest mirror responce struct
type fastest struct {
	FastestMirror string        `json:"fastest_mirror"`
	Latency       time.Duration `json:"latency"`
}

// globally scoped mirrors slice var
var mirrors []string

// read a list of mirrors into a slice
func init() {
	if err := readList("mirrors.list", &mirrors); err != nil {
		log.Fatalf("readList: %s", err)
	}
}

// findFastestHandler returns fastest mirror and latency struct
func findFastestHandler(w http.ResponseWriter, r *http.Request) {
	response := findFastest(&mirrors)
	respJSON, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

// findFastest reads the mirrors slice and returns fastest and latency
func findFastest(mirrors *[]string) fastest {
	mirrorChan := make(chan string)
	latencyChan := make(chan time.Duration)

	for _, mirror := range *mirrors {
		go func(mirror string) {
			// log.Println("Started probing: ", mirror)
			start := time.Now()
			_, err := http.Get(mirror)
			latency := time.Now().Sub(start)
			if err != nil {
				log.Println(err)
				return
			}

			// send results to channel canceling all other goroutines when function returns
			mirrorChan <- mirror
			latencyChan <- latency

			log.Printf("Got the best mirror: %s with latency: %s", mirror, latency)
		}(mirror)
	}

	return fastest{<-mirrorChan, <-latencyChan}
}

// readList reads the file into a mirrors slice
func readList(path string, list *[]string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		*list = append(*list, scanner.Text())
	}

	return scanner.Err()
}

func main() {
	// fmt.Println(mirrors)
	fmt.Println("Starting server")
	http.HandleFunc("/", findFastestHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
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
there is no limit on number of goroutines (when multiple calls are done server will use a lot of outgoing connections)
*/
