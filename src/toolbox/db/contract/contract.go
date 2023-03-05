package contract

import (
	"context"
	"rabbitmq/lab-soltegm.com/src/model"
)

type SharedCollectionContract interface {
	// AddDataToCollection will add data to collection with provided request "id".
	//The "id" automatically will be converted to collection name.
	NewAlert(ctx context.Context, domains []*model.Alert) error
	// GetAllAlerts Retrieve all alerts from the collection.
	GetAllAlerts(ctx context.Context) ([]*model.Alert, error)
}

type TeardownDB interface {
	DropDatabase(ctx context.Context) error
	ClearDatabase(ctx context.Context) error
}
