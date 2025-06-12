package pubsub

import (
	"fmt"
	"encoding/json"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	json_bytes, err := json.Marshal(val);
	if (err != nil) {

		fmt.Printf("Failed to marshal json bytes %v\n", err);
		return err;
	}

	pubbing := amqp.Publishing {
		ContentType: "application/json",
		Body: json_bytes,
	};

	err_pub := ch.PublishWithContext(context.Background(), exchange, key, false, false, pubbing);
	if (err_pub != nil) {

		return err_pub;
	}

	return nil;

}
