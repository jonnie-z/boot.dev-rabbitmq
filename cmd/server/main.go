package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const CONNECTION_STRING = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, err := amqp.Dial(CONNECTION_STRING)
	if err != nil {
		log.Fatalf("unexpected err!: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connection successful!")

	rmqChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("unexpected err!: %v", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.%s", routing.GameLogSlug, "*"),
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalf("unexpected err: %v", err)
	}

	gamelogic.PrintServerHelp()

	running := true
	for running {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			fmt.Println("Sending pause message . . .")
			pubsub.PublishJson(
				rmqChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
		case "resume":
			fmt.Println("Sending resume message . . .")
			pubsub.PublishJson(
				rmqChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
		case "quit":
			fmt.Println("Exiting Program . . .")
			running = false
		default:
			fmt.Println("wtf??")
		}
	}
}
