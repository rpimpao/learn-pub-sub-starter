package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connStr := "amqp://guest:guest@localhost:5672"

	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Println("failed to connect to rabbitmq:", err)
		return
	}
	defer conn.Close()

	chann, err := conn.Channel()
	if err != nil {
		fmt.Println("failed to open a channel:", err)
		return
	}
	defer chann.Close()

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.DurableQueue)
	if err != nil {
		fmt.Println("failed to declare and bind to queue:", err)
		return
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("sending pause message")
			err := pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				fmt.Println("failed to publish message:", err)
				return
			}

		case "resume":
			fmt.Println("sending resume message")
			err := pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				fmt.Println("failed to publish message:", err)
				return
			}

		case "quit":
			fmt.Println("Closing Peril server")
			return
		default:
			fmt.Printf("command '%s' not recognized\n", words[0])
		}
	}
}
