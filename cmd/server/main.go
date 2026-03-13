package main

import (
	"fmt"
	"os"
	"os/signal"

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

	// data, err := json.Marshal(routing.PlayingState{
	// 	IsPaused: true,
	// })
	// if err != nil {
	// 	fmt.Println("failed to marshal data:", err)
	// 	return
	// }

	// pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, data)

	// wait for a signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Closing Peril server")
}
