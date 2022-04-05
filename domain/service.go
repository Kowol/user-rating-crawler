package domain

import (
	"context"
	"fmt"
	"log"
)

type ChannelCrawlerScheduler interface {
	Schedule(ctx context.Context, url Url) error
}

type ChannelCrawlerProcessor interface {
	Crawl(ctx context.Context, url Url) error
}

type ChannelRepository interface {
	Save(ctx context.Context, channel Channel) error
}

type RokuWebCrawler interface {
	CrawlChannel(ctx context.Context, url Url) (*Channel, error)
}

type channelCrawlerProcessor struct {
	webCrawler        RokuWebCrawler
	channelRepository ChannelRepository
}

func NewChannelCrawlerProcessor(webCrawler RokuWebCrawler, repository ChannelRepository) *channelCrawlerProcessor {
	return &channelCrawlerProcessor{
		webCrawler:        webCrawler,
		channelRepository: repository,
	}
}

func (p *channelCrawlerProcessor) Crawl(ctx context.Context, url Url) error {
	log.Printf("Starting crawling url: %s\n", url)
	channel, err := p.webCrawler.CrawlChannel(ctx, url)
	if err != nil {
		log.Printf("could not crawl channel %s, error: %v\n", url, err)
		return fmt.Errorf("could not crawl channel %s, error: %w", url, err)
	}

	err = p.channelRepository.Save(ctx, *channel)
	if err != nil {
		log.Printf("Could not save crawled channel data, url: %s, error: %v\n", url, err)
		return fmt.Errorf("could not save crawled channel data, url: %s, error: %w", url, err)
	}

	log.Printf("Crawled url: %s, received channel data: %+v\n", url, *channel)
	return nil
}
