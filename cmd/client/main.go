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
	fmt.Println("Starting Peril client...")
	const connectionString = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	fmt.Println("Successfully connected to the server!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(username)

	if err = pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, "pause."+username, routing.PauseKey, pubsub.Transient, handlerPause(gameState)); err != nil {
		log.Fatal(err)
	}

	if err = pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, "army_moves."+username, "army_moves.*", pubsub.Transient, handlerMove(gameState)); err != nil {
		log.Fatal(err)
	}

	stillGoing := true
	for stillGoing {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			if err = gameState.CommandSpawn(input); err != nil {
				fmt.Println(err)
			}
		case "move":
			move, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				break
			}

			fmt.Println(move.Player, " made a move to ", move.ToLocation)
			ch, err := connection.Channel()
			if err != nil {
				fmt.Println(err)
				break
			}

			if err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, "army_moves."+username, move); err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println("Move was published Successfully")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			stillGoing = false
		default:
			fmt.Println("That isn't one of the commands, try 'help' if you're unsure.")
		}
	}
}
