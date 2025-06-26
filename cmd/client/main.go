package main

import ( 
   "fmt" 
   "os"
   "time"
   "os/signal"
   "syscall"
   "strconv"
   amqp "github.com/rabbitmq/amqp091-go"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

type MalLog struct {

	Log string
}

func handlerWar(gs *gamelogic.GameState, chann *amqp.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {

		username := gs.GetUsername()
		game_key := routing.GameLogSlug + "." + username 

		fmt.Printf("My user is %s and warring user is %s and %s\n", username, dw.Attacker.Username, dw.Defender.Username)

		defer fmt.Print("> ")

		warOutcome, winner, loser := gs.HandleWar(dw)
		game_log := routing.GameLog {
						CurrentTime: time.Now(),
						Message: fmt.Sprintf("%s won a war against %s/n", winner, loser),  
						Username: username, 
					}
		switch warOutcome {

		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:

			err_pub := pubsub.PublishGob(chann, 
										routing.ExchangePerilTopic, 
										game_key, 
										game_log); 
			if (err_pub != nil) {

				fmt.Printf("Error connecting to amq: %v\n", err_pub);
				return pubsub.NackRequeue;
			}

			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:

							
			err_pub := pubsub.PublishGob(chann, 
										routing.ExchangePerilTopic, 
										game_key, 
										game_log); 
			if (err_pub != nil) {

				fmt.Printf("Error connecting to amq: %v\n", err_pub);
				return pubsub.NackRequeue;
			}

			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:

			game_log.Message = fmt.Sprintf("A war between %s and %s resulted in a draw\n", 
											winner, loser);

			err_pub := pubsub.PublishGob(chann, 
										routing.ExchangePerilTopic, 
										game_key, 
										game_log); 
			if (err_pub != nil) {

				fmt.Printf("Error connecting to amq: %v\n", err_pub);
				return pubsub.NackRequeue;
			}
			return pubsub.Ack
		}

		fmt.Println("error: unknown war outcome")
		return pubsub.NackDiscard
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {

	return func(ps routing.PlayingState) pubsub.AckType {

		defer fmt.Print(">");
		
		gs.HandlePause(ps);

		return pubsub.Ack;
	}

}

func handlerMove(gs *gamelogic.GameState, chann *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {

	return func(move gamelogic.ArmyMove) pubsub.AckType {

		defer fmt.Print(">");
		
		move_outcome := gs.HandleMove(move);

		switch move_outcome {

			case gamelogic.MoveOutcomeSamePlayer:

				return pubsub.NackDiscard;
			case gamelogic.MoveOutComeSafe:

				return pubsub.Ack;
			case gamelogic.MoveOutcomeMakeWar:

					err_pub := pubsub.PublishJSON(
										chann,	
										routing.ExchangePerilTopic,
										routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
										gamelogic.RecognitionOfWar{
											Attacker: move.Player,
											Defender: gs.GetPlayerSnap(),
										},
									)
					if (err_pub != nil) {

						fmt.Printf("Error connecting to amq: %v\n", err_pub);
						return pubsub.NackRequeue;
					}

				return pubsub.Ack;

			default:

				return pubsub.NackDiscard;
		}
			return pubsub.NackDiscard;

	}

}

func main() {
	fmt.Println("Starting Peril client...")
	conn_url := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(conn_url);
	if (err != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err);
		return;
	}
	defer conn.Close();

	chann, err_chan := conn.Channel();
	if (err_chan != nil) {

		return;
	}

	username, err_r := gamelogic.ClientWelcome(); 
	if (err_r != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err_r);
		return;
	}

	fmt.Printf("Connection started\n");

	game_state := gamelogic.NewGameState(username);

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+game_state.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(game_state, chann),
	)
	if err != nil {
		fmt.Printf("could not subscribe to army moves: %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(game_state, chann),
	)
	if err != nil {
		fmt.Printf("could not subscribe to war declarations: %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+game_state.GetUsername(),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(game_state),
	)
	if err != nil {
		fmt.Printf("could not subscribe to pause: %v", err)
	}

	go func() {

		for  {
			words := gamelogic.GetInput();

			switch (words[0]) {

				case "spawn":

					err := game_state.CommandSpawn(words);
					if (err != nil) {

						fmt.Printf("Spawned unit failed\n");
					}
						fmt.Printf("Spawned unit success\n");


				case "move":

					move, err := game_state.CommandMove(words);
					move_key := routing.ArmyMovesPrefix + "." + username;

					err_pub := pubsub.PublishJSON(chann, routing.ExchangePerilTopic, move_key,  move); 
					if (err_pub != nil) {

						fmt.Printf("Error connecting to amq: %v\n", err_pub);
						return;
					}

					fmt.Printf("Published to move to topic %s", routing.ExchangePerilTopic);

					if (err != nil) {

						fmt.Printf("Moved unit failed\n");
					}
						fmt.Printf("Moved unit success\n");

				case "status":

					game_state.CommandStatus();

				case "help":

					gamelogic.PrintClientHelp();

				case "spam":

					spam_count, err := strconv.Atoi(words[1])
					if (err != nil) {

						fmt.Printf("Failed to write spam due to %z", err);
						continue;
					}

					for i := 0; i < spam_count; i++ {

					game_log := routing.GameLog {
									CurrentTime: time.Now(),
									Message: gamelogic.GetMaliciousLog(), 
									Username: username, 
								}
						err_pub := pubsub.PublishGob(
											chann,	
											routing.ExchangePerilTopic,
											routing.GameLogSlug+"."+ username,
											game_log,
										)
						if (err_pub != nil) {

							fmt.Printf("Error connecting to amq: %v\n", err_pub);
						}
					}

					return; 

				case "quit":

					gamelogic.PrintQuit();
					return;

				default:

					fmt.Printf("Invalid command\n");
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT);

	sig := <-sigs;

	fmt.Printf("Received signal %v\n shutting down connections", sig);
}
