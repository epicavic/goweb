package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	router.GET("/uptime", uptime)
	router.GET("/diskuse/:name", diskuse)
	log.Fatalln(http.ListenAndServe("localhost:8080", router))
}

func execCmd(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Sprintln(err)
	}
	return string(out)
}

func uptime(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	response := execCmd("uptime")
	io.WriteString(w, response)
	return
}

func diskuse(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	response := execCmd("df", "-h", params.ByName("name"))
	io.WriteString(w, response)
	return
}

/*
~ $ curl -i -w'\n' localhost:8080/uptime
HTTP/1.1 200 OK
Date: Thu, 18 Feb 2021 12:33:03 GMT
Content-Length: 64
Content-Type: text/plain; charset=utf-8

 14:33:03  up   4:53,  3 users,  load average: 2.12, 1.92, 1.86

~ $ curl -i -w'\n' localhost:8080/diskuse/-l
HTTP/1.1 200 OK
Date: Thu, 18 Feb 2021 12:33:05 GMT
Content-Length: 199
Content-Type: text/plain; charset=utf-8

Filesystem      Size  Used Avail Use% Mounted on
/dev/disk1s5    466G   11G  333G   4% /
/dev/disk1s4    466G  2.1G  333G   1% /private/var/vm
/dev/disk1s3    466G  505M  333G   1% /Volumes/Recovery
*/
