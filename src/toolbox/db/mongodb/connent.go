package mongodb

import (
	"context"
	"fmt"

	"rabbitmq/lab-soltegm.com/src/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func NewDB(ctx context.Context, config *config.MongoDB) (*mongo.Database, error) {
	opts := options.Client().
		ApplyURI(config.DatabaseURL).
		SetConnectTimeout(config.ConnectTimeout).
		SetMaxPoolSize(uint64(config.MaxPoolSize)).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	client, err := mongo.Connect(
		ctx,
		opts,
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping to mongo failed: %w", err)
	}

	database := client.Database(config.DatabaseName)
	return database, nil
}
