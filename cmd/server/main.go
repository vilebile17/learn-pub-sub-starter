package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vilebile17/peril/internal/pubsub"
	"github.com/vilebile17/peril/internal/routing"
)

func main() {
	fmt.Println("Starting Peril server...")
	const connectionString = "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	fmt.Println("Successfully started the server!")

	AMQPch, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	if err = pubsub.PublishJSON(AMQPch, routing.ExchangePerilDirect, routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		}); err != nil {
		log.Fatal(err)
	}
}
