package pubsub 

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"bytes"
	"encoding/json"
	"encoding/gob"
	"fmt"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota 
	SimpleQueueTransient  
)


func unmarshallerJSON[T any](body []byte) (T, error) {
	var decoded T
	err := json.Unmarshal(body, &decoded)
	return decoded, err
}

func decoderGob[T any](body []byte) (T, error) {
	buffer := bytes.NewBuffer(body)
	dec := gob.NewDecoder(buffer)
	var decoded T
	err := dec.Decode(&decoded)
	return decoded, err
}


func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	return subscribe[T]( 
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unmarshallerJSON,
	)
}



func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	return subscribe[T]( 
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		decoderGob,
		// func(data []byte) (T, error) {
		// 	buffer := bytes.NewBuffer(data)
		// 	dec := gob.NewDecoder(buffer)
		// 	var decoded T
		// 	err := dec.Decode(&decoded)
		// 	return decoded, err
		// },
	)
}


func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
	decoder func([]byte) (T, error),
) error {
	_, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)	
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	deliveryCh, err := ch.Consume(
		queueName, 
		"", 
		false, 
		false, 
		false, 
		false, 
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
		for delivery := range deliveryCh {
			decoded, err := decoder(delivery.Body)
			if err != nil {
				fmt.Printf("error unmarshalling: %v", err)	
				continue
			}
			switch handler(decoded) {
			case Ack: 
				delivery.Ack(false)
			case NackRequeue:
				delivery.Nack(false, true)
			case NackDiscard:
				delivery.Nack(false, false)
			}
		}
	}()
	return nil
}


func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, 
) (*amqp.Channel, amqp.Queue, error) {
	amqpCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}


	queue, err := amqpCh.QueueDeclare(
		queueName, 
		queueType == SimpleQueueDurable, 	// durable 
		queueType != SimpleQueueDurable,	// delete when unused
		queueType != SimpleQueueDurable,	// exclusive
		false, 					// no-wait
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},						// args
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}

	err = amqpCh.QueueBind(
		queueName, 
		key, 		// routing key 
		exchange, 	// exchange
		false, 		// no-wait
		nil, 		// args
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	
	return amqpCh, queue, nil
}
