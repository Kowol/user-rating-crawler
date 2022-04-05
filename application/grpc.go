package application

import (
	"context"
	"go-web-crawler-service/domain"
	grpcwebcrawler "go-web-crawler-service/protobuf/webcrawler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	grpcwebcrawler.UnimplementedWebCrawlerServiceServer
	publisher domain.ChannelCrawlerScheduler
}

func NewServer(publisher domain.ChannelCrawlerScheduler) *server {
	return &server{publisher: publisher}
}

func (s *server) Crawl(ctx context.Context, request *grpcwebcrawler.CrawlerRequest) (*grpcwebcrawler.Empty, error) {
	url, err := domain.NewURL(request.Url)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "request validation failed")
	}

	err = s.publisher.Schedule(ctx, *url)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to publish message")
	}

	return &grpcwebcrawler.Empty{}, nil
}

func (s *server) CrawlBatch(ctx context.Context, request *grpcwebcrawler.BatchCrawlerRequest) (
	*grpcwebcrawler.Empty,
	error,
) {
	for _, urlInBatch := range request.Urls { // TODO: Some more sophisticated solution?
		_, err := s.Crawl(ctx, urlInBatch)

		if err != nil {
			return nil, err // TODO: It should probably report which urls failed to publish
		}
	}

	return &grpcwebcrawler.Empty{}, nil
}
