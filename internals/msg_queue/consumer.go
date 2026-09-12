package msgqueue

import (
	"context"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type MessageProcessor interface {
	ProcessData(key string) error
}

// jgn return error check via log aja dulu
func ConsumerOrder(ctx context.Context, ch *amqp091.Channel, exchg, queue string, worker string, event MessageProcessor) error {

	chConsumer, err := ch.ConsumeWithContext(
		ctx,
		queue,
		worker, // worker as consumer, yang akan menerima dari QUEUE, misal [order-worker-1, order-worker-2, order-worker-3]
		false,  // autoAck, after consume dianggap delivered aja pkoknya. Bisa ilang klo server crash
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("err chConsumer: %v", err)
		return err
	}

	for msg := range chConsumer {
		log.Printf("received msg: %s", string(msg.Body))

		err := event.ProcessData(string(msg.Body))
		if err != nil {
			log.Printf("failed to process message: %v | body: %s", err, string(msg.Body))

			// Nack pesan ini, jangan requeue kalau errornya pasti gagal lagi (bad payload, dll)
			if nackErr := msg.Nack(false, false); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue // lanjut ke pesan berikutnya, JANGAN return
		}

		if ackErr := msg.Ack(false); ackErr != nil {
			log.Printf("failed to ack message: %v", ackErr)
			continue
		}

		log.Printf("msg processed and acked successfully")
	}

	return nil
}
