package infrastructure

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-web-crawler-service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"testing"
)

const (
	testRepoChannelURL      domain.Url             = "https://google.com/"
	testRepoApplicationName domain.ApplicationName = "Google"
	testRepoRating          domain.Rating          = 3.8
	testRepoRatingsAmount   domain.RatingsAmount   = 999
)

func TestChannelRepository_Save(t *testing.T) {
	options := mtest.NewOptions().ClientType(mtest.Mock).CollectionName(channelCollection)
	mt := mtest.New(t, options)
	defer mt.Close()

	mt.Run(
		"save channel successfully", func(t *mtest.T) {
			ctx := context.Background()
			channel := domain.NewChannel(
				testRepoApplicationName,
				testRepoChannelURL,
				testRepoRating,
				testRepoRatingsAmount,
			)

			t.AddMockResponses(
				mtest.CreateSuccessResponse(
					bson.E{
						Key:   "n",
						Value: 1,
					},
					bson.E{
						Key:   "nModified",
						Value: 1,
					},
				),
			)

			repository := NewMongoChannelRepository(t.DB)
			err := repository.Save(ctx, *channel)

			require.NoError(t, err)
			assert.NotNil(t, t.GetSucceededEvent())
		},
	)

	mt.Run(
		"save channel error", func(t *mtest.T) {
			ctx := context.Background()
			channel := domain.NewChannel(
				testRepoApplicationName,
				testRepoChannelURL,
				testRepoRating,
				testRepoRatingsAmount,
			)

			t.AddMockResponses(
				mtest.CreateCommandErrorResponse(
					mtest.CommandError{
						Code:    100,
						Message: "test error",
						Name:    "test",
					},
				),
			)

			repository := NewMongoChannelRepository(t.DB)
			err := repository.Save(ctx, *channel)

			require.Error(t, err)
		},
	)
}
