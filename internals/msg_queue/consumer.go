package msgqueue

import (
	"context"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type EventEnqueue interface {
	ProcessData(key string) error
}

// jgn return error check via log aja dulu
func ConsumerOrder(ctx context.Context, ch *amqp091.Channel, exchg, queue string, worker string, event EventEnqueue) {

	chConsumer, err := ch.ConsumeWithContext(
		ctx,
		queue,
		worker, // worker as consumer, yang akan menerima dari QUEUE, misal [order-worker-1, order-worker-2, order-worker-3]
		true,   // autoAck, after consume dianggap delivered aja pkoknya. Bisa ilang klo server crash
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("err chConsumer: %v", err)
	}

	for msg := range chConsumer {

		err := event.ProcessData(string(msg.Body))
		if err != nil {
			log.Printf("err after ProcessData : %v", err)
			// what's best scenario
		}

		msg.Ack(false) // jgn di acknowledge dulu karena nanti langsung hilang, dan menganggap event data sudah diproses

	}
}
