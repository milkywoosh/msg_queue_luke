package msgqueue

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/domain"
)

func SetupMQ(conn *amqp091.Connection, declare domain.OrderQueueSetup) error {

	// setup exchange
	// setup queue
	// setup binding queue

	// note: cukup pake 1 channel aja untuk declare, lifecycle channel ends up after semua declaration is done
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		declare.ExchangeName,
		declare.TypeExchange,
		true,
		false,
		false,
		false,
		nil, // args ampq.Table??
	)
	if err != nil {
		log.Println("err exchange declare: ", err)
		return err
	}

	// Queue
	// Passive
	// Durable
	// AutoDelete
	// Exclusive
	// NoWait
	// Arguments
	queueArgs := amqp091.Table{
		amqp091.QueueTypeArg: amqp091.QueueTypeQuorum, // Resolves to "x-queue-type": "quorum"

		// Optional: Add poison pill protection (highly recommended for quorum queues)
		"x-delivery-limit": int32(5),
	}

	_, err = ch.QueueDeclare(
		declare.QueueName,
		true,
		false,
		false,
		false,
		queueArgs,
	)
	if err != nil {
		log.Println("err queue declare: ", err)
		return err
	}

	err = ch.QueueBind(declare.QueueName, declare.RouteKeyName, declare.ExchangeName, false, nil)
	if err != nil {
		log.Println("err queue binding: ", err)
		return err
	}
	// kapan dipake
	// ch.ExchangeUnbind()

	// kapan dipake
	// ch.ExchangeDelete()

	return nil
}
