package stock

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sveatlo/night_snack/internal/events"
	"github.com/sveatlo/night_snack/internal/repository"
	stock_pb "github.com/sveatlo/night_snack/proto/stock"
)

type Repository struct {
	*repository.Base

	log              zerolog.Logger
	eventsCollection *mongo.Collection
	stockCollection  *mongo.Collection
}

func NewRepository(nc *nats.EncodedConn, mongoDB *mongo.Database, log zerolog.Logger) (repo *Repository, err error) {
	log = log.With().Str("component", "stock/repository").Logger()
	base, err := repository.NewBase(nc, mongoDB, log)
	if err != nil {
		return
	}

	repo = &Repository{
		Base: base,

		log:              log,
		stockCollection:  mongoDB.Collection("stock"),
		eventsCollection: mongoDB.Collection("events"),
	}

	err = repo.stockCollection.Drop(context.Background())
	if err != nil {
		return
	}

	err = repo.loadFromEventsStore()
	if err != nil {
		err = fmt.Errorf("cannot retrieve events from event store: %w", err)
		return
	}

	_, err = nc.Subscribe(repo.GetTopic(&EventStockIncreased{}), repo.handleEventStockIncreased)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventStockIncreased: %w", err)
		return
	}
	_, err = nc.Subscribe(repo.GetTopic(&EventStockDecreased{}), repo.handleEventStockDecreased)
	if err != nil {
		err = fmt.Errorf("cannot create subscription for EventStockDecreased: %w", err)
		return
	}

	return
}

func (repo *Repository) SaveEvents(aggregateID string, aggregateEvents []events.Event, originalVersion int) (err error) {
	return repo.Base.SaveEvents("stock", aggregateID, aggregateEvents, originalVersion)
}

func (repo *Repository) LoadAggregate(id string) (aggregate events.AggregateDB, err error) {
	return repo.Base.LoadAggregate("stock", id)
}

func (repo *Repository) IncreaseStock(ctx context.Context, itemID string, n int32) (event *EventStockIncreased, err error) {
	aggregate, err := repo.LoadAggregate(itemID)
	if err != nil {
		return
	}

	event = &EventStockIncreased{
		ItemID: itemID,
		N:      n,
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

func (repo *Repository) DecreaseStock(ctx context.Context, itemID string, n int32) (event *EventStockDecreased, err error) {
	aggregate, err := repo.LoadAggregate(itemID)
	if err != nil {
		return
	}

	res := repo.stockCollection.FindOne(ctx, bson.M{"_id": itemID})
	err = res.Err()
	if err != nil {
		return
	}
	stock := &Stock{}
	res.Decode(stock)
	if stock.N < n {
		err = fmt.Errorf("not enough in stock")
		return
	}

	event = &EventStockDecreased{
		ItemID: itemID,
		N:      n,
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

func (repo *Repository) loadFromEventsStore() (err error) {
	var aggregates []events.AggregateDB
	{
		var cursor *mongo.Cursor
		cursor, err = repo.eventsCollection.Find(context.Background(), bson.M{"category": "stock"})
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
			case "increased":
				event = EventStockIncreasedFromData(eventDB.Data)
			case "decreased":
				event = EventStockDecreasedFromData(eventDB.Data)
			default:
				err = fmt.Errorf("unknown event for stock: %v", eventDB.Type)
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
	case *EventStockIncreased:
		err = repo.applyEventStockIncreased(e)
	case *EventStockDecreased:
		err = repo.applyEventStockDecreased(e)

	default:
		err = errors.New("event not supported")
	}

	return
}

func (repo *Repository) handleEventStockIncreased(eventPb *stock_pb.StockIncreased) error {
	return repo.applyEventStockIncreased(EventStockIncreasedFromProto(eventPb))
}

func (repo *Repository) applyEventStockIncreased(event *EventStockIncreased) (err error) {
	s := &Stock{}
	res := repo.stockCollection.FindOne(context.Background(), bson.M{"_id": event.ItemID})
	err = res.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return
	} else if err == nil {
		res.Decode(&s)
	}

	s.ApplyEvent(event)

	_, err = repo.stockCollection.UpdateOne(context.Background(), bson.M{"_id": event.ItemID}, bson.M{"$set": s}, options.Update().SetUpsert(true))
	if err != nil {
		return
	}

	return
}

func (repo *Repository) handleEventStockDecreased(eventPb *stock_pb.StockDecreased) error {
	return repo.applyEventStockDecreased(EventStockDecreasedFromProto(eventPb))
}

func (repo *Repository) applyEventStockDecreased(event *EventStockDecreased) (err error) {
	s := &Stock{}
	res := repo.stockCollection.FindOne(context.Background(), bson.M{"_id": event.ItemID})
	err = res.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return
	} else if err == nil {
		res.Decode(&s)
	}

	s.ApplyEvent(event)

	_, err = repo.stockCollection.UpdateOne(context.Background(), bson.M{"_id": event.ItemID}, bson.M{"$set": s}, options.Update().SetUpsert(true))
	if err != nil {
		return
	}
	if err != nil {
		return
	}

	return
}
