package tests

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"go-web-crawler-service/config"
	grpcwebcrawler "go-web-crawler-service/protobuf/webcrawler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net"
	"sync"
	"testing"
	"time"
)

const (
	testTimeout = 10 * time.Second

	testAcceptanceApplicationName                 = "Netflix"
	testAcceptanceApplicationRating               = "3.8"
	testAcceptanceApplicationRatingsAmount uint32 = 4195815
)

type channelMongoDTO struct {
	ApplicationName string    `bson:"applicationName"`
	Url             string    `bson:"url"`
	Rating          string    `bson:"rating"`
	NumberOfRatings uint32    `bson:"numberOfRatings"`
	UpdatedAt       time.Time `bson:"updatedAt"`
}

func TestAcceptanceWebCrawlerProcessesChannel_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg, err := config.ParseConfig()
	require.NoError(t, err)

	wg := &sync.WaitGroup{}

	waitForAPI(t, resolveGRPCHost(t, cfg))

	grpcConnection := connectGRPC(t, ctx, resolveGRPCHost(t, cfg), wg)
	grpcClient := grpcwebcrawler.NewWebCrawlerServiceClient(grpcConnection)
	db := createDB(t, ctx, cfg, wg)

	fakeSite := resolveFakeSiteURL(t)

	_, err = grpcClient.Crawl(
		ctx, &grpcwebcrawler.CrawlerRequest{
			Url: fakeSite,
		},
	)

	require.NoError(t, err)

	channelExistsInDB := func(url string) func() bool {
		return func() bool {

			var dto channelMongoDTO

			err := db.Collection("channel").FindOne(ctx, bson.M{"url": url}).Decode(&dto)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					return false
				}

				t.Fatalf("unexpected error on db %s", err)
			}

			if dto.ApplicationName != testAcceptanceApplicationName ||
				dto.Rating != testAcceptanceApplicationRating ||
				dto.NumberOfRatings != testAcceptanceApplicationRatingsAmount ||
				dto.Url != fakeSite {
				t.Fatalf("channel found but with different data %+v", dto)
			}

			return true
		}
	}

	assertWithTimeout(t, testTimeout, channelExistsInDB(fakeSite))

	_ = db.Drop(ctx)
	cancel()
	wg.Wait()
}

func assertWithTimeout(t *testing.T, timeoutVal time.Duration, cond func() bool) {
	const tickVal = 10 * time.Millisecond
	tick := time.Tick(tickVal)
	timeout := time.After(timeoutVal)

	for {
		select {
		case <-timeout:
			t.Fatalf("could not assert given condition without %s", timeoutVal)
		case <-tick:
			if cond() {
				return
			}
		}
	}
}

func waitForAPI(t *testing.T, host string) {
	assertWithTimeout(
		t, testTimeout, func() bool {
			timeout := 1 * time.Second
			_, err := net.DialTimeout("tcp", host, timeout)
			if err != nil {
				return false
			}

			return true
		},
	)
}
