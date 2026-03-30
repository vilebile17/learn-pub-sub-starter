package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

type QueueType int

const (
	Transient QueueType = iota
	Durable
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, handleSubscribeJSON)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, handleSubscribeGob)
}

func handleSubscribeJSON[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	return result, nil
}

func handleSubscribeGob[T any](data []byte) (T, error) {
	var buffer bytes.Buffer
	var result T
	buffer.Write(data)

	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	if err = ch.Qos(10, 0, false); err != nil {
		return err
	}

	deliveryCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveryCh {
			result, err := unmarshaller(delivery.Body)
			if err != nil {
				fmt.Println("An error occured when unmarshalling: ", err)
				if err = delivery.Nack(false, false); err != nil {
					fmt.Println(err)
				}
				continue
			}

			ack := handler(result)
			switch ack {
			case Ack:
				if err = delivery.Ack(false); err != nil {
					fmt.Println(err)
				}
			case NackRequeue:
				if err = delivery.Nack(false, true); err != nil {
					fmt.Println(err)
				}
			case NackDiscard:
				if err = delivery.Nack(false, false); err != nil {
					fmt.Println(err)
				}
			}
		}
	}()

	return nil
}
