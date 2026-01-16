package main 

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)



func handlerLog() func(gamelog routing.GameLog) pubsub.Acktype {
	return func(gamelog routing.GameLog) pubsub.Acktype {
		defer fmt.Println("> ")
		err := gamelogic.WriteLog(gamelog)
		if err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
