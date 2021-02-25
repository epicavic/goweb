package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	scheme = "amqp"
	user   = "guest"
	pass   = "guest"
	host   = "localhost"
	port   = 5672
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

func main() {
	// try to dial
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
	defer conn.Close()

	channel, err := conn.Channel()
	handleError("Fetching channel failed", err)
	defer channel.Close()

	testQueue, err := channel.QueueDeclare(
		"test", // Name of the queue
		false,  // Message is persisted or not
		false,  // Delete message when unused
		false,  // Exclusive
		false,  // No Waiting time
		nil,    // Extra args
	)
	handleError("Queue creation failed", err)

	serverTime := time.Now()
	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(serverTime.String()),
	}

	err = channel.Publish(
		"",             // exchange
		testQueue.Name, // routing key(Queue)
		false,          // mandatory
		false,          // immediate
		message,
	)
	handleError("Failed to publish a message", err)
	log.Printf("Published a message to the queue: %v\n", serverTime)

	messages, err := channel.Consume(
		testQueue.Name, // queue
		"",             // consumer
		true,           // auto-acknowledge
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	handleError("Failed to register a consumer", err)

	go func() {
		for message := range messages {
			log.Printf("Received a message from the queue: %s", message.Body)
		}
	}()

	// blockForever()
}
