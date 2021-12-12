package restaurant

import (
	"time"

	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type Restaurant struct {
	ID        string    `gorm:"primaryKey" bson:"_id"`
	Name      string    `gorm:"not null;uniqueIndex" bson:"name"`
	DeletedAt time.Time `gorm:"null" bson:"deleted_at"`

	MenuCategories []MenuCategory `bson:"menu_categories"`
}

func NewFromEvents(events []Event) (r *Restaurant, err error) {
	r = &Restaurant{}

	for _, event := range events {
		r.ApplyEvent(event)
	}

	return
}

func (r *Restaurant) ApplyEvent(event Event) {
	if r == nil {
		*r = Restaurant{}
	}

	switch e := event.(type) {
	case *EventCreated:
		r.ID = e.ID
		r.Name = e.Name
	case *EventUpdated:
		r.Name = e.Name
	case *EventDeleted:
		r.DeletedAt = e.DeletedAt
	}
}

func (r *Restaurant) ToProto() *restaurant_pb.Restaurant {
	return &restaurant_pb.Restaurant{
		Id:   r.ID,
		Name: r.Name,
	}
}
