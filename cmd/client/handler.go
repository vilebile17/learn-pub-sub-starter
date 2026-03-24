package main

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vilebile17/peril/internal/gamelogic"
	"github.com/vilebile17/peril/internal/pubsub"
	"github.com/vilebile17/peril/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		fmt.Println("Just Acked")
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutcomeSafe:
			fmt.Println("Just Acked")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				}); err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			fmt.Println("Can't wage war on yourself...")
			return pubsub.NackDiscard
		default:
			fmt.Println("Just Nacked and discarded")
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(publishCh *amqp.Channel, username, message string) error {
	gl := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    username,
	}

	if err := pubsub.PublishGob(publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		gl); err != nil {
		return err
	}

	return nil
}

func handlerWar(publishCh *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(rw)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			fmt.Printf("%v won a war against %v", rw.Attacker.Username, rw.Defender.Username)
		case gamelogic.WarOutcomeYouWon:
			fmt.Printf("%v won a war against %v", rw.Defender.Username, rw.Attacker.Username)
		case gamelogic.WarOutcomeDraw:
			fmt.Printf("A war between %v and %v resulted in a draw", rw.Defender.Username, rw.Attacker.Username)
		default:
			fmt.Printf("Err, how did you get a different outcome: %v\n", outcome)
			return pubsub.NackDiscard
		}

		if err := publishGameLog(publishCh,
			gs.Player.Username,
			fmt.Sprintf("Handled a war between %v and %v", rw.Attacker.Username, rw.Defender.Username)); err != nil {
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
