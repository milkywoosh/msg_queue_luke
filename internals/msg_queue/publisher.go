package msgqueue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type PublisherChan struct {
	mu sync.Mutex // kayanya gak kepake, karena lock sudah build in di lib amqp
	ch *amqp091.Channel
}

func NewPubChan(conn *amqp091.Connection) (*PublisherChan, error) {
	newChan, err := conn.Channel()
	if err != nil {
		log.Printf("err NewPubChan: %s", err.Error())
		return nil, err
	}

	if err := newChan.Confirm(false); err != nil {
		return nil, err
	}

	return &PublisherChan{
		ch: newChan,
	}, nil
}

func (p *PublisherChan) Close() error {
	return p.ch.Close()
}

func (p *PublisherChan) GetCh() *amqp091.Channel {
	return p.ch
}

// letak di main goroutine
func PublishOrder(ctx context.Context, ch *amqp091.Channel, exchg, routingKey, orderID, userId string) error {

	deferedConf, err := ch.PublishWithDeferredConfirmWithContext(
		ctx,
		exchg,
		routingKey, // connect exchange and queue
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			UserId:      userId,
			Timestamp:   time.Now(),
			Body:        []byte(orderID),
		},
	)
	if err != nil {
		return err
	}

	if deferedConf == nil {
		// setahu saya, nil berarti channel belum confirm mode
		return errors.New("channel not in confirm mode")
	}

	acked, err := deferedConf.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("wait confirm: %w", err)
	}
	if !acked {
		return errors.New("message broker rejected entry")
	}
	return nil

}
