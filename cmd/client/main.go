package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	connStr := "amqp://guest:guest@localhost:5672"

	fmt.Println("Starting Peril client...")
	conn, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Println("failed to connect to rabbitmq:", err)
		return
	}
	defer conn.Close()

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("failed to get username:", err)
		return
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, user)
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.TransientQueue)
	if err != nil {
		fmt.Println("failed to declare and bind to queue:", err)
		return
	}

	state := gamelogic.NewGameState(user)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := state.CommandSpawn(words)
			if err != nil {
				fmt.Println("failed to spawn units: ", err)
				continue
			}

		case "move":
			_, err := state.CommandMove(words)
			if err != nil {
				fmt.Println("failed to move units: ", err)
				continue
			}

		case "status":
			state.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("spamming is not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("command '%s' not recognized\n", words[0])
		}
	}
}
