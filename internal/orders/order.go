package orders

import (
	"github.com/sveatlo/night_snack/internal/events"
	"github.com/sveatlo/night_snack/internal/restaurant"
)

type Order struct {
	ID         string                 `bson:"_id"`
	Status     string                 `bson:"status"`
	Restaurant *restaurant.Restaurant `bson:"restaurant"`
	Items      []*restaurant.MenuItem `bson:"items"`
}

func NewFromEvents(events []events.Event) (s *Order) {
	s = &Order{}

	for _, event := range events {
		s.ApplyEvent(event)
	}

	return
}

func (s *Order) ApplyEvent(event events.Event) {
	switch e := event.(type) {
	case *EventOrderCreated:
		s.ID = e.ID
		s.Status = e.Status
		s.Restaurant = e.Restaurant
		s.Items = e.Items
	case *EventStatusUpdated:
		s.Status = e.Status
	}
}
