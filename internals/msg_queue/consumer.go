package msgqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"path/filepath"

	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/mail"
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

		// msg.MessageId
		// msg.Exchange
		// msg.Priority

		log.Printf("received msg: %s", string(msg.Body))

		err := event.ProcessData(string(msg.Body), string(msg.Body))
		// if ada error then Nack
		if err != nil {
			log.Printf("failed to process message: %v | body: %s", err, string(msg.Body))

			// Nack pesan ini, jangan requeue kalau errornya pasti gagal lagi (bad payload, dll)
			if nackErr := msg.Nack(false, false); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue // lanjut ke pesan berikutnya, JANGAN return
		}

		// if ada NO ERR then Ack
		if ackErr := msg.Ack(false); ackErr != nil {
			log.Printf("failed to ack message: %v", ackErr)
			continue
		}

		log.Printf("msg processed and acked successfully")
	}

	return nil
}

func ConsumerEmailNotif(ctx context.Context, conn *amqp091.Connection, exchg, queue string, worker string, event mail.EmailSender) error {

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

		// msg.MessageId
		// msg.Exchange
		// msg.Priority

		// msg.Body as []byte() , can i convert to struct of mail.EmailPayload

		var payload mail.EmailPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			log.Printf("failed to unmarshal email payload: %v", err)
			msg.Nack(false, false) // reject, no requeue if payload rusak
			continue
		}

		err := event.SendEmail(payload)
		// if ada error then Nack
		if err != nil {
			log.Printf("failed to process message: %v | body: %s", err, payload)

			// Nack pesan ini, jangan requeue kalau errornya pasti gagal lagi (bad payload, dll)
			if nackErr := msg.Nack(false, false); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue // lanjut ke pesan berikutnya, JANGAN return
		}

		// if ada NO ERR then Ack
		if ackErr := msg.Ack(false); ackErr != nil {
			log.Printf("failed to ack message: %v", ackErr)
			continue
		}

		log.Printf("msg processed and acked successfully")
	}

	return nil
}
