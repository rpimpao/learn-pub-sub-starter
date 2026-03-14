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

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	// create (if the queue doest exist yet) and bind to the queue
	chann, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		fmt.Println("failed to declare/bind to queue: ", err)
		return err
	}

	// consume from the channel
	deliveryChannel, err := chann.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println("failed to consume messages from channel")
		return err
	}

	unmarshaller := func(body []byte) (T, error) {
		var data T
		err := json.Unmarshal(body, &data)
		return data, err
	}

	go func() {
		defer chann.Close()
		// process messages in the channel
		for msg := range deliveryChannel {
			data, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Println("failed to unmarshal message: ", err)
				continue
			}

			// callback with unmarshaled data
			handler(data)

			// ack message to delete it from the queue
			msg.Ack(false)
		}
	}()

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
