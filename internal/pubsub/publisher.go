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

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	var queue amqp.Queue
	var err error

	chann, err_chan := conn.Channel();
	if (err_chan != nil) {

		return nil, queue, err_chan;
	}

	var durable, auto_delete, exclusive, no_wait bool;

	if (simpleQueueType == 0) {

		durable = true;
	} else if (simpleQueueType == 1 {

		auto_delete = true;
		exclusive = false;
	}

	queue, err_q = chann.QueueDeclare(queueName, durable, auto_delete, exclusive, no_wait, nil)
	if (err_q != nil) {

		return chann, queue, err_q;
	}

	err_q = chann.QueueBind(queueName, key, exchange, no_wait, nil)

	return chann, queue, err_q;

}
