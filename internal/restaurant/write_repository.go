package restaurant

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type WriteRepository struct {
	log   zerolog.Logger
	nc    *nats.EncodedConn
	db    *gorm.DB
	mongo *mongo.Collection
}

func NewWriteRepository(nc *nats.EncodedConn, db *gorm.DB, mongo *mongo.Database, log zerolog.Logger) (repo *WriteRepository, err error) {
	repo = &WriteRepository{
		log:   log.With().Str("component", "restaurant/write_repository").Logger(),
		nc:    nc,
		db:    db,
		mongo: mongo.Collection("events"),
	}

	err = db.AutoMigrate(&Restaurant{}, &MenuCategory{}, &MenuItem{})
	if err != nil {
		err = fmt.Errorf("migration failed: %w", err)
		return
	}

	// _, err = nc.Subscribe(repo.getTopic(&EventCreated{}), repo.handleCreateCommand)
	// if err != nil {
	//     err = fmt.Errorf("cannot create subscription for EventCreated: %w", err)
	// }

	return
}

func (repo *WriteRepository) getTopic(event Event) string {
	return fmt.Sprintf("%s.%s", event.EventCategory(), event.EventType())
}

func (repo *WriteRepository) getEventModel(event Event) bson.D {
	return bson.D{
		{Key: "timestamp", Value: time.Now()},
		{Key: "category", Value: event.EventCategory()},
		{Key: "type", Value: event.EventType()},
		{Key: "aggregateID", Value: event.AggregateID()},
		{Key: "data", Value: event.Data()},
	}
}

func (repo *WriteRepository) publish(event Event) error {
	repo.log.Debug().Interface("proto", event.ToProto()).Str("proto-type", fmt.Sprintf("%T", event.ToProto())).Msg("publishing proto event")
	return repo.nc.Publish(repo.getTopic(event), event.ToProto())
}

func (repo *WriteRepository) Create(ctx context.Context, name string) (event *EventCreated, err error) {
	id, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("cannot generate UUID: %w", err)
		return
	}

	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Create(&Restaurant{
			ID:   id.String(),
			Name: name,
		})
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventCreated{
			ID:   id.String(),
			Name: name,
		}

		_, err = repo.mongo.InsertOne(ctx, repo.getEventModel(event))
		if err != nil {
			err = fmt.Errorf("cannot create event record: %w", err)
			return
		}

		err = repo.publish(event)
		if err != nil {
			return
		}

		return
	})

	return
}

func (repo *WriteRepository) Update(ctx context.Context, id, name string) (event *EventUpdated, err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.First(&Restaurant{}, "id = ?", id)
		if res.Error != nil {
			err = fmt.Errorf("cannot find persistent record with such ID: %w", res.Error)
			return
		}

		res = tx.Save(&Restaurant{
			ID:   id,
			Name: name,
		})
		if res.Error != nil {
			err = fmt.Errorf("cannot create persistent record: %w", res.Error)
			return
		}

		event = &EventUpdated{
			ID:   id,
			Name: name,
		}

		_, err = repo.mongo.InsertOne(ctx, repo.getEventModel(event))
		if err != nil {
			err = fmt.Errorf("cannot create event record: %w", err)
			return
		}

		err = repo.publish(event)
		if err != nil {
			return
		}

		return
	})

	return
}

func (repo *WriteRepository) Delete(ctx context.Context, id string) (event *EventDeleted, err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.First(&Restaurant{}, "id = ?", id)
		if res.Error != nil {
			err = fmt.Errorf("cannot find persistent record with such ID: %w", res.Error)
			return
		}

		res = tx.Delete(&Restaurant{
			ID: id,
		})
		if res.Error != nil {
			err = fmt.Errorf("cannot create persistent record: %w", res.Error)
			return
		}

		event = &EventDeleted{
			ID:        id,
			DeletedAt: time.Now(),
		}

		_, err = repo.mongo.InsertOne(ctx, repo.getEventModel(event))
		if err != nil {
			err = fmt.Errorf("cannot create event record: %w", err)
			return
		}

		err = repo.publish(event)
		if err != nil {
			return
		}

		return
	})

	return
}

//
// func (repo *Repository) handleCreateCommand(cmd *restaurant_pb.RestaurantCreated) error {
//     event := EventCreatedFromProto(cmd)
//     repo.log.Debug().Interface("data", event).Msg("create event received")
//
//     return nil
// }
