package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("unexpected err: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	pubsub.SubscribeJson(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJson(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, gameState.GetUsername()),
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gameState, ch),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	err = pubsub.SubscribeJson(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gameState, ch),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

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
			armyMove, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Printf("Unsuccessful move!: %v\n", err)
			}

			pubsub.PublishJson(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, gameState.GetUsername()),
				armyMove,
			)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(input) < 2 {
				fmt.Printf("USAGE: spawn {int}")
			}

			amt, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Printf("error converting amount: %v", err)
			}

			for range amt {
				msg := gamelogic.GetMaliciousLog()

				pubsub.PublishGob(
					ch,
					routing.ExchangePerilTopic,
					fmt.Sprintf("%s.%s", routing.GameLogSlug, gameState.GetUsername()),
					routing.GameLog{
						CurrentTime: time.Now(),
						Message: msg,
						Username: gameState.GetUsername(),
					},
				)
			}
		case "quit":
			gamelogic.PrintQuit()
			running = false
		default:
			fmt.Println("wtf??")
		}
	}
}
