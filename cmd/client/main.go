package main

import (
	"fmt"
	"os"
	"os/signal"

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Closing Peril client")
}
