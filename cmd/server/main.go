package main

import (
	"encoding/json"
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

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("sending pause message")
			data, err := json.Marshal(routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				fmt.Println("failed to marshal data:", err)
				return
			}

			pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, data)

		case "resume":
			fmt.Println("sending resume message")
			data, err := json.Marshal(routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				fmt.Println("failed to marshal data:", err)
				return
			}

			pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, data)

		case "quit":
			fmt.Println("Closing Peril server")
			return
		default:
			fmt.Printf("command '%s' not recognized\n", words[0])
		}
	}
}
