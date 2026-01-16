package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Starting Peril server...")
	const connStr = "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Error setting connection with RabbitMQ: %v", err)
	}
	defer connection.Close()

	amqpCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error creating amqp channel: %v", err)
	}
	

	// _, queue, err := pubsub.DeclareAndBind(
	// 		connection, 
	// 		routing.ExchangePerilTopic, 
	// 		routing.GameLogSlug,
	// 		routing.GameLogSlug + ".*",
	// 		pubsub.SimpleQueueDurable,
	// )
	// if err != nil {
	// 	log.Fatalf("could not subscribe to pause: %v", err)
	// }
	// fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	err = pubsub.SubscribeGob(
			connection, 
			routing.ExchangePerilTopic, 
			routing.GameLogSlug,
			routing.GameLogSlug + ".*",
			pubsub.SimpleQueueDurable,
			handlerLog(),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v subscribed and bound!\n", routing.GameLogSlug)

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()	
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause": 
			fmt.Println("Sending pause message...")
			err = pubsub.PublishJSON(
				amqpCh, 
				routing.ExchangePerilDirect, 
				routing.PauseKey, 
				routing.PlayingState{
					IsPaused: true,
				},
			) 
			if err != nil {
				log.Fatalf("error pubsubing in main server: %v", err)
			}
		case "resume":
			fmt.Println("Sending resume message...")
			err = pubsub.PublishJSON(
				amqpCh, 
				routing.ExchangePerilDirect, 
				routing.PauseKey, 
				routing.PlayingState{
					IsPaused: false,
				},
			) 
			if err != nil {
				log.Fatalf("error pubsubing in main server: %v", err)
			}
		case "quit": 
			log.Println("quitting server...")
			return
		default:	
			fmt.Println("invalid command")
		}
	}
}



