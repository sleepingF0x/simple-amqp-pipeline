package amqp

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

type Producer struct {
	conn         *Connection
	channelMutex sync.Mutex
	active       bool
}

func NewProducer(addr string, exchange string, exchangeType string) (*Producer, error) {
	p := &Producer{
		conn:   NewConnection(addr, exchange, exchangeType, true),
		active: false,
	}
	err := p.conn.Connect()
	if err != nil {
		log.Println("new producer error : ", err)
		return nil, err
	} else {
		p.active = true
	}
	return p, nil
}

func (p *Producer) Stop() {
	p.active = false
	close(p.conn.quit)
	if !p.conn.conn.IsClosed() {
		if err := p.conn.channel.Cancel("", true); err != nil {
			log.Println("AMQP consumer - channel cancel failed: ", err)
		}
		if err := p.conn.conn.Close(); err != nil {
			log.Println("AMQP consumer - connection close failed: ", err)
		}
	}
	defer log.Println("producer AMQP shutdown OK")
	return
}

func (p *Producer) Publish(msg []byte, routingKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
	clear:
		for {
			select {
			case <-p.conn.connected:
			default:
				break clear
			}
		}
		if !p.active {
			log.Println("producer not prepared")
			return
		}
		p.channelMutex.Lock()
		err := p.conn.channel.PublishWithContext(ctx,
			p.conn.exchange, // exchange
			routingKey,      // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "application/json",
				ContentEncoding: "",
				DeliveryMode:    amqp.Persistent,
				Body:            msg,
				Priority:        0,
			},
		)
		p.channelMutex.Unlock()
		if err != nil {
			log.Println("AMQP exchange Publish: ", err)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}
}
