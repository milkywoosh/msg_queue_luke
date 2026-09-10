package msgqueue

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/domain"
)

func SetupMQ(ch *amqp091.Channel, declare domain.OrderQueueSetup) error {

	// setup exchange
	// setup queue
	// setup binding queue

	err := ch.ExchangeDeclare(
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
	_, err = ch.QueueDeclare(
		declare.QueueName,
		true,
		false,
		false,
		false,
		nil,
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
