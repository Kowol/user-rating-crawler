package tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-web-crawler-service/domain"
	"go-web-crawler-service/infrastructure"
	"testing"
)

const (
	testIntegrationApplicationName                 = "Netflix"
	testIntegrationApplicationRating               = 3.8
	testIntegrationApplicationRatingsAmount uint32 = 4195815
)

func TestIntegrationRodRokuWebCrawler_CrawlChannel_Success(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	defer ctx.Done()

	fakeSiteURL := resolveFakeSiteURL(t)
	browser := getHeadlessBrowser(ctx)
	crawler := infrastructure.NewRodRokuWebCrawler(browser)

	url, err := domain.NewURL(fakeSiteURL)
	require.NoError(t, err)

	channel, err := crawler.CrawlChannel(ctx, *url)
	require.NoError(t, err)
	assert.EqualValues(t, testIntegrationApplicationName, channel.ApplicationName)
	assert.EqualValues(t, testIntegrationApplicationRating, channel.Rating)
	assert.EqualValues(t, testIntegrationApplicationRatingsAmount, channel.NumberOfRatings)
	assert.EqualValues(t, fakeSiteURL, channel.Url)
}

func TestIntegrationRodRokuWebCrawler_CrawlChannel_EmptyRating_Success(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	defer ctx.Done()

	fakeSiteURL := resolveFakeSiteURL(t)
	browser := getHeadlessBrowser(ctx)
	crawler := infrastructure.NewRodRokuWebCrawler(browser)

	url, err := domain.NewURL(fmt.Sprintf("%s/no-data.html", fakeSiteURL))
	require.NoError(t, err)

	channel, err := crawler.CrawlChannel(ctx, *url)
	require.NoError(t, err)
	assert.EqualValues(t, 0, channel.Rating)
	assert.EqualValues(t, 0, channel.NumberOfRatings)
}

func TestIntegrationRodRokuWebCrawler_CrawlChannel_NoElementsFound_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	defer ctx.Done()

	fakeSiteURL := resolveFakeSiteURL(t)
	browser := getHeadlessBrowser(ctx)
	crawler := infrastructure.NewRodRokuWebCrawler(browser)

	url, err := domain.NewURL(fmt.Sprintf("%s/invalid.html", fakeSiteURL))
	require.NoError(t, err)

	_, err = crawler.CrawlChannel(ctx, *url)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestIntegrationRodRokuWebCrawler_CrawlChannel_PageNotFound_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	defer ctx.Done()

	browser := getHeadlessBrowser(ctx)
	crawler := infrastructure.NewRodRokuWebCrawler(browser)

	url, err := domain.NewURL("http://localhost:9999")
	require.NoError(t, err)

	_, err = crawler.CrawlChannel(ctx, *url)
	require.Error(t, err)
}
