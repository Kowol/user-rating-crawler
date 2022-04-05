package main

import (
	"context"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go-web-crawler-service/application"
	"go-web-crawler-service/cmd"
	"go-web-crawler-service/config"
	"go-web-crawler-service/domain"
	"go-web-crawler-service/infrastructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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

	db, err := getMongoDB(ctx, cfg.Database.DSN, cfg.Database.DatabaseName, notifyStart, notifyDone)
	if err != nil {
		log.Fatalf("failed to create mongo connection: %v", err)
	}

	browser := getHeadlessBrowser(ctx)
	webCrawler := infrastructure.NewRodRokuWebCrawler(browser)

	repo := infrastructure.NewMongoChannelRepository(db)
	processor := domain.NewChannelCrawlerProcessor(webCrawler, repo)
	app := application.NewAmqpApplication(ch, cfg.AMQP.QueueName, processor, cfg.Crawler.WorkersAmount)

	err = app.Run(ctx, notifyStart, notifyDone)
	if err != nil {
		log.Fatalf("failed to run amqp app: %v", err)
	}

	wg.Wait()
}

func getMongoDB(
	ctx context.Context,
	dsn string,
	databaseName string,
	notifyStart func(),
	notifyDone func(),
) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	notifyStart()
	go func() {
		defer notifyDone()
		<-ctx.Done()
		err := client.Disconnect(ctx)
		if err != nil {
			log.Printf("failed to disconnect from db: %v\n", err)
		} else {
			log.Println("MongoDB connection closed")
		}
	}()

	return client.Database(databaseName), nil
}

func getHeadlessBrowser(ctx context.Context) *rod.Browser {
	u := launcher.New().Bin("/usr/bin/chromium-browser").MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	go func() {
		<-ctx.Done()
		browser.MustClose()
	}()

	return browser
}
