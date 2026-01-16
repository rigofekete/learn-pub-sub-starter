package main 

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)



func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Println("> ")
		mvOutcome:= gs.HandleMove(move)
		switch mvOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix + "." + gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error publishing war declaration: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")

		return pubsub.NackDiscard
	}
}


func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Println("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}


func handlerLog() func(routing.GameLog) pubsub.Acktype {
	return func(gl routing.GameLog) pubsub.Acktype {
		defer fmt.Println("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Println("> ")
		 outcome, winner, loser := gs.HandleWar(rw)
		 switch outcome {
		 case gamelogic.WarOutcomeNotInvolved:
			 return pubsub.NackRequeue
		 case gamelogic.WarOutcomeNoUnits:
			 return pubsub.NackDiscard
		 case gamelogic.WarOutcomeOpponentWon:
			 err := publishLog(
				 gs.GetUsername(), 
				 publishCh, 
				 fmt.Sprintf("%s won a war against %s", winner, loser),
			 )
			 if err != nil {
				return pubsub.NackRequeue
			 }
			 return pubsub.Ack
		 case gamelogic.WarOutcomeYouWon:
			 err := publishLog(
				 gs.GetUsername(), 
				 publishCh, 
				 fmt.Sprintf("%s won a war against %s", winner, loser),
			 )
			 if err != nil {
				return pubsub.NackRequeue
			 }
			 return pubsub.Ack
		 case gamelogic.WarOutcomeDraw:
			 err := publishLog(
				 gs.GetUsername(), 
				 publishCh, 
				 fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
			 )
			 if err != nil {
				return pubsub.NackRequeue
			 }
			 return pubsub.Ack
		 default:
			 fmt.Println("Invalid outcome")
			 return pubsub.NackDiscard
		 }
	}
}



func publishLog(username string, publishCh *amqp.Channel, message string) error {
	gl := routing.GameLog{
		CurrentTime: 	time.Now(),
		Message:    	message,
		Username:   	username,
	}
	return pubsub.PublishGob(
		publishCh, 
		routing.ExchangePerilTopic,
		routing.GameLogSlug + "." + username, 
		gl,
	)
 }
