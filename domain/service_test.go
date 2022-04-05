package domain

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testServiceChannelURL      Url             = "https://google.com/"
	testServiceApplicationName ApplicationName = "Google"
	testServiceRating          Rating          = 3.8
	testServiceRatingsAmount   RatingsAmount   = 999
)

func TestChannelCrawlerProcessor_Crawl_Success(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &channelRepositoryMock{}
	webCrawlerMock := &rokuWebCrawlerMock{}

	channel := NewChannel(
		testServiceApplicationName,
		testServiceChannelURL,
		testServiceRating,
		testServiceRatingsAmount,
	)
	webCrawlerMock.On("CrawlChannel", ctx, testServiceChannelURL).Return(channel, nil)
	repositoryMock.On("Save", ctx, *channel).Return(nil)

	processor := NewChannelCrawlerProcessor(webCrawlerMock, repositoryMock)

	err := processor.Crawl(ctx, testServiceChannelURL)
	require.NoError(t, err)
	webCrawlerMock.AssertExpectations(t)
	repositoryMock.AssertExpectations(t)
}

func TestChannelCrawlerProcessor_Crawl_WebCrawlerFailed_ReturnsError(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &channelRepositoryMock{}
	webCrawlerMock := &rokuWebCrawlerMock{}

	crawlerErr := errors.New("crawler error")
	webCrawlerMock.On("CrawlChannel", ctx, testServiceChannelURL).Return(nil, crawlerErr)

	processor := NewChannelCrawlerProcessor(webCrawlerMock, repositoryMock)

	err := processor.Crawl(ctx, testServiceChannelURL)
	require.Error(t, err)
	require.ErrorIs(t, err, crawlerErr)
	webCrawlerMock.AssertExpectations(t)
	repositoryMock.AssertExpectations(t)
}

func TestChannelCrawlerProcessor_Crawl_RepositoryFailed_ReturnsError(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &channelRepositoryMock{}
	webCrawlerMock := &rokuWebCrawlerMock{}

	channel := NewChannel(
		testServiceApplicationName,
		testServiceChannelURL,
		testServiceRating,
		testServiceRatingsAmount,
	)
	webCrawlerMock.On("CrawlChannel", ctx, testServiceChannelURL).Return(channel, nil)

	repoErr := errors.New("repo err")
	repositoryMock.On("Save", ctx, *channel).Return(repoErr)

	processor := NewChannelCrawlerProcessor(webCrawlerMock, repositoryMock)

	err := processor.Crawl(ctx, testServiceChannelURL)
	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
	webCrawlerMock.AssertExpectations(t)
	repositoryMock.AssertExpectations(t)
}
