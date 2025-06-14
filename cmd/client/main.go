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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT);

	sig := <-sigs;

	fmt.Printf("Received signal %v\n shutting down connections", sig);
}
