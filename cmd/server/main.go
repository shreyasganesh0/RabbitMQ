package main

import (
   "fmt" 
   "os"
   "os/signal"
   "syscall"
   amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	conn_url := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(conn_url);
	if (err != nil) {

		fmt.Printf("Error connecting to amq: %v\n");
		return;
	}
	defer conn.Close();

	chann, err_ch := conn.Channel();
	if (err_ch != nil) {

		fmt.Printf("Error connecting to amq: %v\n");
		return;
	}

	fmt.Printf("Connection started\n");

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT);

	sig := <-sigs;

	fmt.Printf("Received signal %v\n shutting down connections", sig);

}
