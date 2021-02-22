package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/emicklei/go-restful"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// trainTable create sql statement
	trainTable = `
	CREATE TABLE IF NOT EXISTS train (
           ID INTEGER PRIMARY KEY AUTOINCREMENT,
           DRIVER_NAME VARCHAR(64) NULL,
           OPERATING_STATUS BOOLEAN
        )
	`
	// stationTable create sql statement
	stationTable = `
	CREATE TABLE IF NOT EXISTS station (
          ID INTEGER PRIMARY KEY AUTOINCREMENT,
          NAME VARCHAR(64) NULL,
          OPENING_TIME TIME NULL,
          CLOSING_TIME TIME NULL
        )
	`
	// scheduleTable create sql statement
	scheduleTable = `
	CREATE TABLE IF NOT EXISTS schedule (
	  ID INTEGER PRIMARY KEY AUTOINCREMENT,
          TRAIN_ID INT,
          STATION_ID INT,
          ARRIVAL_TIME TIME,
          FOREIGN KEY (TRAIN_ID) REFERENCES train(ID),
          FOREIGN KEY (STATION_ID) REFERENCES station(ID)
        )
	`
)

// train holds train information
type train struct {
	ID              int
	DriverName      string `json:"driver"`
	OperatingStatus bool   `json:"status"`
}

// station holds station information
type station struct {
	ID          int
	Name        string
	OpeningTime time.Time
	ClosingTime time.Time
}

// schedule links both trains and stations
type schedule struct {
	ID          int
	TrainID     int
	StationID   int
	ArrivalTime time.Time
}

// declare global db var
var db *sql.DB

// InitDB sets up connection pool with global db var and initializes database with empty tables
// db pointer shouldn't be passed as argument if we're using global db var
func InitDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./trainapi.db") // values must be assigned and not initialized (otherwise local vars will take precedence)
	if err != nil {
		return err
	}
	// fmt.Printf("InitDB: type - %T, value - %v\n", db, db)

	tables := []string{trainTable, stationTable, scheduleTable}
	for _, table := range tables {
		statement, err := db.Prepare(table)
		if err != nil {
			return err
		}
		_, err = statement.Exec()
		if err != nil {
			log.Printf("Failed to create table %s, ERROR: %s", table, err)
			return err
		}
	}
	log.Println("All tables created/initialized successfully!")
	return nil
}

// Register adds paths and routes to container
func (t *train) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/v1/trains").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/{train-id}").To(t.getTrain))
	ws.Route(ws.POST("").To(t.createTrain))
	ws.Route(ws.DELETE("/{train-id}").To(t.removeTrain))
	container.Add(ws)
}

// POST http://localhost:8080/v1/trains
func (t train) createTrain(r *restful.Request, w *restful.Response) {
	decoder := json.NewDecoder(r.Request.Body)
	var b train
	decoder.Decode(&b)
	// log.Println(b.DriverName, b.OperatingStatus)

	// fmt.Printf("createTrain: type - %T, value - %v\n", db, db)
	statement, _ := db.Prepare("INSERT INTO train (DRIVER_NAME, OPERATING_STATUS) VALUES (?, ?)")
	result, err := statement.Exec(b.DriverName, b.OperatingStatus)
	if err != nil {
		log.Println(err)
		w.AddHeader("Content-Type", "text/plain")
		w.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	ID, _ := result.LastInsertId()
	b.ID = int(ID)
	w.WriteHeaderAndEntity(http.StatusCreated, b)
}

// GET http://localhost:8080/v1/trains/[ID]
func (t train) getTrain(r *restful.Request, w *restful.Response) {
	id := r.PathParameter("train-id")
	err := db.QueryRow("SELECT ID, DRIVER_NAME, OPERATING_STATUS FROM train WHERE ID=?", id).Scan(&t.ID, &t.DriverName, &t.OperatingStatus)
	if err != nil {
		log.Println(err)
		w.AddHeader("Content-Type", "text/plain")
		w.WriteErrorString(http.StatusNotFound, "Train could not be found.")
		return
	}
	w.WriteEntity(t)
}

// DELETE http://localhost:8080/v1/trains/[ID]
func (t train) removeTrain(r *restful.Request, w *restful.Response) {
	id := r.PathParameter("train-id")
	statement, _ := db.Prepare("DELETE FROM train WHERE ID=?")
	_, err := statement.Exec(id)
	if err != nil {
		log.Println(err)
		w.AddHeader("Content-Type", "text/plain")
		w.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// entrypoint
func main() {
	if err := InitDB(); err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("main: type - %T, value - %v\n", db, db)

	container := restful.NewContainer()
	container.Router(restful.CurlyRouter{})
	t := train{}
	t.Register(container)

	log.Println("start listening on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", container))
}

/*
// POST
$ curl -i -w '\n' http://localhost:8080/v1/trains -H 'cache-control: no-cache' -H 'content-type: application/json' -d '{"driver": "Veronica", "status": true}'
HTTP/1.1 201 Created
Content-Type: application/json
Date: Mon, 22 Feb 2021 07:49:26 GMT
Content-Length: 52

{
 "ID": 1,
 "driver": "Veronica",
 "status": true
}

// GET
$ curl -i -w '\n' http://localhost:8080/v1/trains/1
HTTP/1.1 200 OK
Content-Type: application/json
Date: Mon, 22 Feb 2021 07:50:16 GMT
Content-Length: 52

{
 "ID": 1,
 "driver": "Veronica",
 "status": true
}

// DELETE
$ curl -X DELETE -i -w '\n' http://localhost:8080/v1/trains/1
HTTP/1.1 200 OK
Date: Mon, 22 Feb 2021 07:57:15 GMT
Content-Length: 0

*/
