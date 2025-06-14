package main

import "fmt"

func main() {
	fmt.Println("Starting Peril client...")
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

	fmt.Printf("Connection started\n");

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT);

	sig := <-sigs;

	fmt.Printf("Received signal %v\n shutting down connections", sig);
}
