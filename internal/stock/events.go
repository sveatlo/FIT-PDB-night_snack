package stock

import (
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"

	"github.com/sveatlo/night_snack/internal/events"
	stock_pb "github.com/sveatlo/night_snack/proto/stock"
)

var (
	_ events.Event = &EventStockIncreased{}
	_ events.Event = &EventStockDecreased{}
)

type EventStockIncreased struct {
	ItemID string `bson:"item_id,omitempty" json:"item_id,omitempty"`
	N      int32  `bson:"n,omitempty" json:"n,omitempty"`
}

func EventStockIncreasedFromProto(cmd *stock_pb.StockIncreased) *EventStockIncreased {
	return &EventStockIncreased{
		ItemID: cmd.GetItemId(),
		N:      cmd.GetN(),
	}
}

func EventStockIncreasedFromData(data bson.M) *EventStockIncreased {
	return &EventStockIncreased{
		ItemID: data["item_id"].(string),
		N:      data["n"].(int32),
	}
}

func (e *EventStockIncreased) EventCategory() string { return "stock" }
func (e *EventStockIncreased) EventType() string     { return "increased" }
func (e *EventStockIncreased) AggregateID() string   { return e.ItemID }
func (e *EventStockIncreased) Data() bson.M {
	return bson.M{
		"item_id": e.ItemID,
		"n":       e.N,
	}
}
func (e *EventStockIncreased) ToProto() proto.Message {
	return &stock_pb.StockIncreased{
		ItemId: e.ItemID,
		N:      e.N,
	}
}

type EventStockDecreased struct {
	ItemID string `bson:"item_id,omitempty" json:"item_id,omitempty"`
	N      int32  `bson:"n,omitempty" json:"n,omitempty"`
}

func EventStockDecreasedFromProto(cmd *stock_pb.StockDecreased) *EventStockDecreased {
	return &EventStockDecreased{
		ItemID: cmd.GetItemId(),
		N:      cmd.GetN(),
	}
}

func EventStockDecreasedFromData(data bson.M) *EventStockDecreased {
	return &EventStockDecreased{
		ItemID: data["item_id"].(string),
		N:      data["n"].(int32),
	}
}

func (e *EventStockDecreased) EventCategory() string { return "stock" }
func (e *EventStockDecreased) EventType() string     { return "decreased" }
func (e *EventStockDecreased) AggregateID() string   { return e.ItemID }
func (e *EventStockDecreased) Data() bson.M {
	return bson.M{
		"item_id": e.ItemID,
		"n":       e.N,
	}
}
func (e *EventStockDecreased) ToProto() proto.Message {
	return &stock_pb.StockDecreased{
		ItemId: e.ItemID,
		N:      e.N,
	}
}
