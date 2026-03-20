package main

import (
	"fmt"
	"strconv"
	"time"

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

	chann, err := conn.Channel()
	if err != nil {
		fmt.Println("failed to create channel:", err)
		return
	}

	state := gamelogic.NewGameState(user)

	// subscribe to pause queue
	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, user)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.TransientQueue, handlerPause(state))
	if err != nil {
		fmt.Println("failed to subscribe to queue:", err)
		return
	}

	// subscribe to the move queue
	queueName = fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, user)
	routingKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, queueName, routingKey, pubsub.TransientQueue, handlerMove(state, chann))
	if err != nil {
		fmt.Println("failed to subscribe to queue:", err)
		return
	}

	// subscribe to war queue
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.DurableQueue, handlerRecognitionOfWar(state, chann))
	if err != nil {
		fmt.Println("failed to subscribe to queue:", err)
		return
	}

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
			move, err := state.CommandMove(words)
			if err != nil {
				fmt.Println("failed to move units: ", err)
				continue
			}

			err = pubsub.PublishJSON(chann, routing.ExchangePerilTopic, routingKey, move)
			if err != nil {
				fmt.Println("failed to publish move message:", err)
				return
			}
			fmt.Println("Move published ok!")

		case "status":
			state.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) != 2 {
				fmt.Println("wrong number of arguments")
				continue
			}

			n, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("invalid number: ", words[1])
				continue
			}

			for range n {
				log := gamelogic.GetMaliciousLog()
				_ = publishGameLog(chann, user, log) // ignore error, we be spamming
			}

		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("command '%s' not recognized\n", words[0])
		}
	}
}

func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
