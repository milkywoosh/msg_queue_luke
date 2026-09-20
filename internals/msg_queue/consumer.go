package msgqueue

import (
	"context"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type MessageProcessor interface {
	ProcessData(string, string) error
}

// jgn return error check via log aja dulu
func ConsumerOrder(ctx context.Context, conn *amqp091.Connection, exchg, queue string, worker string, event MessageProcessor) error {

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	delivConsumer, err := ch.ConsumeWithContext(
		ctx,
		queue,
		worker, // worker as consumer, yang akan menerima dari QUEUE, misal [order-worker-1, order-worker-2, order-worker-3]
		false,  // autoAck=true, after consume dianggap delivered aja pkoknya. Bisa ilang klo server crash
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("err chConsumer: %v", err)
		return err
	}

	for msg := range delivConsumer {
		log.Printf("received msg: %s", string(msg.Body))

		err := event.ProcessData(string(msg.Body), string(msg.Body))
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
