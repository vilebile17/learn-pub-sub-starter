package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vilebile17/peril/internal/gamelogic"
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

	gameLogsCH, _, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", "durable")
	if err != nil {
		log.Fatal(err)
	}
	defer gameLogsCH.Close()

	gamelogic.PrintServerHelp()
	stillGoing := true
	for stillGoing {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			fmt.Println("Pausing the game...")
			if err = pubsub.PublishJSON(AMQPch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
				log.Fatal(err)
			}
		case "resume":
			fmt.Println("Resuming the game...")
			if err = pubsub.PublishJSON(AMQPch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
				log.Fatal(err)
			}
		case "quit":
			fmt.Println("Quiting the game...")
			stillGoing = false
		default:
			fmt.Println("I didn't understand that command: ", input[0])
		}
	}
}
