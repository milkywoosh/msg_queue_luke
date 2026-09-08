package main

import (
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/utils"
)

func main() {

	cfg, err := utils.LoadConfig("./")
	log.Println(cfg)
	if err != nil {
		log.Printf("err load config: %s", err.Error())
		return
	}

	//    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	log.Printf("cfg rabbit mq: %s", cfg.RabbitMQPassword)
	connAmpq, err := amqp091.Dial(cfg.RabbitMQDial)
	if err != nil {
		panic(err)
	}
	defer connAmpq.Close()
	fmt.Printf("test :%s", "message queue")
}
