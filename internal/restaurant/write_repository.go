package restaurant

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"github.com/sveatlo/night_snack/internal/events"
	"github.com/sveatlo/night_snack/internal/repository"
)

type WriteRepository struct {
	*repository.Base

	log zerolog.Logger
	nc  *nats.EncodedConn
	db  *gorm.DB
}

func NewWriteRepository(nc *nats.EncodedConn, db *gorm.DB, mongo *mongo.Database, log zerolog.Logger) (repo *WriteRepository, err error) {
	log = log.With().Str("component", "restaurant/write_repository").Logger()
	base, err := repository.NewBase(nc, mongo, log)
	if err != nil {
		return
	}

	repo = &WriteRepository{
		Base: base,

		log: log,
		nc:  nc,
		db:  db,
	}

	err = db.AutoMigrate(&Restaurant{}, &MenuCategory{}, &MenuItem{})
	if err != nil {
		err = fmt.Errorf("migration failed: %w", err)
		return
	}

	return
}

func (repo *WriteRepository) SaveEvents(aggregateID string, aggregateEvents []events.Event, originalVersion int) (err error) {
	return repo.Base.SaveEvents("restaurant", aggregateID, aggregateEvents, originalVersion)
}

func (repo *WriteRepository) LoadAggregate(id string) (aggregate events.AggregateDB, err error) {
	return repo.Base.LoadAggregate("restaurant", id)
}

func (repo *WriteRepository) Create(ctx context.Context, name string) (event *EventCreated, err error) {
	id, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("cannot generate UUID: %w", err)
		return
	}

	aggregate, err := repo.LoadAggregate(id.String())
	if err != nil {
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

		err = repo.SaveEvents(aggregate.ID, []events.Event{event}, aggregate.Version)
		if err != nil {
			return
		}

		err = repo.Publish(event)
		if err != nil {
			return
		}

		return
	})

	return
}

func (repo *WriteRepository) Update(ctx context.Context, id, name string) (event *EventUpdated, err error) {
	aggregate, err := repo.LoadAggregate(id)
	if err != nil {
		return
	}

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

		err = repo.SaveEvents(id, []events.Event{event}, aggregate.Version)
		if err != nil {
			err = fmt.Errorf("cannot create event record: %w", err)
			return
		}

		err = repo.Publish(event)
		if err != nil {
			return
		}

		return
	})

	return
}

func (repo *WriteRepository) Delete(ctx context.Context, id string) (event *EventDeleted, err error) {
	aggregate, err := repo.LoadAggregate(id)
	if err != nil {
		return
	}

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

		err = repo.SaveEvents(id, []events.Event{event}, aggregate.Version)
		if err != nil {
			err = fmt.Errorf("cannot create event record: %w", err)
			return
		}

		err = repo.Publish(event)
		if err != nil {
			return
		}

		return
	})

	return
}

func (repo *WriteRepository) CreateMenuCategory(ctx context.Context, restaurantID, name string) (event *EventMenuCategoryCreated, err error) {
	aggregate, err := repo.LoadAggregate(restaurantID)
	if err != nil {
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("cannot generate UUID: %w", err)
		return
	}
	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Create(&MenuCategory{
			ID:           id.String(),
			RestaurantID: restaurantID,
			Name:         name,
		})
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventMenuCategoryCreated{
			ID:           id.String(),
			RestaurantID: restaurantID,
			Name:         name,
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
	})

	return
}

func (repo *WriteRepository) UpdateMenuCategory(ctx context.Context, id, name string) (event *EventMenuCategoryUpdated, err error) {
	menuCategory := &MenuCategory{}
	res := repo.db.First(&menuCategory, "id = ?", id)
	if res.Error != nil {
		err = fmt.Errorf("cannot find persistent record with such ID: %w", res.Error)
		return
	}
	menuCategory.Name = name

	aggregate, err := repo.LoadAggregate(menuCategory.RestaurantID)
	if err != nil {
		return
	}

	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Save(&menuCategory)
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventMenuCategoryUpdated{
			ID:           id,
			RestaurantID: menuCategory.RestaurantID,
			Name:         name,
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
	})

	return
}

func (repo *WriteRepository) DeleteMenuCategory(ctx context.Context, id string) (event *EventMenuCategoryDeleted, err error) {
	menuCategory := &MenuCategory{}
	res := repo.db.First(&menuCategory, "id = ?", id)
	if res.Error != nil {
		err = fmt.Errorf("cannot find persistent record with such ID: %w", res.Error)
		return
	}

	aggregate, err := repo.LoadAggregate(menuCategory.RestaurantID)
	if err != nil {
		return
	}

	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Delete(&menuCategory)
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventMenuCategoryDeleted{
			ID:           id,
			RestaurantID: menuCategory.RestaurantID,
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
	})

	return
}

func (repo *WriteRepository) CreateMenuItem(ctx context.Context, restaurantID, categoryID, name, description string) (event *EventMenuItemCreated, err error) {
	aggregate, err := repo.LoadAggregate(restaurantID)
	if err != nil {
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("cannot generate UUID: %w", err)
		return
	}
	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Create(&MenuItem{
			ID:             id.String(),
			MenuCategoryID: categoryID,
			Name:           name,
			Description:    description,
		})
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventMenuItemCreated{
			ID:           id.String(),
			RestaurantID: restaurantID,
			CategoryID:   categoryID,
			Name:         name,
			Description:  description,
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
	})

	return
}

func (repo *WriteRepository) UpdateMenuItem(ctx context.Context, restaurantID, categoryID, id, name, description string) (event *EventMenuItemUpdated, err error) {
	menuItem := &MenuItem{}
	res := repo.db.First(&menuItem, "id = ?", id)
	if res.Error != nil {
		err = fmt.Errorf("cannot find persistent record with such ID: %w", res.Error)
		return
	}
	menuItem.Name = name
	menuItem.Description = description

	aggregate, err := repo.LoadAggregate(restaurantID)
	if err != nil {
		return
	}

	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Save(&menuItem)
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventMenuItemUpdated{
			ID:           id,
			RestaurantID: restaurantID,
			CategoryID:   categoryID,
			Name:         name,
			Description:  description,
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
	})

	return
}

func (repo *WriteRepository) DeleteMenuItem(ctx context.Context, restaurantID, id string) (event *EventMenuItemDeleted, err error) {
	menuItem := &MenuItem{}
	res := repo.db.First(&menuItem, "id = ?", id)
	if res.Error != nil {
		err = fmt.Errorf("cannot find persistent record with such ID: %w", res.Error)
		return
	}

	aggregate, err := repo.LoadAggregate(restaurantID)
	if err != nil {
		return
	}

	err = repo.db.Transaction(func(tx *gorm.DB) (err error) {
		res := tx.Delete(&MenuItem{}, "id = ?", id)
		err = res.Error
		if err != nil {
			err = fmt.Errorf("cannot create persistent record: %w", err)
			return
		}

		event = &EventMenuItemDeleted{
			ID:           id,
			RestaurantID: restaurantID,
			CategoryID:   menuItem.MenuCategoryID,
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
	})

	return
}
