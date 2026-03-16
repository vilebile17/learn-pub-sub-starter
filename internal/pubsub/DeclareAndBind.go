package pubsub

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key, queueType string) (*amqp.Channel, amqp.Queue, error) {
	if queueType != "durable" && queueType != "transient" {
		return nil, amqp.Queue{}, errors.New("queueType must be either 'durable' or 'transient'")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	q, err := ch.QueueDeclare(queueName, queueType == "durable", queueType == "transient", queueType == "transient", false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err = ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
