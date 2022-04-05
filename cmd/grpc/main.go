package main

import (
	"context"
	"fmt"
	"go-web-crawler-service/application"
	"go-web-crawler-service/cmd"
	"go-web-crawler-service/config"
	"go-web-crawler-service/infrastructure"
	grpcwebcrawler "go-web-crawler-service/protobuf/webcrawler"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("got error when parsing config %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	wg := &sync.WaitGroup{}
	notifyStart := func() {
		wg.Add(1)
	}

	notifyDone := func() {
		wg.Done()
	}

	conn, err := cmd.GetAMQPConn(ctx, cfg.AMQP.URL, notifyStart, notifyDone)
	if err != nil {
		log.Fatalf("failed to open RabbitMQ connection: %v", err)
	}

	ch, err := cmd.GetAMQPChannel(ctx, conn, cancel, notifyStart, notifyDone)
	if err != nil {
		log.Fatalf("failed to open RabbitMQ channel: %v", err)
	}

	err = cmd.InitializeAMQPExchange(ch, cfg.AMQP.ExchangeName, cfg.AMQP.QueueName, cfg.AMQP.RoutingKey)
	if err != nil {
		log.Fatalf("failed to initialize queues and exchanges: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.ServerPort))
	if err != nil {
		log.Fatalf("failed to listen on %d", cfg.GRPC.ServerPort)
	}

	publisher := infrastructure.NewAmqpPublisher(ch, cfg.AMQP.ExchangeName, cfg.AMQP.RoutingKey)

	grpcServer := grpc.NewServer()
	grpcwebcrawler.RegisterWebCrawlerServiceServer(
		grpcServer,
		application.NewServer(publisher),
	)

	notifyStart()
	go func() {
		defer notifyDone()
		<-ctx.Done()
		grpcServer.GracefulStop()
		log.Println("GRPC server stooped")
	}()

	log.Printf("starting GRPC server on port %d\n", cfg.GRPC.ServerPort)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("could not start GRPC server: %v", err)
	}

	wg.Wait()
}
