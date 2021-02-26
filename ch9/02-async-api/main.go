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

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

// rabbitmq connection details
const (
	rabbitScheme = "amqp"
	rabbitUser   = "guest"
	rabbitPass   = "guest"
	rabbitHost   = "localhost"
	rabbitPort   = 5672
	rabbitQueue  = "job"

	redisScheme = "redis"
	redisHost   = "localhost"
	redisPort   = 6379
	redisDB     = 0 // default database
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

// Log is for worker-A (saves information from the message into a database)
type Log struct {
	ClientTime time.Time `json:"client_time"` // timestamp provided by client (parsed from http request query params)
}

// CallBack is for worker-B (posts information to a callback that was received as part of a request)
type CallBack struct {
	CallBackURL string `json:"callback_url"` // callback URL to post data to
}

// Mail is for worker-C (sends an email)
type Mail struct {
	EmailAddress string `json:"email_address"` // email address to send a message to
}

// Job represents a job item
type Job struct {
	ID        uuid.UUID   `json:"uuid"`       // set a job ID
	Type      string      `json:"type"`       // set a job type "A", "B", or "C"
	ExtraData interface{} `json:"extra_data"` // set extra data. used as placeholder for for Log, Callback, and Mail structs
}

// Workers does the job
type Workers struct {
	RabbitConn  *amqp.Connection // holds a queue connection. this way workers can read messages from the queue
	RedisClient *redis.Client    // holds a cache connection. this way workers can write job status to the cache
}

// run method reads from message queue and fires up worker(A|B|C) when new message arrives
func (w *Workers) run() {
	log.Printf("Starting up workers...")
	channel, err := w.RabbitConn.Channel()
	handleError("Fetching channel failed", err)
	defer channel.Close()

	jobQueue, err := channel.QueueDeclare(
		rabbitQueue, // Name of the queue
		false,       // Message is persisted or not
		false,       // Delete message when unused
		false,       // Exclusive
		false,       // No Waiting time
		nil,         // Extra args
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

	// fire up goroutine which reads messages from the queue and launches the workers
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

	// keep listening to the queue forever
	blockForever()
}

// worker-A mock
func (w *Workers) dbWork(job Job) {
	// fmt.Println(job.ExtraData) // map[client_time:2021-02-26T12:55:55+02:00]
	// x.(T), if the asserted type T is a concrete type, then the type assertion checks whether x’s dynamic type is identical to T
	// if this check succeeds, the result of the type assertion is x’s dynamic value, whose type is of course T
	// In other words, a type assertion to a concrete type extracts the concrete value from its operand
	// If the check fails, then the operation panics
	result := job.ExtraData.(map[string]interface{}) // interface type assertion
	log.Printf("Worker %s: extracting data..., JOB: %s", job.Type, result)
	w.RedisClient.Set(w.RedisClient.Context(), job.ID.String(), "STARTED", 0)
	time.Sleep(2 * time.Second)
	log.Printf("Worker %s: saving data to database..., JOB: %s", job.Type, job.ID)
	w.RedisClient.Set(w.RedisClient.Context(), job.ID.String(), "DONE", 0)
}

// worker-B mock
func (w *Workers) callbackWork(job Job) {
	log.Printf("Worker %s: performing some long running process..., JOB: %s", job.Type, job.ID)
	w.RedisClient.Set(w.RedisClient.Context(), job.ID.String(), "STARTED", 0)
	time.Sleep(30 * time.Second)
	log.Printf("Worker %s: posting the data back to the given callback..., JOB: %s", job.Type, job.ID)
	w.RedisClient.Set(w.RedisClient.Context(), job.ID.String(), "DONE", 0)
}

// worker-C mock
func (w *Workers) emailWork(job Job) {
	log.Printf("Worker %s: sending the email..., JOB: %s", job.Type, job.ID)
	w.RedisClient.Set(w.RedisClient.Context(), job.ID.String(), "STARTED", 0)
	time.Sleep(2 * time.Second)
	log.Printf("Worker %s: sent the email successfully, JOB: %s", job.Type, job.ID)
	w.RedisClient.Set(w.RedisClient.Context(), job.ID.String(), "DONE", 0)
}

// JobServer holds api handler functions
type JobServer struct {
	Queue       amqp.Queue       // holds a queue itself
	Channel     *amqp.Channel    // holds a queue channel
	RabbitConn  *amqp.Connection // holds a queue connection. this way handlers can write messages to the queue
	RedisClient *redis.Client    // holds a cache connection. this way handlers can read job status from the cache
}

// publish job json to the queue
func (s *JobServer) publish(jsonBody []byte) error {
	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonBody,
	}
	err := s.Channel.Publish(
		"",          // exchange
		rabbitQueue, // routing key(Queue)
		false,       // mandatory
		false,       // immediate
		message,
	)
	handleError("Failed to push message to the queue", err)

	return err
}

// Takes an incoming request and tries to create an instant Job ID and populate with extra_data
// Once it successfully places the job in the queue, it returns the Job ID to the caller
// The workers who are already started and listening to the job queue pick those tasks and execute them concurrently
// Work Type A handler
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

// Work Type B handler
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

// Work Type C handler
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

// Work status handler
func (s *JobServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	uuid := queryParams.Get("uuid")
	w.Header().Set("Content-Type", "application/json")
	jobStatus := s.RedisClient.Get(s.RedisClient.Context(), uuid)
	status := map[string]string{"ID": uuid, "Status": jobStatus.Val()}
	response, err := json.Marshal(status)
	handleError("Cannot create response for client", err)

	w.Write(response)
}

// getServer establishes rabbitmq and redis connections
func getServer() JobServer {
	sleepPeriod := 5 * time.Second
	rabbitConnectionString := fmt.Sprintf("%s://%s:%s@%s:%d/", rabbitScheme, rabbitUser, rabbitPass, rabbitHost, rabbitPort)
	redisConnectionString := fmt.Sprintf("%s://%s:%d/%d", redisScheme, redisHost, redisPort, redisDB)
	var rabbitConn *amqp.Connection
	var redisClient *redis.Client
	var err error
	for {
		rabbitConn, err = amqp.Dial(rabbitConnectionString)
		if err == nil {
			break
		}
		fmt.Printf("Waiting for RabbitMQ. Sleeping for %v\n", sleepPeriod)
		time.Sleep(sleepPeriod)
	}
	for {
		opt, err := redis.ParseURL(redisConnectionString)
		handleError("Failed to parse redis options", err)
		redisClient = redis.NewClient(opt)
		_, err = redisClient.Ping(redisClient.Context()).Result()
		if err == nil {
			break
		}
		fmt.Printf("Waiting for Redis. Sleeping for %v\n", sleepPeriod)
		time.Sleep(sleepPeriod)
	}

	// defer conn.Close() // should be closed in main func

	rabbitChan, err := rabbitConn.Channel()
	handleError("Fetching channel failed", err)
	// defer channel.Close() // should be closed in main func

	jobQueue, err := rabbitChan.QueueDeclare(
		rabbitQueue, // Name of the queue
		false,       // Message is persisted or not
		false,       // Delete message when unused
		false,       // Exclusive
		false,       // No Waiting time
		nil,         // Extra args
	)
	handleError("Job queue creation failed", err)

	return JobServer{RabbitConn: rabbitConn, Channel: rabbitChan, Queue: jobQueue, RedisClient: redisClient}
}

func main() {
	jobServer := getServer()

	go func(conn JobServer) {
		workerProcess := Workers{
			RabbitConn:  conn.RabbitConn,
			RedisClient: conn.RedisClient,
		}
		workerProcess.run()
	}(jobServer)

	router := mux.NewRouter()
	router.HandleFunc("/job/database", jobServer.asyncDBHandler)
	router.HandleFunc("/job/mail", jobServer.asyncMailHandler)
	router.HandleFunc("/job/callback", jobServer.asyncCallbackHandler)
	router.HandleFunc("/job/status", jobServer.statusHandler)

	httpServer := &http.Server{
		Handler:      router,
		Addr:         "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(httpServer.ListenAndServe())

	// close opened resources
	defer jobServer.Channel.Close()
	defer jobServer.RabbitConn.Close()
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

$ curl -i -w'\n' 'localhost:8080/job/callback'
HTTP/1.1 200 OK
Date: Fri, 26 Feb 2021 13:57:54 GMT
Content-Length: 74
Content-Type: text/plain; charset=utf-8

{"uuid":"55972329-8b4e-4706-a814-648e9112bc12","type":"B","extra_data":""}

$ curl -i -w'\n' 'localhost:8080/job/status?uuid=55972329-8b4e-4706-a814-648e9112bc12'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 26 Feb 2021 13:58:07 GMT
Content-Length: 64

{"ID":"55972329-8b4e-4706-a814-648e9112bc12","Status":"STARTED"}

$ curl -i -w'\n' 'localhost:8080/job/status?uuid=55972329-8b4e-4706-a814-648e9112bc12'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 26 Feb 2021 13:59:12 GMT
Content-Length: 61

{"ID":"55972329-8b4e-4706-a814-648e9112bc12","Status":"DONE"}
*/
