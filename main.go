package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"

	"msgqueue-luke.com/v2/internals/db"
	"msgqueue-luke.com/v2/internals/domain"
	msgqueue "msgqueue-luke.com/v2/internals/msg_queue"
	"msgqueue-luke.com/v2/internals/router"
	"msgqueue-luke.com/v2/internals/service"
	"msgqueue-luke.com/v2/internals/utils"
)

// prefered at main goroutine

func main() {

	ctxBg := context.Background()

	cfg, err := utils.LoadConfig("./")
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

	chAmpq, err := connAmpq.Channel()
	if err != nil {
		log.Printf("connAmpq.Channel: %v", err)
		panic(err)
	}

	excDirectSetup := domain.NewOrderQueueSetup("order.exchange", "direct", "order.create", "order.queue")
	err = msgqueue.SetupMQ(chAmpq, excDirectSetup)
	if err != nil {
		log.Printf("msgqueue.SetupMQ: %v", err)
		panic(err)
	}

	errWaitgroup, ctx := errgroup.WithContext(ctxBg)
	if err != nil {
		log.Printf("errWaitGroup: %v", err)
		panic(err)
	}

	newDb := db.NewStoreMain()
	newOrder := service.NewOrderProcess(newDb)

	newServer, err := router.NewServer(cfg, chAmpq, newOrder)
	if err != nil {
		panic(err)
	}

	go msgqueue.ConsumerOrder(ctx, chAmpq, excDirectSetup.ExchangeName, excDirectSetup.QueueName, "worker-order-1", newOrder)

	errWaitgroup.Go(func() error {
		log.Printf("before start")
		err = newServer.Start()
		log.Printf("after start...")
		if err != nil {
			log.Printf("errWaitGroup: %v", err)
			return err
		}
		return nil
	})

	err = errWaitgroup.Wait()
	if err != nil {
		log.Printf("errWaitGroup last tail: %v", err)
		panic(err)
	}

}
