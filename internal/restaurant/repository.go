package restaurant

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AggregateDB struct {
	ID      string    `bson:"_id"`
	Version int       `bson:"version"`
	Events  []EventDB `bson:"events"`
}
type EventDB struct {
	AggregateID string    `bson:"_id"`
	Timestamp   time.Time `bson:"timestamp"`
	Category    string    `bson:"category"`
	Type        string    `bson:"type"`
	Data        bson.M    `bson:"data"`
}

type BaseRepository struct {
	log              zerolog.Logger
	eventsCollection *mongo.Collection
}

func NewBaseRepository(mongoDB *mongo.Database, log zerolog.Logger) (b *BaseRepository, err error) {
	b = &BaseRepository{
		log:              log,
		eventsCollection: mongoDB.Collection("events"),
	}
	return
}

func (repo *BaseRepository) getAggregateDBModel(event Event, version int) EventDB {
	return EventDB{
		AggregateID: event.AggregateID(),
		Timestamp:   time.Now(),
		Category:    event.EventCategory(),
		Type:        event.EventType(),
		Data:        event.Data(),
	}
}

func (repo *BaseRepository) loadAggregate(id string) (aggregate AggregateDB, err error) {
	res := repo.eventsCollection.FindOne(context.Background(), bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		aggregate = AggregateDB{
			ID:      id,
			Version: 0,
			Events:  []EventDB{},
		}
		return
	}
	if res.Err() != nil {
		err = fmt.Errorf("cannot load aggregate from DB: %w", res.Err())
		return
	}

	res.Decode(&aggregate)

	return
}

func (repo *BaseRepository) saveEvents(aggregateID string, events []Event, originalVersion int) (err error) {
	eventsDB := make([]EventDB, len(events))

	for i, event := range events {
		eventsDB[i] = repo.getAggregateDBModel(event, originalVersion+i)
	}

	// Either insert a new aggregate or append to an existing.
	if originalVersion == 0 {
		aggregate := AggregateDB{
			ID:      aggregateID,
			Version: len(eventsDB),
			Events:  eventsDB,
		}

		_, err = repo.eventsCollection.InsertOne(context.Background(), aggregate)
		if err != nil {
			err = fmt.Errorf("cannot insert new aggregate: %w", err)
			return err
		}
	} else {
		query := bson.M{"_id": aggregateID}
		repo.eventsCollection.UpdateOne(
			context.Background(),
			query,
			bson.M{
				"$push": bson.M{"events": bson.M{"$each": eventsDB}},
				"$inc":  bson.M{"version": len(eventsDB)},
			},
		)
		if err != nil {
			err = fmt.Errorf("cannot add event to aggregate: %w", err)
			return err
		}
	}

	return
}
