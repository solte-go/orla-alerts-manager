package db

import (
	"context"

	"orla-alert/solte.lab/src/config"
	"orla-alert/solte.lab/src/toolbox/db/contract"
	"orla-alert/solte.lab/src/toolbox/db/mongodb"
	sharedcollection "orla-alert/solte.lab/src/toolbox/db/mongodb/collections"

	"go.uber.org/zap"
)

type DB struct {
	SharedDB contract.SharedCollectionContract
}

func InitDatabase(ctx context.Context, conf config.Config, logger *zap.Logger) (*DB, error) {
	db, err := mongodb.NewDB(ctx, conf.MongoDB)
	if err != nil {
		return nil, err
	}

	instance := &DB{
		SharedDB: sharedcollection.NewSharedCollection(db.Collection(conf.MongoDB.Collection), logger),
	}

	return instance, nil
}
