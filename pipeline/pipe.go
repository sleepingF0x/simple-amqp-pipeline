package pipeline

import (
	"log"
	"simple-amqp-pipeline/amqp"
	"simple-amqp-pipeline/config"
)

type Pipeline struct {
	src           *amqp.Consumer
	dst           *amqp.Producer
	dstRoutingKey string
}

func NewPipeline(srcConf config.RabbitMQConf, destConf config.RabbitMQConf, workers int) (*Pipeline, error) {
	c := &Pipeline{}
	src := amqp.NewConsumer(
		srcConf.Uri,
		srcConf.Exchange.ExchangeName,
		srcConf.Exchange.ExchangeType,
		srcConf.Queue.QueueName,
		srcConf.Queue.RoutingKey,
		srcConf.Queue.Durable,
		srcConf.Queue.AutoDelete,
		srcConf.Queue.Arguments,
		c.move,
		workers,
	)
	dst, err := amqp.NewProducer(destConf.Uri, destConf.Exchange.ExchangeName, destConf.Exchange.ExchangeType)

	c.src = src
	c.dst = dst
	c.dstRoutingKey = destConf.Queue.RoutingKey
	defer log.Println("create new pipeline")
	return c, err
}

func (p *Pipeline) move(msg []byte) error {
	p.dst.Publish(msg, p.dstRoutingKey)
	return nil
}

func (p *Pipeline) Start() error {
	return p.src.Start()
}

func (p *Pipeline) Stop() {
	p.src.Stop()
	p.dst.Stop()
}
