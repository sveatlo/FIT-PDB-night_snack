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

	err = repo.loadFromEventsStore()
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

	_, err = nc.Subscribe(repo.getTopic(&EventMenuCategoryCreated{}), repo.handleEventMenuCategoryCreated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventMenuCategoryCreated: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.getTopic(&EventMenuCategoryUpdated{}), repo.handleEventMenuCategoryUpdated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventMenuCategoryUpdated: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.getTopic(&EventMenuCategoryDeleted{}), repo.handleEventMenuCategoryDeleted)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventMenuCategoryDeleted: %w", err)
		return
	}

	_, err = nc.Subscribe(repo.getTopic(&EventMenuItemCreated{}), repo.handleEventMenuItemCreated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventMenuItemCreated: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.getTopic(&EventMenuItemUpdated{}), repo.handleEventMenuItemUpdated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventMenuItemUpdated: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.getTopic(&EventMenuItemDeleted{}), repo.handleEventMenuItemDeleted)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventMenuItemDeleted: %w", err)
		return
	}

	return
}

func (repo *ReadRepository) getTopic(event Event) string {
	return fmt.Sprintf("%s.%s", event.EventCategory(), event.EventType())
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

	case *EventMenuCategoryCreated:
		err = repo.applyEventMenuCategoryCreated(e)
	case *EventMenuCategoryUpdated:
		err = repo.applyEventMenuCategoryUpdated(e)
	case *EventMenuCategoryDeleted:
		err = repo.applyEventMenuCategoryDeleted(e)

	case *EventMenuItemCreated:
		err = repo.applyEventMenuItemCreated(e)
	case *EventMenuItemUpdated:
		err = repo.applyEventMenuItemUpdated(e)
	case *EventMenuItemDeleted:
		err = repo.applyEventMenuItemDeleted(e)

	default:
		err = errors.New("event not supported")
	}

	return
}

func (repo *ReadRepository) loadFromEventsStore() (err error) {
	var aggregates []AggregateDB
	{
		var cursor *mongo.Cursor
		cursor, err = repo.eventsCollection.Find(context.Background(), bson.M{})
		if err != nil {
			err = fmt.Errorf("cannot get events from store: %w", err)
			return
		}
		cursor.All(context.Background(), &aggregates)
	}

	repo.log.Trace().Interface("events", aggregates).Msg("loaded events from store")

	for _, aggregate := range aggregates {
		for _, eventDB := range aggregate.Events {
			var event Event

			switch eventDB.Type {
			case "created":
				event = EventCreatedFromData(eventDB.Data)
			case "updated":
				event = EventUpdatedFromData(eventDB.Data)
			case "deleted":
				event = EventDeletedFromData(eventDB.Data)

			case "menucategorycreated":
				event = EventMenuCategoryCreatedFromData(eventDB.Data)
			case "menucategoryupdated":
				event = EventMenuCategoryUpdatedFromData(eventDB.Data)
			case "menucategorydeleted":
				event = EventMenuCategoryDeletedFromData(eventDB.Data)

			case "menuitemcreated":
				event = EventMenuItemCreatedFromData(eventDB.Data)
			case "menuitemupdated":
				event = EventMenuItemUpdatedFromData(eventDB.Data)
			case "menuitemdeleted":
				event = EventMenuItemDeletedFromData(eventDB.Data)
			}

			err = repo.applyEvent(event)
			if err != nil {
				break
			}
		}
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

func (repo *ReadRepository) handleEventMenuCategoryCreated(eventPb *restaurant_pb.MenuCategoryCreated) error {
	return repo.applyEventMenuCategoryCreated(EventMenuCategoryCreatedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventMenuCategoryCreated(event *EventMenuCategoryCreated) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.RestaurantID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.RestaurantID}, r)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	return
}

func (repo *ReadRepository) handleEventMenuCategoryUpdated(eventPb *restaurant_pb.MenuCategoryUpdated) error {
	return repo.applyEventMenuCategoryUpdated(EventMenuCategoryUpdatedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventMenuCategoryUpdated(event *EventMenuCategoryUpdated) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.RestaurantID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.RestaurantID}, r)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	return
}

func (repo *ReadRepository) handleEventMenuCategoryDeleted(eventPb *restaurant_pb.MenuCategoryDeleted) error {
	return repo.applyEventMenuCategoryDeleted(EventMenuCategoryDeletedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventMenuCategoryDeleted(event *EventMenuCategoryDeleted) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.RestaurantID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.RestaurantID}, r)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	return
}

func (repo *ReadRepository) handleEventMenuItemCreated(eventPb *restaurant_pb.MenuItemCreated) error {
	return repo.applyEventMenuItemCreated(EventMenuItemCreatedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventMenuItemCreated(event *EventMenuItemCreated) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.RestaurantID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.RestaurantID}, r)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	return
}

func (repo *ReadRepository) handleEventMenuItemUpdated(eventPb *restaurant_pb.MenuItemUpdated) error {
	return repo.applyEventMenuItemUpdated(EventMenuItemUpdatedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventMenuItemUpdated(event *EventMenuItemUpdated) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.RestaurantID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.RestaurantID}, r)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	return
}

func (repo *ReadRepository) handleEventMenuItemDeleted(eventPb *restaurant_pb.MenuItemDeleted) error {
	return repo.applyEventMenuItemDeleted(EventMenuItemDeletedFromProto(eventPb))
}

func (repo *ReadRepository) applyEventMenuItemDeleted(event *EventMenuItemDeleted) (err error) {
	res := repo.restaurantsCollection.FindOne(context.Background(), bson.M{"_id": event.RestaurantID})
	if res.Err() != nil {
		err = res.Err()
		return
	}

	r := &Restaurant{}
	res.Decode(&r)
	repo.log.Trace().Interface("event", event).Interface("restaurant", r).Msg("pre-apply menuitemdeleted event")
	r.ApplyEvent(event)

	res = repo.restaurantsCollection.FindOneAndReplace(context.Background(), bson.M{"_id": event.RestaurantID}, r)
	if res.Err() != nil {
		err = res.Err()
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
