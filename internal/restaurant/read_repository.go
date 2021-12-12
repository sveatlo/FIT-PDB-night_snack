package restaurant

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type ReadRepository struct {
	log zerolog.Logger
	nc  *nats.EncodedConn
	db  *gorm.DB

	restaurantsCollection *mongo.Collection
	eventsCollection      *mongo.Collection
}

func NewReadRepository(nc *nats.EncodedConn, mongoDB *mongo.Database, log zerolog.Logger) (repo *ReadRepository, err error) {
	repo = &ReadRepository{
		log: log.With().Str("component", "restaurant/read_repository").Logger(),
		nc:  nc,

		restaurantsCollection: mongoDB.Collection("restaurants"),
		eventsCollection:      mongoDB.Collection("events"),
	}

	err = repo.restaurantsCollection.Drop(context.Background())
	if err != nil {
		return
	}

	err = repo.LoadFromEventsStore()
	if err != nil {
		err = fmt.Errorf("cannot retrieve events from event store: %w", err)
		return
	}

	_, err = nc.Subscribe(repo.getTopic(&EventCreated{}), repo.handleEventCreated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventCreated: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.getTopic(&EventUpdated{}), repo.handleEventUpdated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventUpdated: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.getTopic(&EventDeleted{}), repo.handleEventDeleted)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventDeleted: %w", err)
		return
	}

	return
}

func (repo *ReadRepository) LoadFromEventsStore() (err error) {
	var eventsByAggregateID []struct {
		ID     string `bson:"_id" json:"id,omitempty"`
		Events []struct {
			Type string `bson:"type,omitempty" json:"type,omitempty"`
			Data bson.M `bson:"data,omitempty" json:"data,omitempty"`
		} `bson:"events" json:"events,omitempty"`
	}
	{
		groupStage := bson.D{{
			Key: "$group",
			Value: bson.D{
				{Key: "_id", Value: "$aggregateID"},
				{Key: "events", Value: bson.D{{
					Key: "$push", Value: bson.D{
						{Key: "type", Value: "$type"},
						{Key: "data", Value: "$data"},
					}},
				}},
			}},
		}
		var cursor *mongo.Cursor
		cursor, err = repo.eventsCollection.Aggregate(context.Background(), mongo.Pipeline{groupStage})
		if err != nil {
			err = fmt.Errorf("cannot get events from store: %w", err)
			return
		}
		cursor.All(context.Background(), &eventsByAggregateID)
	}

	repo.log.Trace().Interface("events", eventsByAggregateID).Msg("loaded events from store")

	for _, aggregate := range eventsByAggregateID {
		for _, eventRaw := range aggregate.Events {
			var event Event

			switch eventRaw.Type {
			case "created":
				event = EventCreatedFromData(eventRaw.Data)
			case "updated":
				event = EventUpdatedFromData(eventRaw.Data)
			case "deleted":
				repo.log.Warn().Str("event_type", eventRaw.Type).Msg("applying deleted")
				event = EventDeletedFromData(eventRaw.Data)
			}

			err = repo.applyEvent(event)
			if err != nil {
				break
			}
		}
	}

	return
}

func (repo *ReadRepository) getTopic(event Event) string {
	return fmt.Sprintf("%s.%s", event.EventCategory(), event.EventType())
}

func (repo *ReadRepository) getEventModel(event Event) bson.D {
	return bson.D{
		{Key: "category", Value: event.EventCategory()},
		{Key: "type", Value: event.EventType()},
		{Key: "aggregateID", Value: event.AggregateID()},
		{Key: "data", Value: event.Data()},
	}
}

func (repo *ReadRepository) applyEvent(event Event) (err error) {
	repo.log.Trace().
		Str("event", event.EventType()).
		Str("id", event.AggregateID()).
		Interface("data", event.Data()).
		Msg("applying event")
	defer func() {
		if err != nil {
			repo.log.Error().
				Err(err).
				Str("event", event.EventType()).
				Str("id", event.AggregateID()).
				Interface("data", event.Data()).
				Msg("event application failed")
		}
	}()

	switch e := event.(type) {
	case *EventCreated:
		err = repo.applyEventCreated(e)
	case *EventUpdated:
		err = repo.applyEventUpdated(e)
	case *EventDeleted:
		err = repo.applyEventDeleted(e)
	default:
		err = errors.New("event not supported")
	}

	return
}

func (repo *ReadRepository) handleEventCreated(eventPb *restaurant_pb.RestaurantCreated) error {
	return repo.applyEventCreated(EventCreatedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventCreated(event *EventCreated) (err error) {
	r := &Restaurant{}
	r.ApplyEvent(event)

	_, err = repo.restaurantsCollection.InsertOne(context.Background(), r)
	if err != nil {
		return
	}

	return
}

func (repo *ReadRepository) handleEventUpdated(eventPb *restaurant_pb.RestaurantUpdated) error {
	return repo.applyEventUpdated(EventUpdatedFromProto(eventPb))

}

func (repo *ReadRepository) applyEventUpdated(event *EventUpdated) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.ID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.ID}, r)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	return
}

func (repo *ReadRepository) handleEventDeleted(eventPb *restaurant_pb.RestaurantDeleted) error {
	return repo.applyEventDeleted(EventDeletedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventDeleted(event *EventDeleted) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.ID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	_, err = repo.restaurantsCollection.DeleteOne(context.Background(), bson.M{"_id": event.ID})
	if err != nil {
		return
	}

	return
}

func (repo *ReadRepository) GetAll(ctx context.Context) (restaurants []*Restaurant, err error) {
	cur, err := repo.restaurantsCollection.Find(ctx, bson.D{})
	if err != nil {
		err = fmt.Errorf("query failed: %w", err)
		return
	}

	err = cur.All(ctx, &restaurants)
	if err != nil {
		err = fmt.Errorf("cursor decode failed: %w", err)
		return
	}

	return
}

func (repo *ReadRepository) Get(ctx context.Context, id string) (restaurant *Restaurant, err error) {
	res := repo.restaurantsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() != nil {
		err = fmt.Errorf("query failed: %w", res.Err())
		return
	}

	restaurant = &Restaurant{}
	err = res.Decode(restaurant)
	if err != nil {
		err = fmt.Errorf("decode failed: %w", err)
		return
	}

	return
}
