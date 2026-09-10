package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"

	"msgqueue-luke.com/v2/internals/router"
	"msgqueue-luke.com/v2/internals/utils"
)

type StoreMain struct {
	Key   string
	Value string
}

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

	newServer, err := router.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	errWaitgroup, _ := errgroup.WithContext(ctxBg)
	if err != nil {
		log.Printf("errWaitGroup: %v", err)
		panic(err)
	}

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
