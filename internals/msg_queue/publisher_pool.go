package msgqueue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type PublisherPool struct {
	conn     *amqp091.Connection
	pool     chan *amqp091.Channel // pool of channel berisi struct(*amqp091.Channel)
	sizePool int
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
	p := &PublisherPool{conn: conn, pool: make(chan *amqp091.Channel, sizePool), sizePool: sizePool}

	// create pool channel amqp up to 3(size) slot
	for i := 0; i < p.sizePool; i++ {
		ch, err := p.newChannel()
		if err != nil {
			return nil, err
		}
		// populating ch ke pool CHANNEL buffer isi amqp091
		p.pool <- ch
	}

	log.Printf("len pool after populated: %d", len(p.pool))

	return p, nil
}

func (p *PublisherPool) acquire(ctx context.Context) (*amqp091.Channel, error) {
	var ch *amqp091.Channel = nil
	select {
	// p.pool = pool of amqp091.Channel yang Open dan Closed
	case ch = <-p.pool:
		// kapan dan kenapa channel isClosed: [connection mati, exception dr broker, di-close()
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

	if deferredConfirm == nil {
		return errors.New("channel not in confirm mode")
	}

	ctxTO, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()

	ack, err := deferredConfirm.WaitContext(ctxTO)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("confirm timeout deadline exceeded: %w", err)
		}
		return fmt.Errorf("wait confirm: %w", err)
	}

	if !ack {
		return ErrBrokerNotConfirm
	}

	// publish berhasil dan broker sudah ack
	return nil

}

func (p *PublisherPool) CloseAllChann(ctx context.Context) error {

	var errs []error

	log.Printf("CloseAllChann ....")
	log.Printf("ctx err at start: %v, len(pool)=%d, cap(pool)=%d", ctx.Err(), len(p.pool), cap(p.pool))

	for i := 0; i < p.sizePool; i++ {
		select {
		case ch := <-p.pool:
			log.Printf("case Pool")
			if !ch.IsClosed() {
				if err := ch.Close(); err != nil {
					errs = append(errs, fmt.Errorf("close channel %d: %w", i, err))
				}
				log.Printf("Closed Pool done")
			}
			log.Printf("case Not Closed Pool")
		// pool nggak ngasih channel tepat waktu, fallback nya ini boy
		case <-ctx.Done():
			log.Printf("case Done")
			errs = append(errs, fmt.Errorf("ctx done at channel %d: %w", i, ctx.Err()))
			// klo ctx.Err() nil pasti err is nil

		}
	}

	// result nil klo gak semua yg di-append (err is nil)
	return errors.Join(errs...)
}

// func (p *PublisherPool)
