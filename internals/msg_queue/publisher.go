package msgqueue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type PublisherChan struct {
	ch *amqp091.Channel
}

func NewPubChan(conn *amqp091.Connection) (*PublisherChan, error) {
	newChan, err := conn.Channel()
	if err != nil {
		log.Printf("err NewPubChan: %s", err.Error())
		return nil, err
	}

	if err := newChan.Confirm(false); err != nil {
		_ = newChan.Close()
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
func PublishOrder(ctx context.Context, chPublisher *amqp091.Channel, exchg, routingKey, orderID, userId string) error {

	deferedConf, err := chPublisher.PublishWithDeferredConfirmWithContext(
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
		// probs nil berarti channel belum confirm mode
		return errors.New("channel not in confirm mode")
	}

	// wait context kasih 5 detik agar goroutine tidak tunggu too long
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()

	acked, err := deferedConf.WaitContext(ctxWithTimeout)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("confirm wait timeout: %w", err)
		}
		return fmt.Errorf("wait confirm: %w", err)
	}
	if !acked {
		return errors.New("message broker rejected entry")
	}
	return nil

}
