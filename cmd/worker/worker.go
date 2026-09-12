package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/domain"
	msgqueue "msgqueue-luke.com/v2/internals/msg_queue"
	"msgqueue-luke.com/v2/internals/utils"
)

// consumer as worker

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	fmt.Println("worrker")

	ctxBg := context.Background()
	ctx, stop := signal.NotifyContext(ctxBg, os.Interrupt, syscall.SIGTERM)
	defer stop()

	duration := 1 * time.Minute

	// gimana propagate value in context? bring unique identifier
	ctxTimeout, cancel := context.WithTimeout(ctxBg, 1*time.Minute)
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

	chCons, err := connAmpq.Channel()
	if err != nil {
		log.Printf("connAmpq.Channel: %v", err)
		panic(err)
	}

	// note: declare exchange dan queue boleh dilakukan 2x selama parameter sama semua, karena sifatnya IDEMPOTEN
	excDirectSetup := domain.NewOrderQueueSetup("order.exchange", "direct", "order.create", "order.queue")
	err = msgqueue.SetupMQ(chCons, excDirectSetup)
	if err != nil {
		log.Printf("msgqueue.SetupMQ: %v", err)
		panic(err)
	}

	msgs, err := chCons.Consume(
		excDirectSetup.QueueName, // queue
		"worker-newtask-1",       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("Done")
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	select {
	case <-ctx.Done():
		fmt.Println("context canceled")
		return
	case <-ctxTimeout.Done():
		fmt.Printf("context timeout %s", duration)
		return
	}

}
