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
	"msgqueue-luke.com/v2/internals/storage"
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

	// pubChan, err := msgqueue.NewPubChan(connAmpq)

	publisherPool, err := msgqueue.NewPublisherPool(connAmpq, 3)
	if err != nil {
		log.Fatalf("err NewPubChan: %s", err.Error())
	}

	// note sebaiknya channel Consumer di-define di scope function each consumer agar create CHAN berbeda dari 1 connection awal
	//  untuk setup exchange, routeKey dan queue karena akan dipakai consumer

	excDirectSetupOrderStore := domain.NewQueueSetup("order.exchange", "direct", "order.create", "order.store")
	err = msgqueue.SetupMQ(connAmpq, excDirectSetupOrderStore)
	if err != nil {
		log.Printf("msgqueue.SetupMQ: %v", err)
		panic(err)
	}

	excDirectSetupNotifEmail := domain.NewQueueSetup("order.exchange", "direct", "order.notif.email", "email")
	err = msgqueue.SetupMQ(connAmpq, excDirectSetupNotifEmail)
	if err != nil {
		log.Printf("msgqueue.SetupMQ: %v", err)
		panic(err)
	}

	waitGroup, ctxWg := errgroup.WithContext(ctx)

	newDb := db.NewStoreMain()
	newNotifEmail := utils.NewNotifEmail() // create pointer
	newOrder := service.NewOrderProcess(newDb)

	newClientS3, err := storage.NewClientObjectS3(ctx, cfg.AccessKeyS3, cfg.SecretKeyS3, cfg.AddressS3)
	if err != nil {
		log.Fatal(err)
	}

	objectStorage := storage.NewSeaweedS3(newClientS3)

	const bucket = "scmt"
	if err := objectStorage.EnsureBucket(ctx, bucket); err != nil {
		log.Fatal(err)
	}

	newServer, err := router.NewServer(cfg, publisherPool, newOrder, objectStorage)
	if err != nil {
		panic(err)
	}

	waitGroup.Go(func() error {
		return msgqueue.ConsumerOrder(ctxWg, connAmpq, excDirectSetupOrderStore.ExchangeName, excDirectSetupOrderStore.QueueName, "worker-order-1", newOrder)
	})

	waitGroup.Go(func() error {
		return msgqueue.ConsumerOrder(ctxWg, connAmpq, excDirectSetupNotifEmail.ExchangeName, excDirectSetupNotifEmail.QueueName, "worker-order-2", newNotifEmail)
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
		<-ctx.Done()

		// <-ctx.Done() // tunggu sinyal cancel dari goroutine lain yang error
		log.Println("shutting down http server...")

		// harus ctx bg baru
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer publisherPool.CloseAllChann(shutdownCtx)

		defer cancel()
		return newServer.Shutdown(shutdownCtx) // asumsi router.Server punya method ini
	})

	err = waitGroup.Wait()
	if err != nil {
		// normally, log fatal untuk close all service. Tpi harus setelah service shutdown
		log.Fatalf("waitGroup last tail: %v", err)

	}

}
