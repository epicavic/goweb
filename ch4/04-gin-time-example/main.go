package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"serverTime": time.Now().UTC()})
	})
	r.Run("localhost:8080")
}

/*
$ go run main.go
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /                         --> main.main.func1 (3 handlers)
[GIN-debug] Listening and serving HTTP on localhost:8080
[GIN] 2021/02/22 - 10:39:54 | 200 |      40.318µs |       127.0.0.1 | GET      "/"

$ curl -i -w'\n' localhost:8080
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 22 Feb 2021 08:39:54 GMT
Content-Length: 44

{"serverTime":"2021-02-22T08:39:54.289457Z"}
*/
