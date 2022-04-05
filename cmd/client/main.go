package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"go-web-crawler-service/config"
	grpcwebcrawler "go-web-crawler-service/protobuf/webcrawler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const (
	grpcHostEnv = "GRPC_HOST"
)

var (
	csvFile = flag.String("csv", "", "Path to CSV file contains all the urls to index")
)

// This one is ugly, I didn't focus on it :)
func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("got error when parsing config %v", err)
	}

	grpcConnString, err := resolveGRPCHost(cfg)
	if err != nil {
		log.Fatalf("could not resolve grpc host %v", err)
	}

	conn, err := grpc.DialContext(ctx, grpcConnString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect with GRPC server")
	}
	defer conn.Close()

	client := grpcwebcrawler.NewWebCrawlerServiceClient(conn)

	f, err := os.Open(*csvFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	urlsToSend := make(chan string)
	wg := &sync.WaitGroup{}

	go func() {
		data := make([]*grpcwebcrawler.CrawlerRequest, 0)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if len(data) > 0 {
					log.Printf("try to send batch to server %v", data)
					_, err := client.CrawlBatch(ctx, &grpcwebcrawler.BatchCrawlerRequest{Urls: data})

					if err != nil {
						panic(err)
					}
					log.Printf("sent batch to server")
					wg.Add(-len(data))
					data = make([]*grpcwebcrawler.CrawlerRequest, 0)
				}
			case d := <-urlsToSend:
				log.Printf("adding new url %s", d)
				data = append(data, &grpcwebcrawler.CrawlerRequest{Url: d})
			}
		}
	}()

	csvReader := csv.NewReader(f)

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(rec) != 1 {
			log.Fatal("unsupported CSV format")
		}

		if rec[0] == "page_url" {
			continue
		}

		wg.Add(1)
		urlsToSend <- rec[0]
	}

	if err != nil {
		log.Fatalf("unable to send channels for indexing, %v", err)
	}

	wg.Wait()
}

func resolveGRPCHost(cfg *config.Config) (string, error) {
	host, ok := os.LookupEnv(grpcHostEnv)
	if !ok {
		return "", fmt.Errorf("%s env is required", grpcHostEnv)
	}

	return fmt.Sprintf("%s:%d", host, cfg.GRPC.ServerPort), nil
}
