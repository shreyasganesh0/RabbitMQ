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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {

	return func(ps routing.PlayingState) {

		defer fmt.Print(">");
		
		gs.HandlePause(ps);
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

	username, err_r := gamelogic.ClientWelcome(); 
	if (err_r != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err_r);
		return;
	}

	fmt.Printf("Connection started\n");

	queue_name := routing.PauseKey + "." + username;

	game_state := gamelogic.NewGameState(username);
	handler_func := handlerPause(game_state);

	err_sub := pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queue_name, routing.PauseKey, pubsub.Transient, handler_func);
	if (err_sub != nil) {

		fmt.Printf("Error subscribing to queue %v\n", err_sub);
		return;
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

					_, err := game_state.CommandMove(words);
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
