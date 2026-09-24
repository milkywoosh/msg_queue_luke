package msgqueue

import (
	"context"
	"errors"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type PublisherPool struct {
	conn *amqp091.Connection
	pool chan *amqp091.Channel // pool of channel berisi struct(*amqp091.Channel)
}

func (p *PublisherPool) newChannel() (*amqp091.Channel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, err
	}

	// set mode confirm agar bisa di deferConfirm pakai API  ch.PublishWithDeferredConfirmWithContext()
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return nil, err
	}

	return ch, nil
}

func NewPublisherPool(conn *amqp091.Connection, sizePool int) (*PublisherPool, error) {
	p := &PublisherPool{conn: conn, pool: make(chan *amqp091.Channel, sizePool)}

	// create pool channel amqp up to 3(size) slot
	for i := 0; i < sizePool; i++ {
		ch, err := p.newChannel()
		if err != nil {
			return nil, err
		}
		p.pool <- ch
	}

	log.Printf("len pool after populated: %d", len(p.pool))

	return p, nil
}

func (p *PublisherPool) acquire(ctx context.Context) (*amqp091.Channel, error) {

	select {
	// p.pool = pool of amqp091.Channel yang Open dan Closed
	case ch := <-p.pool:
		if ch.IsClosed() {
			// if ch closed, trus create new amqp channl
			ch, err := p.newChannel()
			if err != nil {
				p.pool <- ch // balikin ke pool amqp091 channel
				// balikin ke request http, resource currently busy; pake err info general
				return nil, err
			}
		}
		return ch, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

var ErrBrokerNotConfirm = errors.New("Error Broker not confirm message")

func (p *PublisherPool) Publish(ctx context.Context, exchange, routingKey string, msg amqp091.Publishing) error {

	ch, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// balikin channel amqp091 ke pool if finnished
		p.pool <- ch
	}()

	deferredConfirm, err := ch.PublishWithDeferredConfirmWithContext(ctx, exchange, routingKey, false, false, msg)
	if err != nil {
		return err
	}
	ack, err := deferredConfirm.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !ack {
		return ErrBrokerNotConfirm
	}

	// publish berhasil dan broker sudah ack
	return nil

}

// func (p *PublisherPool)
