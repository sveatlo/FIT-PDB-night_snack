package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/sveatlo/night_snack/internal/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Base struct {
	log zerolog.Logger
	nc  *nats.EncodedConn

	eventsCollection *mongo.Collection
}

func NewBase(nc *nats.EncodedConn, mongoDB *mongo.Database, log zerolog.Logger) (b *Base, err error) {
	b = &Base{
		log:              log,
		nc:               nc,
		eventsCollection: mongoDB.Collection("events"),
	}
	return
}

func (repo *Base) GetTopic(event events.Event) string {
	return fmt.Sprintf("%s.%s", event.EventCategory(), event.EventType())
}

func (repo *Base) Publish(event events.Event) error {
	repo.log.Debug().Interface("proto", event.ToProto()).Str("proto-type", fmt.Sprintf("%T", event.ToProto())).Msg("publishing proto event")
	return repo.nc.Publish(repo.GetTopic(event), event.ToProto())
}

func (repo *Base) getAggregateDBModel(event events.Event, version int) events.EventDB {
	return events.EventDB{
		AggregateID: event.AggregateID(),
		Timestamp:   time.Now(),
		Category:    event.EventCategory(),
		Type:        event.EventType(),
		Data:        event.Data(),
	}
}

func (repo *Base) LoadAggregate(category, id string) (aggregate events.AggregateDB, err error) {
	res := repo.eventsCollection.FindOne(context.Background(), bson.M{"_id": id, "category": category})
	if res.Err() == mongo.ErrNoDocuments {
		aggregate = events.AggregateDB{
			ID:       id,
			Category: category,
			Version:  0,
			Events:   []events.EventDB{},
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

func (repo *Base) SaveEvents(eventCategory, aggregateID string, aggregateEvents []events.Event, originalVersion int) (err error) {
	eventsDB := make([]events.EventDB, len(aggregateEvents))

	for i, event := range aggregateEvents {
		eventsDB[i] = repo.getAggregateDBModel(event, originalVersion+i)
	}

	// Either insert a new aggregate or append to an existing.
	if originalVersion == 0 {
		aggregate := events.AggregateDB{
			ID:       aggregateID,
			Category: eventCategory,
			Version:  len(eventsDB),
			Events:   eventsDB,
		}

		_, err = repo.eventsCollection.InsertOne(context.Background(), aggregate)
		if err != nil {
			err = fmt.Errorf("cannot insert new aggregate: %w", err)
			return err
		}
	} else {
		query := bson.M{"_id": aggregateID, "version": originalVersion}
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
