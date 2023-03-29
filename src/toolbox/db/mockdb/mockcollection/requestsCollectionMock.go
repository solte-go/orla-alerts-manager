package requestscollection

import (
	"context"
	"orla-alerts/solte.lab/src/model"
	"orla-alerts/solte.lab/src/toolbox/db/contract"
)

type RequestCollectionMock struct {
}

func NewRequestCollectionMock() contract.SharedCollectionContract {
	return &RequestCollectionMock{}
}

func (rc *RequestCollectionMock) NewAlert(ctx context.Context, alerts []*model.Alert) error {
	return nil
}

func (rc *RequestCollectionMock) GetAllAlerts(ctx context.Context) ([]*model.Alert, error) {
	return nil, nil
}

func (rc *RequestCollectionMock) DropDataBase(ctx context.Context) error {
	return nil
}

func (rc *RequestCollectionMock) ClearDatabase(ctx context.Context) error {
	return nil
}
