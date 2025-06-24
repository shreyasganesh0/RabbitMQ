package main

import ( 
   "fmt" 
   "os"
   "os/signal"
   "syscall"
   amqp "github.com/rabbitmq/amqp091-go"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func handlerWar(gs *gamelogic.GameState) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, _, _ := gs.HandleWar(dw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
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

					move_key := routing.WarRecognitionsPrefix + "." + gs.GetUsername();
					err_pub := pubsub.PublishJSON(chann, routing.ExchangePerilTopic, move_key,  move); 
					if (err_pub != nil) {

						fmt.Printf("Error connecting to amq: %v\n", err_pub);
					}

				return pubsub.NackRequeue;
			case gamelogic.MoveOutcomeMakeWar:

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
		handlerWar(game_state),
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

					fmt.Printf("Spamming not allowed yet\n");

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
