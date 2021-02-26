/*
The server should save information to the database as one operation.
It should send an email to the given email address.
It should perform a long-running job and POST the result to a callback (web-hook).

The client can fire an API and receive a job ID back.
The job is pushed onto a queue with the respective message format.
A worker picks the job and starts performing it.
The worker saves the result on various endpoints and sends the status to the database.
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

const (
	scheme = "amqp"
	user   = "guest"
	pass   = "guest"
	host   = "localhost"
	port   = 5672
	queue  = "job"
)

// handleError handles error checking
func handleError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// blockForever wait forever to receive from channel
func blockForever() {
	c := make(chan struct{})
	<-c
}

// Job represents a job item
type Job struct {
	ID        uuid.UUID   `json:"uuid"`
	Type      string      `json:"type"`
	ExtraData interface{} `json:"extra_data"`
}

// Log is for worker-A (saves information from the message into a database)
type Log struct {
	ClientTime time.Time `json:"client_time"`
}

// CallBack is for worker-B (posts information to a callback that was received as part of a request)
type CallBack struct {
	CallBackURL string `json:"callback_url"`
}

// Mail is for worker-C (sends an email)
type Mail struct {
	EmailAddress string `json:"email_address"`
}

// Workers does the job. It holds a queue connection
type Workers struct {
	conn *amqp.Connection
}

// run method reads from message queue and fires up worker(A|B|C) when new message arrives
func (w *Workers) run() {
	log.Printf("Starting up workers...")
	channel, err := w.conn.Channel()
	handleError("Fetching channel failed", err)
	defer channel.Close()

	jobQueue, err := channel.QueueDeclare(
		queue, // Name of the queue
		false, // Message is persisted or not
		false, // Delete message when unused
		false, // Exclusive
		false, // No Waiting time
		nil,   // Extra args
	)
	handleError("Job queue fetch failed", err)

	messages, err := channel.Consume(
		jobQueue.Name, // queue name
		"",            // consumer
		true,          // auto-acknowledge
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	handleError("Messages fetch failed", err)

	go func() {
		for message := range messages {

			job := Job{}
			err = json.Unmarshal(message.Body, &job)
			handleError("Unable to load queue message", err)
			log.Printf("Workers received a message from the queue: %s", job)

			switch job.Type {
			case "A":
				w.dbWork(job)
			case "B":
				w.callbackWork(job)
			case "C":
				w.emailWork(job)
			}
		}
	}()
	log.Printf("All workers are up and running")

	blockForever()
}

// worker-A
func (w *Workers) dbWork(job Job) {
	result := job.ExtraData.(map[string]interface{})
	log.Printf("Worker %s: extracting data..., JOB: %s", job.Type, result)
	time.Sleep(2 * time.Second)
	log.Printf("Worker %s: saving data to database..., JOB: %s", job.Type, job.ID)
}

// worker-B
func (w *Workers) callbackWork(job Job) {
	log.Printf("Worker %s: performing some long running process..., JOB: %s", job.Type, job.ID)
	time.Sleep(2 * time.Second)
	log.Printf("Worker %s: posting the data back to the given callback..., JOB: %s", job.Type, job.ID)
}

// worker-C
func (w *Workers) emailWork(job Job) {
	log.Printf("Worker %s: sending the email..., JOB: %s", job.Type, job.ID)
	time.Sleep(2 * time.Second)
	log.Printf("Worker %s: sent the email successfully, JOB: %s", job.Type, job.ID)
}

// JobServer holds api handler functions
type JobServer struct {
	Queue   amqp.Queue
	Channel *amqp.Channel
	Conn    *amqp.Connection
}

func (s *JobServer) publish(jsonBody []byte) error {
	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonBody,
	}
	err := s.Channel.Publish(
		"",    // exchange
		queue, // routing key(Queue)
		false, // mandatory
		false, // immediate
		message,
	)
	handleError("Failed to push message to the queue", err)

	return err
}

func (s *JobServer) asyncDBHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.NewRandom()
	handleError("Error while generating new UUID", err)

	queryParams := r.URL.Query()
	unixTime, err := strconv.ParseInt(queryParams.Get("client_time"), 10, 64)
	handleError("Error while converting client_time", err)

	clientTime := time.Unix(unixTime, 0)
	jsonBody, err := json.Marshal(Job{ID: jobID, Type: "A", ExtraData: Log{ClientTime: clientTime}})
	handleError("JSON body creation failed", err)

	if err := s.publish(jsonBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBody)
}

func (s *JobServer) asyncCallbackHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.NewRandom()
	handleError("Error while generating new UUID", err)

	jsonBody, err := json.Marshal(Job{ID: jobID, Type: "B", ExtraData: ""})
	handleError("JSON body creation failed", err)

	if err := s.publish(jsonBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBody)
}

func (s *JobServer) asyncMailHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.NewRandom()
	handleError("Error while generating new UUID", err)

	jsonBody, err := json.Marshal(Job{ID: jobID, Type: "C", ExtraData: ""})
	handleError("JSON body creation failed", err)

	if err := s.publish(jsonBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBody)
}

func getServer(queue string) JobServer {
	sleepPeriod := 5 * time.Second
	connectionString := fmt.Sprintf("%s://%s:%s@%s:%d/", scheme, user, pass, host, port)
	var conn *amqp.Connection
	var err error
	for {
		conn, err = amqp.Dial(connectionString)
		if err == nil {
			break
		}
		fmt.Printf("Waiting for RabbitMQ. Sleeping for %v\n", sleepPeriod)
		time.Sleep(sleepPeriod)
	}
	// defer conn.Close() // should be closed in main func

	channel, err := conn.Channel()
	handleError("Fetching channel failed", err)
	// defer channel.Close() // should be closed in main func

	jobQueue, err := channel.QueueDeclare(
		queue, // Name of the queue
		false, // Message is persisted or not
		false, // Delete message when unused
		false, // Exclusive
		false, // No Waiting time
		nil,   // Extra args
	)
	handleError("Job queue creation failed", err)

	return JobServer{Conn: conn, Channel: channel, Queue: jobQueue}
}

func main() {
	jobServer := getServer(queue)

	go func(conn *amqp.Connection) {
		workerProcess := Workers{
			conn: jobServer.Conn,
		}
		workerProcess.run()
	}(jobServer.Conn)

	router := mux.NewRouter()
	router.HandleFunc("/job/database", jobServer.asyncDBHandler)
	router.HandleFunc("/job/mail", jobServer.asyncMailHandler)
	router.HandleFunc("/job/callback", jobServer.asyncCallbackHandler)

	httpServer := &http.Server{
		Handler:      router,
		Addr:         "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(httpServer.ListenAndServe())

	// close opened resources
	defer jobServer.Channel.Close()
	defer jobServer.Conn.Close()
}

/*
$ bash -x run.sh
+ trap ctrl_c INT
+ docker run --rm --name rabbitmq -p 5672:5672 -d rabbitmq
6b517daa1d9ccddbe1f20379d52aa7d194b02b74e292e9823ee768e9eeb66e2c
+ go run main.go
Waiting for RabbitMQ. Sleeping for 5s
Waiting for RabbitMQ. Sleeping for 5s
2021/02/26 11:42:07 Starting up workers...
2021/02/26 11:42:07 All workers are up and running

$ curl -i -w'\n' 'localhost:8080/job/database?client_time='"$(date +%s)"
HTTP/1.1 200 OK
Date: Fri, 26 Feb 2021 09:48:48 GMT
Content-Length: 115
Content-Type: text/plain; charset=utf-8

{"uuid":"eac5888c-b78c-4f5d-93e5-0d3452750fe5","type":"A","extra_data":{"client_time":"2021-02-26T11:48:48+02:00"}}

2021/02/26 11:48:48 Workers received a message from the queue: {eac5888c-b78c-4f5d-93e5-0d3452750fe5 A map[client_time:2021-02-26T11:48:48+02:00]}
2021/02/26 11:48:48 Worker A: extracting data..., JOB: map[client_time:2021-02-26T11:48:48+02:00]
2021/02/26 11:48:50 Worker A: saving data to database..., JOB: eac5888c-b78c-4f5d-93e5-0d3452750fe5

$ curl -i -w'\n' 'localhost:8080/job/callback'
HTTP/1.1 200 OK
Date: Fri, 26 Feb 2021 09:50:35 GMT
Content-Length: 74
Content-Type: text/plain; charset=utf-8

{"uuid":"2770cff0-732a-441d-9b12-34676c2ee8a7","type":"B","extra_data":""}

2021/02/26 11:50:35 Workers received a message from the queue: {2770cff0-732a-441d-9b12-34676c2ee8a7 B }
2021/02/26 11:50:35 Worker B: performing some long running process..., JOB: 2770cff0-732a-441d-9b12-34676c2ee8a7
2021/02/26 11:50:37 Worker B: posting the data back to the given callback..., JOB: 2770cff0-732a-441d-9b12-34676c2ee8a7

$ curl -i -w'\n' 'localhost:8080/job/mail'
HTTP/1.1 200 OK
Date: Fri, 26 Feb 2021 09:50:40 GMT
Content-Length: 74
Content-Type: text/plain; charset=utf-8

{"uuid":"69750d6d-dd5e-4b9f-8498-3afe01a7b30b","type":"C","extra_data":""}

2021/02/26 11:50:40 Workers received a message from the queue: {69750d6d-dd5e-4b9f-8498-3afe01a7b30b C }
2021/02/26 11:50:40 Worker C: sending the email..., JOB: 69750d6d-dd5e-4b9f-8498-3afe01a7b30b
2021/02/26 11:50:42 Worker C: sent the email successfully, JOB: 69750d6d-dd5e-4b9f-8498-3afe01a7b30b
*/
