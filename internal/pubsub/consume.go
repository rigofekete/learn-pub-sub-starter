package pubsub 

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"encoding/json"
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



func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	_, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)	
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		// log.Fatalf("channel.open: %s", err)
		return err
	}

	deliveryCh, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		defer ch.Close()
		for delivery := range deliveryCh {
			var decoded T
			err = json.Unmarshal(delivery.Body, &decoded)
			if err != nil {
				fmt.Printf("error unmarshalling: %v", err)	
				continue
			}

			ack := handler(decoded)
			if ack == Ack { 
				err = delivery.Ack(false)
				if err != nil {
					fmt.Printf("error on delivery.Ack: %v", err)	
					continue
				}
				fmt.Println("Ack")
				continue
			}
			if ack == NackRequeue {
				err = delivery.Nack(false, true)
				if err != nil {
					fmt.Printf("error on delivery.Nack (requeue): %v", err)	
					continue
				}
				fmt.Println("NackRequeue")
				continue
			}
			err = delivery.Nack(false, false)
			if err != nil {
				fmt.Printf("error on delivery.Nack (Discard): %v", err)	
				continue
			}
			fmt.Println("NackDiscard")
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



