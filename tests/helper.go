package tests

import (
	"context"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/require"
	"go-web-crawler-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"sync"
	"testing"
)

const (
	fakeSiteUrlEnv = "FAKE_SITE"
	grpcHostEnv    = "GRPC_HOST"
)

func resolveFakeSiteURL(t *testing.T) string {
	url, ok := os.LookupEnv(fakeSiteUrlEnv)
	require.True(t, ok, "%s env is required", fakeSiteUrlEnv)

	return url
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

func connectGRPC(t *testing.T, ctx context.Context, url string, wg *sync.WaitGroup) *grpc.ClientConn {
	conn, err := grpc.DialContext(
		ctx,
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	require.NoError(t, err)

	wg.Add(1)
	go func() {
		<-ctx.Done()
		_ = conn.Close()
		fmt.Println("GRPC Client closed")
		wg.Done()
	}()

	return conn
}

func resolveGRPCHost(t *testing.T, cfg *config.Config) string {
	host, ok := os.LookupEnv(grpcHostEnv)
	require.True(t, ok, "%s env is required", grpcHostEnv)

	return fmt.Sprintf("%s:%d", host, cfg.GRPC.ServerPort)
}

func createDB(t *testing.T, ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) *mongo.Database {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.DSN))
	require.NoError(t, err)

	wg.Add(1)
	go func() {
		<-ctx.Done()
		_ = client.Disconnect(ctx)
		fmt.Println("DB closed")
		wg.Done()
	}()

	return client.Database(cfg.Database.DatabaseName)
}
