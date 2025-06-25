package pubsub

import (
	"fmt"
	"encoding/json"
	"encoding/gob"
	"bytes"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type AckType int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)


func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {

	var buf bytes.Buffer;
	enc := gob.NewEncoder(&buf);

	err := enc.Encode(val);
	if (err != nil) {

		fmt.Printf("Failed to marshal gob bytes %v\n", err);
		return err;
	}

	pubbing := amqp.Publishing {
		ContentType: "application/gob",
		Body: buf.Bytes(),
	};

	err_pub := ch.PublishWithContext(context.Background(), exchange, key, false, false, pubbing);
	if (err_pub != nil) {

		return err_pub;
	}

	return nil;

}

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
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	var queue amqp.Queue
	var err_q error

	chann, err_chan := conn.Channel();
	if (err_chan != nil) {

		return nil, queue, err_chan;
	}

	var durable, auto_delete, exclusive, no_wait bool;

	if (simpleQueueType == Durable) {

		durable = true;
	} else if (simpleQueueType == Transient) {

		auto_delete = true;
		exclusive = true;
	}

	table := make(amqp.Table);
	table["x-dead-letter-exchange"] = "peril_dlx";

	queue, err_q = chann.QueueDeclare(queueName, durable, auto_delete, exclusive, no_wait, table); 
	if (err_q != nil) {

		return chann, queue, err_q;
	}

	err_q = chann.QueueBind(queueName, key, exchange, no_wait, nil)

	return chann, queue, err_q;

}

func SubscribeGob[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
    handler func(T) AckType,
) error {

	chann, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if (err != nil) {

		return err;
	}

	del_chan, err_cons := chann.Consume(queueName, "", false, false, false, false, nil);
	if (err_cons != nil) {

		return err_cons;
	}

	go func() {

		for message := range del_chan {

			var handler_param T
			buf := bytes.NewBuffer(message.Body);

			dec := gob.NewDecoder(buf);
			err_gob := dec.Decode(&handler_param);
			if (err_gob != nil) {

				fmt.Printf("Error while parsing data to gob %z\n", err_gob);
				continue;
			}

			ack := handler(handler_param);

			switch ack {

				case Ack:

					fmt.Printf("Ack type\n");
					err_ack := message.Ack(false);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
				case NackRequeue:

					fmt.Printf("Nack requeue type\n");
					err_ack := message.Nack(false, true);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
				case NackDiscard:

					fmt.Printf("Nack discard type\n");
					err_ack := message.Nack(false, false);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
				default:

					fmt.Printf("Invalid type\n");
					err_ack := message.Nack(false, false);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
			}
		}
	}()

	return nil;
}

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
    handler func(T) AckType,
) error {

	chann, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if (err != nil) {

		return err;
	}

	del_chan, err_cons := chann.Consume(queueName, "", false, false, false, false, nil);
	if (err_cons != nil) {

		return err_cons;
	}

	go func() {

		for message := range del_chan {

			var handler_param T
			err_json := json.Unmarshal(message.Body, &handler_param);
			if (err_json != nil) {

				fmt.Printf("Error while parsing data to json %z\n", err_json);
				continue;
			}

			ack := handler(handler_param);

			switch ack {

				case Ack:

					fmt.Printf("Ack type\n");
					err_ack := message.Ack(false);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
				case NackRequeue:

					fmt.Printf("Nack requeue type\n");
					err_ack := message.Nack(false, true);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
				case NackDiscard:

					fmt.Printf("Nack discard type\n");
					err_ack := message.Nack(false, false);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
				default:

					fmt.Printf("Invalid type\n");
					err_ack := message.Nack(false, false);
					if (err_ack != nil) {

						fmt.Printf("Failed to remove message %z\n", err_ack);
						continue;
					}
			}
		}
	}()

	return nil;
}
