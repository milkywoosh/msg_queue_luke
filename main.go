package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	ctx, stop := signal.NotifyContext(ctxBg, os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	chPub, err := connAmpq.Channel()
	if err != nil {
		log.Printf("connAmpq.Channel: %v", err)
		panic(err)
	}

	// note sebaiknya channel Consumer untuk setup exchange, routeKey dan queue karena akan dipakai consumer
	chCon, err := connAmpq.Channel()
	if err != nil {
		log.Printf("connAmpq.Channel: %v", err)
		panic(err)
	}

	excDirectSetup := domain.NewOrderQueueSetup("order.exchange", "direct", "order.create", "order.queue")
	err = msgqueue.SetupMQ(chCon, excDirectSetup)
	if err != nil {
		log.Printf("msgqueue.SetupMQ: %v", err)
		panic(err)
	}

	waitGroup, ctxWg := errgroup.WithContext(ctx)

	newDb := db.NewStoreMain()
	newOrder := service.NewOrderProcess(newDb)

	newServer, err := router.NewServer(cfg, chPub, newOrder)
	if err != nil {
		panic(err)
	}

	waitGroup.Go(func() error {
		return msgqueue.ConsumerOrder(ctxWg, chCon, excDirectSetup.ExchangeName, excDirectSetup.QueueName, "worker-order-1", newOrder)
	})

	waitGroup.Go(func() error {
		log.Printf("before start")
		err := newServer.Start()
		log.Printf("after start...")
		if err != nil {
			log.Printf("waitGroup: %v", err)
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done() // tunggu sinyal cancel dari goroutine lain yang error
		log.Println("shutting down http server...")
		// harus ctx bg baru
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return newServer.Shutdown(shutdownCtx) // asumsi router.Server punya method ini
	})

	err = waitGroup.Wait()
	if err != nil {
		log.Printf("waitGroup last tail: %v", err)
		panic(err)
	}

}
