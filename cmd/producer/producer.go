package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/domain"
	msgqueue "msgqueue-luke.com/v2/internals/msg_queue"
	"msgqueue-luke.com/v2/internals/utils"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// os.Args
func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hellowww"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func main() {

	//
	ctxBg := context.Background()
	ctx, stop := signal.NotifyContext(ctxBg, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cfg, err := utils.LoadConfig("../../")
	log.Println(cfg)
	if err != nil {
		log.Printf("err load config: %s", err.Error())
		return
	}

	//    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	// log.Printf("cfg rabbit mq: %s", cfg.RabbitMQPassword)
	connAmpq, err := amqp091.Dial(cfg.RabbitMQDial)
	if err != nil {
		panic(err)
	}
	defer connAmpq.Close()
	fmt.Printf("test :%s", "message queue\n")

	chPub, err := connAmpq.Channel()
	if err != nil {
		log.Printf("connAmpq.Channel: %v", err)
		panic(err)
	}

	// note: declare exchange dan queue boleh dilakukan 2x selama parameter sama semua, karena sifatnya IDEMPOTEN
	excDirectSetup := domain.NewOrderQueueSetup("order.exchange", "direct", "order.create", "order.queue")
	err = msgqueue.SetupMQ(chPub, excDirectSetup)
	if err != nil {
		log.Printf("msgqueue.SetupMQ: %v", err)
		panic(err)
	}

	// end

	body := bodyFrom(os.Args)
	err = chPub.PublishWithContext(ctx,
		excDirectSetup.ExchangeName, // exchange
		excDirectSetup.RouteKeyName, // routing key
		false,                       // mandatory
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
}
