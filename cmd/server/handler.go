package main

import (
	"fmt"

	"github.com/vilebile17/peril/internal/gamelogic"
	"github.com/vilebile17/peril/internal/pubsub"
	"github.com/vilebile17/peril/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")

		if err := gamelogic.WriteLog(gl); err != nil {
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
