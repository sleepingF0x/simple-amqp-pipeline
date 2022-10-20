package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Consumer struct {
	conn       *Connection
	queue      string
	routingKey string
	autoDelete bool
	durable    bool
	args       amqp.Table
	handler    func([]byte) error
	workers    int
	quit       chan struct{}
	done       chan error
}

func NewConsumer(
	addr string,
	exchange string,
	exchangeType string,
	queue string,
	routingKey string,
	durable bool,
	autoDelete bool,
	args map[string]interface{},
	handler func([]byte) error,
	workers int) *Consumer {

	c := &Consumer{
		conn:       NewConnection(addr, exchange, exchangeType, true),
		queue:      queue,
		routingKey: routingKey,
		autoDelete: autoDelete,
		durable:    durable,
		handler:    handler,
		args:       args,
		workers:    workers,
		quit:       make(chan struct{}),
		done:       make(chan error),
	}
	return c
}

func (c *Consumer) Start() error {
	var err error

	if err = c.conn.Connect(); err != nil {
		return err
	}

	if _, err = c.conn.channel.QueueDeclare(
		c.queue,      // name
		true,         // durable
		c.autoDelete, // delete when unused
		false,        // exclusive
		false,        // no-wait
		c.args,       // arguments
	); err != nil {
		log.Println("queue declare error: ", err)
		_ = c.conn.channel.Close()
		_ = c.conn.conn.Close()
		return err
	}

	if err = c.conn.channel.QueueBind(
		c.queue,
		c.routingKey,
		c.conn.exchange,
		false,
		c.args,
	); err != nil {
		log.Println("queue bind error: ", err)
		_ = c.conn.channel.Close()
		_ = c.conn.conn.Close()
		return err
	}

	if err = c.conn.channel.Qos(
		c.workers,
		0,
		false,
	); err != nil {
		log.Println("channel Qos error: ", err)
		_ = c.conn.channel.Close()
		_ = c.conn.conn.Close()
		return err
	}

	go c.consumeOnConnect(c.done)
	return err
}

func (c *Consumer) Stop() error {
	close(c.quit)
	close(c.conn.quit)
	if !c.conn.conn.IsClosed() {
		// 关闭 SubMsg message delivery
		if err := c.conn.channel.Cancel("", true); err != nil {
			log.Println("AMQP consumer - channel cancel failed: ", err)
		}

		if err := c.conn.conn.Close(); err != nil {
			log.Println("AMQP consumer - connection close failed: ", err)
		}
	}
	// wait for handle() to exit
	defer log.Println("consumer AMQP shutdown OK")
	return <-c.done
}

func (c *Consumer) Handle(delivery <-chan amqp.Delivery) {
	for d := range delivery {
		go func(delivery amqp.Delivery) {
			if err := c.handler(delivery.Body); err == nil {
				_ = delivery.Ack(false)
			} else {
				// 重新入队，否则未确认的消息会持续占用内存
				_ = delivery.Reject(true)
			}
		}(d)
	}
}

func (c *Consumer) consumeOnConnect(done chan error) {
	cleanup := func() {
		log.Println("consumeOnConnect: deliveries channel closed")
		done <- nil
	}
	defer cleanup()

	var err error
	for {
		select {
		case <-c.conn.connected:
			for i := 0; i < c.workers; i++ {
				var delivery <-chan amqp.Delivery
				if delivery, err = c.conn.channel.Consume(
					c.queue, // queue
					"",      // consumer
					false,   // auto-ack
					false,   // exclusive
					false,   // no-local
					false,   // no-wait
					nil,     // args
				); err != nil {
					if !c.conn.conn.IsClosed() {
						log.Println("channel consume error, closing connection: ", err)
						_ = c.conn.channel.Close()
						_ = c.conn.conn.Close()
					}
				} else {
					go c.Handle(delivery)
				}
			}
		case <-c.quit:
			log.Println("consumer receive quit signal")
			return
		}
	}
}
