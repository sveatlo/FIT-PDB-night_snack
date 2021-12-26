package orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sveatlo/night_snack/internal/events"
	"github.com/sveatlo/night_snack/internal/repository"
	"github.com/sveatlo/night_snack/internal/restaurant"
	orders_pb "github.com/sveatlo/night_snack/proto/orders"
)

type Repository struct {
	*repository.Base

	log              zerolog.Logger
	eventsCollection *mongo.Collection
	ordersCollection *mongo.Collection
}

func NewRepository(nc *nats.EncodedConn, mongoDB *mongo.Database, log zerolog.Logger) (repo *Repository, err error) {
	log = log.With().Str("component", "order/repository").Logger()
	base, err := repository.NewBase(nc, mongoDB, log)
	if err != nil {
		return
	}

	repo = &Repository{
		Base: base,

		log:              log,
		ordersCollection: mongoDB.Collection("orders"),
		eventsCollection: mongoDB.Collection("events"),
	}

	err = repo.ordersCollection.Drop(context.Background())
	if err != nil {
		return
	}

	err = repo.loadFromEventsStore()
	if err != nil {
		err = fmt.Errorf("cannot retrieve events from event store: %w", err)
		return
	}

	_, err = nc.Subscribe(repo.GetTopic(&EventOrderCreated{}), repo.handleEventOrderCreated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventStockIncreased: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.GetTopic(&EventStatusUpdated{}), repo.handleEventStatusUpdated)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventStatusUpdated: %w", err)
		return
	}

	return
}

func (repo *Repository) SaveEvents(aggregateID string, aggregateEvents []events.Event, originalVersion int) (err error) {
	return repo.Base.SaveEvents("order", aggregateID, aggregateEvents, originalVersion)
}

func (repo *Repository) LoadAggregate(id string) (aggregate events.AggregateDB, err error) {
	return repo.Base.LoadAggregate("order", id)
}

func (repo *Repository) loadFromEventsStore() (err error) {
	var aggregates []events.AggregateDB
	{
		var cursor *mongo.Cursor
		cursor, err = repo.eventsCollection.Find(context.Background(), bson.M{"category": "order"})
		if err != nil {
			err = fmt.Errorf("cannot get events from store: %w", err)
			return
		}
		cursor.All(context.Background(), &aggregates)
	}

	repo.log.Trace().Interface("events", aggregates).Msg("loaded events from store")

	for _, aggregate := range aggregates {
		for _, eventDB := range aggregate.Events {
			var event events.Event

			switch eventDB.Type {
			case "created":
				event = EventOrderCreatedFromData(eventDB.Data)
			case "statusupdated":
				event = EventStatusUpdatedFromData(eventDB.Data)
			default:
				err = fmt.Errorf("unknown event for order: %v", eventDB.Type)
				return
			}

			err = repo.applyEvent(event)
			if err != nil {
				break
			}
		}
	}

	return
}

func (repo *Repository) applyEvent(event events.Event) (err error) {
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
	case *EventOrderCreated:
		err = repo.applyEventOrderCreated(e)
	case *EventStatusUpdated:
		err = repo.applyEventStatusUpdated(e)

	default:
		err = errors.New("event not supported")
	}

	return
}

func (repo *Repository) CreateOrder(ctx context.Context, restaurant *restaurant.Restaurant, items []*restaurant.MenuItem) (event *EventOrderCreated, err error) {
	id, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("cannot generate UUID: %w", err)
		return
	}

	aggregate, err := repo.LoadAggregate(id.String())
	if err != nil {
		return
	}

	event = &EventOrderCreated{
		ID:         id.String(),
		Restaurant: restaurant,
		Items:      items,
		Status:     orders_pb.OrderStatus_RECEIVED.String(),
	}

	err = repo.SaveEvents(aggregate.ID, []events.Event{event}, aggregate.Version)
	if err != nil {
		return
	}

	err = repo.Publish(event)
	if err != nil {
		return
	}

	return
}

func (repo *Repository) UpdateStatus(ctx context.Context, id string, status *orders_pb.OrderStatus) (event *EventStatusUpdated, err error) {
	aggregate, err := repo.LoadAggregate(id)
	if err != nil {
		return
	}
	if aggregate.Version == 0 {
		err = fmt.Errorf("no such aggregate")
		return
	}

	event = &EventStatusUpdated{
		ID:     id,
		Status: status.String(),
	}

	err = repo.SaveEvents(aggregate.ID, []events.Event{event}, aggregate.Version)
	if err != nil {
		return
	}

	err = repo.Publish(event)
	if err != nil {
		return
	}

	return
}

func (repo *Repository) handleEventOrderCreated(eventPb *orders_pb.OrderCreated) error {
	return repo.applyEventOrderCreated(EventOrderCreatedFromProto(eventPb))
}

func (repo *Repository) applyEventOrderCreated(event *EventOrderCreated) (err error) {
	o := &Order{}
	o.ApplyEvent(event)

	_, err = repo.ordersCollection.InsertOne(context.Background(), o)
	if err != nil {
		return
	}

	return
}

func (repo *Repository) handleEventStatusUpdated(eventPb *orders_pb.StatusUpdated) error {
	return repo.applyEventStatusUpdated(EventStatusUpdatedFromProto(eventPb))
}

func (repo *Repository) applyEventStatusUpdated(event *EventStatusUpdated) (err error) {
	o := &Order{}
	res := repo.ordersCollection.FindOne(context.Background(), bson.M{"_id": event.ID})
	err = res.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return
	} else if err == nil {
		res.Decode(&o)
	}

	o.ApplyEvent(event)

	_, err = repo.ordersCollection.UpdateOne(context.Background(), bson.M{"_id": event.ID}, bson.M{"$set": o}, options.Update().SetUpsert(true))
	if err != nil {
		return
	}
	if err != nil {
		return
	}

	return
}
