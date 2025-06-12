package pubsub

import (
	"fmt"
	"encoding/json"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	json_bytes, err := json.Marshal(val);
	if (err != nil) {

		fmt.Printf("Failed to marshal json bytes %v\n", err);
		return err;
	]
}
