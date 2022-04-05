package cmd

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func InitializeAMQPExchange(ch *amqp.Channel, exchangeName string, queueName string, routingKey string) error {
	err := ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not declare AMQP exchange, %w", err)
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not declare AMQP queue, %w", err)
	}

	err = ch.QueueBind(queueName, routingKey, exchangeName, true, nil)
	if err != nil {
		return fmt.Errorf("could not bind AMQP queue with the exchange")
	}

	return nil
}

func GetAMQPChannel(
	ctx context.Context,
	conn *amqp.Connection,
	cancel context.CancelFunc,
	notifyStart func(),
	notifyDone func(),
) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not open new channel, %w", err)
	}

	disconnected := make(chan *amqp.Error)
	ch.NotifyClose(disconnected)
	go func() {
		err := <-disconnected
		if err != nil {
			fmt.Printf("AMQP connection lost, %s", err)
		}

		cancel()
	}()

	notifyStart()

	go func() {
		defer notifyDone()
		<-ctx.Done()
		_ = ch.Close()
		fmt.Println("AMQP chanel closed")
	}()

	return ch, nil
}

func GetAMQPConn(ctx context.Context, url string, notifyStart func(), notifyDone func()) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("couldn not connect to AMQP, %w", err)
	}
	notifyStart()
	go func() {
		defer notifyDone()
		<-ctx.Done()
		err := conn.Close()
		if err != nil {
			log.Printf("could not close AMQP connection %s\n", err)
		}

		log.Println("AMQP connection closed")
	}()

	return conn, nil
}
