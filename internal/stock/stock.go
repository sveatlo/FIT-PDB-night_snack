package stock

import (
	"github.com/sveatlo/night_snack/internal/events"
)

type Stock struct {
	ItemID string `bson:"_id,omitempty" json:"item_id,omitempty"`
	N      int32  `bson:"n" json:"n"`
}

func NewFromEvents(events []events.Event) (s *Stock) {
	s = &Stock{
		ItemID: "",
		N:      0,
	}

	for _, event := range events {
		s.ApplyEvent(event)
	}

	return
}

func (s *Stock) ApplyEvent(event events.Event) {
	switch e := event.(type) {
	case *EventStockIncreased:
		s.ItemID = e.ItemID
		s.N += e.N
	case *EventStockDecreased:
		s.ItemID = e.ItemID
		s.N -= e.N
	}
}
