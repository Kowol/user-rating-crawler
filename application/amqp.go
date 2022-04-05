package application

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"go-web-crawler-service/domain"
	"log"
	"time"
)

const (
	consumerTag = "web-crawler"
)

var (
	rateLimitMilliseconds = 200
)

type amqpApp struct {
	ch            *amqp.Channel
	queueName     string
	processor     domain.ChannelCrawlerProcessor
	workersAmount int
}

func NewAmqpApplication(
	ch *amqp.Channel,
	queueName string,
	processor domain.ChannelCrawlerProcessor,
	workersAmount int,
) *amqpApp {
	return &amqpApp{ch: ch, queueName: queueName, processor: processor, workersAmount: workersAmount}
}

func (a *amqpApp) Run(ctx context.Context, notifyStart func(), notifyEnd func()) error {
	log.Println("Starting Crawler worker")

	urlsToProcess, err := a.ch.Consume(a.queueName, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to spawn a consumer, %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = a.ch.Cancel(consumerTag, false)
	}()

	log.Printf("Spawning %d workers\n", a.workersAmount)

	rateLimiter := time.Tick(time.Duration(rateLimitMilliseconds) * time.Millisecond)
	for i := 0; i < a.workersAmount; i++ {
		notifyStart()
		go func() {
			defer notifyEnd()
			a.spawnConsumer(ctx, urlsToProcess, rateLimiter)
			log.Println("channel closed")
		}()
	}

	return nil
}

func (a *amqpApp) spawnConsumer(ctx context.Context, urlsToProcess <-chan amqp.Delivery, rateLimiter <-chan time.Time) {
	for d := range urlsToProcess {
		url, err := domain.NewURL(string(d.Body))
		if err != nil {
			nackErr := d.Nack(false, false)
			if nackErr != nil {
				log.Println("failed to ack/nack message")
			}
		}

		<-rateLimiter

		log.Printf("Starting processing message with url: %s\n", *url)

		start := time.Now()
		processErr := a.processor.Crawl(ctx, *url)
		elapsed := time.Since(start)

		log.Printf("Processing message with url: %s took %s\n", *url, elapsed)

		var ackErr error
		if processErr != nil {
			log.Printf("Failed to consume a message with url, %v\n", processErr)
			ackErr = d.Nack(
				false,
				false,
			) // Message could end up in dead letter queue, we could also configure messages to be rerouted to the processor queue after some time.
			// It mostly fails because of timeouts
		} else {
			log.Printf("Successfully processed message with url: %s\n", *url)
			ackErr = d.Ack(false)
		}

		if ackErr != nil {
			log.Println("failed to ack/nack message")
		}
	}
}
