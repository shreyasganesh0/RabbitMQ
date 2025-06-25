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

func handlerLog() func(routing.GameLog) pubsub.AckType {

	return func(gl routing.GameLog) pubsub.AckType {

		defer fmt.Print(">");
		
		err := gamelogic.WriteLog(gl);
		if (err != nil) {

			return pubsub.NackRequeue;
		}

		return pubsub.Ack;
	}

}

func main() {
	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp();

	conn_url := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(conn_url);
	if (err != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err);
		return;
	}
	fmt.Printf("Connection started\n");
	defer conn.Close();

	queue_name := routing.GameLogSlug; 
	key := routing.GameLogSlug + ".*";
    exchange_name := routing.ExchangePerilTopic;
	chann, _, err_dec := pubsub.DeclareAndBind(conn, exchange_name, queue_name, key, 0); 
	if (err_dec != nil) {

		fmt.Printf("Error with Declare and bind queue %v\n", err_dec);
		return;
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT);
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"*",	
		pubsub.Durable,
		handlerLog(),
	)
	if err != nil {
		fmt.Printf("could not subscribe to army moves: %v", err)
	}

	go func() {
		for {

			words := gamelogic.GetInput();

			switch words[0] {

				case "pause":

					ps := routing.PlayingState {

						IsPaused: true,
					};

					fmt.Printf("In paused state\n"); 
					err_pub := pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, ps); 
				if (err_pub != nil) {

					fmt.Printf("Error connecting to amq: %v\n", err_pub);
					return;
				}
				case "resume":

					ps := routing.PlayingState {

						IsPaused: false,
					};
					fmt.Printf("In resumed state\n"); 
					err_pub := pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, ps); 
				if (err_pub != nil) {

					fmt.Printf("Error connecting to amq: %v\n", err_pub);
					return;
				}

				case "quit":

					fmt.Printf("quitting server\n"); 
					return;

				default:

					fmt.Printf("Invalid word %s\n", words[0]); 
				}
			}
		}()

	sig := <-sigs;

	fmt.Printf("Received signal %v\n shutting down connections", sig);

}
