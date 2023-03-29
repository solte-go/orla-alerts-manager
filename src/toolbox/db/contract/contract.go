package contract

import (
	"context"
	"orla-alerts/solte.lab/src/model"
)

type SharedCollectionContract interface {
	// NewAlert will add data to collection.
	//The "id" automatically will be generated.
	NewAlert(ctx context.Context, domains []*model.Alert) error
	// GetAllAlerts Retrieve all alerts from the collection.
	GetAllAlerts(ctx context.Context) ([]*model.Alert, error)
}

type TeardownDB interface {
	DropDatabase(ctx context.Context) error
	ClearDatabase(ctx context.Context) error
}
