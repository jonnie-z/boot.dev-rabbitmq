package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJson[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes := &bytes.Buffer{}
	encoder := json.NewEncoder(jsonBytes)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}

	ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes.Bytes(),
		},
	)

	return nil
}

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	rmqChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	q, err := rmqChan.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType != Durable,
		queueType != Durable,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = rmqChan.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return rmqChan, q, nil
}

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
    handler func(T),
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	deliveryCh, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	go func() {
		defer ch.Close()
		for delivery := range deliveryCh {
			var b T
			decoder := json.NewDecoder(bytes.NewBuffer(delivery.Body))
			decoder.Decode(b)
			handler(b)

			delivery.Ack(false)
		}
	}()

	return nil
}
