package infrastructure

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"go-web-crawler-service/domain"
	"log"
)

type amqpPublisher struct {
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func NewAmqpPublisher(channel *amqp.Channel, exchange string, routingKey string) *amqpPublisher {
	return &amqpPublisher{
		channel:    channel,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (p *amqpPublisher) Schedule(ctx context.Context, url domain.Url) error {
	err := p.channel.Publish(
		p.exchange, p.routingKey, false, false, amqp.Publishing{
			Headers:      amqp.Table{},
			ContentType:  "text/plain",
			Body:         []byte(url),
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		log.Printf("Failed to publish message with url %s\n", url)
		return fmt.Errorf("failed to publish message with url %s", url)
	}

	log.Printf("Successfully published message to crawl, url: %s\n", url)
	return nil
}
