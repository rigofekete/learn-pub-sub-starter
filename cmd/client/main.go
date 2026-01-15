package main

import (
	"fmt"
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
	
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	fmt.Println("Starting Peril client...")
	const connStr = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Error setting connection with RabbitMQ: %v", err)
	}
	defer connection.Close()

	userName, err := gamelogic.ClientWelcome()  
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}

	// _, _, err = pubsub.DeclareAndBind(
	// 		connection, 
	// 		routing.ExchangePerilTopic, 
	// 		routing.ArmyMovesPrefix + "." + userName,
	// 		routing.ArmyMovesPrefix + ".*",
	// 		pubsub.SimpleQueueTransient,
	// )

	gameState := gamelogic.NewGameState(userName)
	
	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection, 
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix + "." + gameState.GetUsername(),
		routing.ArmyMovesPrefix + ".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, publishCh),
	)
	if err != nil {
		log.Fatalf("Error subscribing move JSON: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix + ".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, publishCh),
	)
	if err != nil {
		log.Fatalf("Error subscribing war JSON: %v", err)
	}
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)	
				continue
			}
		case "move":
			mv, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)	
				continue
			}

			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix + "." + gameState.Player.Username,
				mv,
			)
			if err != nil {
				fmt.Printf("error moving unit in client: %s\n", err)	
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp() 
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("command not allowed")
		}
	}

	// fmt.Println("Closing Peril client...")
	//
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan

}


