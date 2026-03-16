package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) routing.AckType {
	return func(ps routing.PlayingState) routing.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return routing.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, chann *amqp.Channel) func(gamelogic.ArmyMove) routing.AckType {
	return func(am gamelogic.ArmyMove) routing.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(am)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return routing.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(chann, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: am.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Println("failed to publish war recognition message:", err)
				return routing.NackRequeue
			}
			return routing.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return routing.NackDiscard
		default:
			return routing.NackDiscard
		}
	}
}

func handlerRecognitionOfWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) routing.AckType {
	return func(rw gamelogic.RecognitionOfWar) routing.AckType {
		defer fmt.Print("> ")

		outcome, _, _ := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return routing.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return routing.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
			return routing.Ack
		default:
			fmt.Print("error processing recognition of war, message discarded")
			return routing.NackDiscard
		}
	}
}
