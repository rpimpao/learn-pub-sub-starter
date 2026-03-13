package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	DurableQueue SimpleQueueType = iota
	TransientQueue
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // SimpleQueueType is an "enum" type I made to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	chann, err := conn.Channel()
	if err != nil {
		fmt.Println("failed to open a channel:", err)
		return nil, amqp.Queue{}, err
	}
	defer chann.Close()

	durable := queueType == DurableQueue
	queue, err := chann.QueueDeclare(queueName, durable, !durable, !durable, false, nil)
	if err != nil {
		fmt.Println("failed to declare a queue:", err)
		return nil, amqp.Queue{}, err
	}

	if err = chann.QueueBind(queueName, key, exchange, false, nil); err != nil {
		fmt.Println("failed to bind queue to exchange")
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue to exchange")
	}

	return chann, queue, nil
}
