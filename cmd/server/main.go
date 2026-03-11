package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
)

func main() {
	connStr := "amqp://guest:guest@localhost:5672"

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Println("Starting Peril server...")

	// wait for a signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Closing Peril server")
}
