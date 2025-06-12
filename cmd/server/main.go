package main

import (
   "fmt" 
   "os"
   "os/signal"
   "syscall"
   amqp "github.com/rabbitmq/amqp091-go"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
   "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func main() {
	fmt.Println("Starting Peril server...")

	conn_url := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(conn_url);
	if (err != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err);
		return;
	}
	defer conn.Close();

	chann, err_ch := conn.Channel();
	if (err_ch != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err_ch);
		return;
	}

	ps := routing.PlayingState {

		IsPaused: true,
	};

	err_pub := pubsub.PublishJSON(chann, routing.ExchangePerilDirect, routing.PauseKey, ps); 
	if (err_pub != nil) {

		fmt.Printf("Error connecting to amq: %v\n", err_pub);
		return;
	}

	fmt.Printf("Connection started\n");

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT);

	sig := <-sigs;

	fmt.Printf("Received signal %v\n shutting down connections", sig);

}
