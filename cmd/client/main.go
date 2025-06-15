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
	queue_name := routing.PauseKey + "." + username;

	_, _, err_dec := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queue_name, routing.PauseKey, 1); 
	if (err_dec != nil) {

		fmt.Printf("Error with Declare and bind queue %v\n", err_dec);
		return;
	}

	fmt.Printf("Connection started\n");

	game_state := gamelogic.NewGameState(username);
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
