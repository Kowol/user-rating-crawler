package infrastructure

import (
	"context"
	"fmt"
	"go-web-crawler-service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	channelCollection = "channel"
)

type mongoChannelRepository struct {
	db *mongo.Database
}

func NewMongoChannelRepository(db *mongo.Database) *mongoChannelRepository {
	return &mongoChannelRepository{
		db: db,
	}
}

type channelMongoDTO struct {
	ApplicationName string    `bson:"applicationName"`
	Url             string    `bson:"url"`
	Rating          string    `bson:"rating"`
	NumberOfRatings uint32    `bson:"numberOfRatings"`
	UpdatedAt       time.Time `bson:"updatedAt"`
}

func newChannelMongoDTO(channel domain.Channel) channelMongoDTO {
	return channelMongoDTO{
		ApplicationName: string(channel.ApplicationName),
		Url:             string(channel.Url),
		Rating:          fmt.Sprintf("%.1f", channel.Rating),
		NumberOfRatings: uint32(channel.NumberOfRatings),
		UpdatedAt:       time.Now(),
	}
}

func (r *mongoChannelRepository) Save(ctx context.Context, channel domain.Channel) error {
	dto := newChannelMongoDTO(channel)
	upsert := true
	_, err := r.getCollection().UpdateOne(
		ctx,
		bson.M{"applicationName": channel.ApplicationName}, // TODO: Create unique index on app name
		bson.M{"$set": dto},
		&options.UpdateOptions{Upsert: &upsert},
	)

	if err != nil {
		return fmt.Errorf("failed to save channel in MongoDB collection: %v, error: %w", dto, err)
	}

	return nil
}

func (r *mongoChannelRepository) getCollection() *mongo.Collection {
	return r.db.Collection(channelCollection)
}
