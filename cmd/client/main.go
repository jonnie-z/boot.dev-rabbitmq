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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("unexpected err: %v", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("unexpected err: %v", err)
	}

	gameState := gamelogic.NewGameState(username)
	fmt.Printf("gs in main: %v\n", &gameState)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)

	running := true
	for running {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				fmt.Printf("Unsuccessful spawn!: %v\n", err)
			}
		case "move":
			_, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Printf("Unsuccessful move!: %v\n", err)
			} else {
				fmt.Println("Move successful!")
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			running = false
		default:
			fmt.Println("wtf??")
		}
	}
}
