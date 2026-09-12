package msgqueue

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// letak di main goroutine
func PublishOrder(ctx context.Context, ch *amqp091.Channel, exchg, routingKey, orderID, userId string) error {

	err := ch.Confirm(false)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		ctx,
		exchg,
		routingKey, // connect exchange and queue
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			UserId:      userId,
			Timestamp:   time.Now(),
			Body:        []byte(orderID),
		},
	)
	if err != nil {
		return err
	}

	confirms := ch.NotifyPublish(make(chan amqp091.Confirmation, 1))

	receiveConfimation := <-confirms

	log.Printf("test jalan publish...")

	if !receiveConfimation.Ack {
		return errors.New("message broker rejected entry")
	}
	return nil

}
