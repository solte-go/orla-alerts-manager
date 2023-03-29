package sharedcollection

import (
	"context"
	"fmt"

	"orla-alert/solte.lab/src/model"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type SharedCollection struct {
	logger     *zap.Logger
	collection *mongo.Collection
}

// NewSharedCollection Creates new instance of RequestsCollection.
func NewSharedCollection(collection *mongo.Collection, logger *zap.Logger) *SharedCollection {
	return &SharedCollection{
		logger:     logger,
		collection: collection,
	}
}

// NewAlert will add data to collection with provided request "id".
// The "id" automatically will be converted to collection name.
func (sh *SharedCollection) NewAlert(ctx context.Context, data []*model.Alert) error {
	operations := make([]mongo.WriteModel, 0, len(data))

	if len(data) == 0 {
		sh.logger.Debug("there is no data for update")
		return nil
	}

	for _, d := range data {
		operation := mongo.NewInsertOneModel()
		operation.SetDocument(d)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(false)

	result, err := sh.collection.BulkWrite(ctx, operations, &bulkOption)
	if err != nil {
		return errors.Wrap(err, "mongodb: can't update/replace documents. error")
	}

	sh.logger.Debug(fmt.Sprintf("Documents inserted: %v, modifided: %v", result.UpsertedCount, result.ModifiedCount))
	return nil
}

// GetAllAlerts Retrieve all alerts from the collection.
func (sh *SharedCollection) GetAllAlerts(ctx context.Context) ([]*model.Alert, error) {
	var result []*model.Alert

	cursor, err := sh.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "mongodb: error occurred during request GetAllAlerts(). error")
	}

	if err = cursor.All(ctx, &result); err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "mongodb: error occurred during decoding data GetProjects(). error")
		}
	}
	return result, nil
}

func (sh *SharedCollection) ClearDatabase(ctx context.Context) error {
	err := sh.collection.Drop(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't delete document")
	}
	return nil
}

// DropDataBase this func not included to RequestsCollectionContract. Execution will drop hole database.
// Only used in a test and cast to an interface within a performing test.
func (rc *SharedCollection) DropDatabase(ctx context.Context) error {
	if err := rc.collection.Database().Drop(ctx); err != nil {
		return fmt.Errorf("couldn't drop collection: %w", err)
	}
	return nil
}

// // EnrichWithDRSData Enriching Record with DRS service data.
// func (rc *SharedCollection) EnrichWithDRSData(ctx context.Context, id uint64, input <-chan model.UpdateNRDWithDRS) error {
// 	cursor := rc.db.Collection(rc.tempIndexName(id))

// 	start := time.Now()
// 	var (
// 		bulkSize = 1000
// 		res      model.UpdateNRDWithDRS
// 		ok       bool
// 	)

// 	operations := make([]mongo.WriteModel, 0, bulkSize)

// 	for {
// 		select {
// 		case res, ok = <-input:
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		}
// 		if !ok {
// 			break
// 		}

// 		operation := mongo.NewUpdateOneModel()
// 		operation.SetFilter(bson.D{{"_id", bson.D{{"$eq", model.CreateHashID(res.Domain, id)}}}})

// 		operation.SetUpdate(bson.D{{"$set", bson.D{{"DRSPart", res.DRSPart}}}})

// 		operations = append(operations, operation)

// 		if len(operations) == bulkSize {
// 			bulkOption := options.BulkWriteOptions{}
// 			bulkOption.SetOrdered(false)

// 			result, err := cursor.BulkWrite(ctx, operations, &bulkOption)
// 			if err != nil {
// 				return errors.Wrap(err, "mongodb: can't update/replace documents. error")
// 			}
// 			rc.logger.Debug(fmt.Sprintf("Documents matched: %v, Documents modified: %v", result.MatchedCount, result.ModifiedCount))
// 			operations = operations[:0]
// 		}
// 	}

// 	if len(operations) > 0 {
// 		bulkOption := options.BulkWriteOptions{}
// 		bulkOption.SetOrdered(false)

// 		result, err := cursor.BulkWrite(ctx, operations, &bulkOption)
// 		if err != nil {
// 			return errors.Wrap(err, "mongodb: can't update/replace documents. error")
// 		}
// 		rc.logger.Debug(fmt.Sprintf("Documents matched: %v, Documents modified: %v", result.MatchedCount, result.ModifiedCount))
// 	}
// 	rc.logger.Debug(fmt.Sprintf("request took %v", time.Now().Sub(start)))
// 	return nil
// }

// // DeleteReQueuedRecords deletes re-queued messages from the collection so that the same entry is not re-queued multiple times.
// func (rc *SharedCollection) DeleteReQueuedRecords(ctx context.Context, id uint64, recordsHashID []string) error {
// 	cursor := rc.db.Collection(rc.tempIndexName(id))
// 	operations := make([]mongo.WriteModel, 0, len(recordsHashID))

// 	for _, hashID := range recordsHashID {
// 		operation := mongo.NewDeleteOneModel()
// 		operation.SetFilter(bson.M{"_id": hashID})
// 		operations = append(operations, operation)
// 	}

// 	bulkOption := options.BulkWriteOptions{}
// 	bulkOption.SetOrdered(false)

// 	result, err := cursor.BulkWrite(ctx, operations, &bulkOption)
// 	if err != nil {
// 		return errors.Wrap(err, "mongodb: can't update/replace documents. error")
// 	}

// 	rc.logger.Debug(fmt.Sprintf("Documents deleted: %v", result.DeletedCount))
// 	return nil
// }

// // DeleteCollection deletes the collection with the provided id and returns an error if encountered.
// func (rc *SharedCollection) DeleteCollection(ctx context.Context, id uint64) error {
// 	err := rc.db.Collection(rc.tempIndexName(id)).Drop(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	rc.logger.Debug(fmt.Sprintf("Collection dropped: %v", rc.tempIndexName(id)))
// 	return nil
// }

// // tempIndexName internal function to convert "id" to collection name.
// func (rc *SharedCollection) tempIndexName(id uint64) string {
// 	return fmt.Sprintf("nrd_queue_%d", id)
// }

// ClearDatabase this func not included to RequestsCollectionContract. Execution will drop collection with provided name.
// Only used in a test and cast to an interface within a performing test.

// operation.SetFilter(bson.M{"_id": d.HashID})
// 		operation.SetUpdate(bson.A{
// 			bson.M{
// 				"$set": bson.M{
// 					"NRDQueue": bson.M{
// 						"$cond": bson.M{
// 							"if": bson.M{
// 								"$or": bson.A{
// 									bson.M{
// 										"$gt": bson.A{
// 											d.Priority,
// 											"$NRDQueue.priority",
// 										},
// 									},
// 									bson.M{
// 										"$eq": bson.A{
// 											"$NRDQueue",
// 											nil,
// 										},
// 									},
// 								},
// 							},
// 							"then": domain.NRDQueue,
// 							"else": "$NRDQueue",
// 						},
// 					},
// 				},
// 			},
// 		})
