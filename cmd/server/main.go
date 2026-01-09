package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github/rigofekete/learn-pub-sub-starter/internal/pubsub"
	"github/rigofekete/learn-pub-sub-starter/internal/routing"
	"fmt"
	"os"
	"os/signal"
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
	
	val := routing.PlayingState {
		IsPaused:	true, 
	}

	err := pubsub.PublishJSON(amqpCh, routing.ExchangePerilDirect, routing.PauseKey, val) 
	if err != nil {
		logFatalf("error publishing in main server: %v", err)
	}


	fmt.Println("Connection successful!")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan 

	fmt.Println("Shutting down program...")
}	
