package orders

import (
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"

	"github.com/sveatlo/night_snack/internal/events"
	"github.com/sveatlo/night_snack/internal/restaurant"
	orders_pb "github.com/sveatlo/night_snack/proto/orders"
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

var (
	_ events.Event = &EventOrderCreated{}
)

type EventOrderCreated struct {
	ID         string
	Restaurant *restaurant.Restaurant
	Items      []*restaurant.MenuItem
	Status     string
}

func EventOrderCreatedFromProto(cmd *orders_pb.OrderCreated) *EventOrderCreated {
	items := make([]*restaurant.MenuItem, len(cmd.Items))
	for i, item := range cmd.Items {
		items[i] = restaurant.NewMenuItemFromProto(item)
	}

	return &EventOrderCreated{
		ID:         cmd.Id,
		Status:     cmd.Status.String(),
		Restaurant: restaurant.NewRestaurantFromProto(cmd.Restaurant),
		Items:      items,
	}
}

func EventOrderCreatedFromData(data bson.M) *EventOrderCreated {
	rData := data["restaurant"].(bson.M)
	r := &restaurant.Restaurant{
		ID:   rData["_id"].(string),
		Name: rData["name"].(string),
	}
	items := []*restaurant.MenuItem{}
	itemsData := data["items"].(bson.A)
	for _, itemDataR := range itemsData {
		itemData := itemDataR.(bson.M)
		item := &restaurant.MenuItem{
			ID:             itemData["_id"].(string),
			MenuCategoryID: itemData["category_id"].(string),
			Name:           itemData["name"].(string),
			Description:    itemData["description"].(string),
		}
		items = append(items, item)
	}
	return &EventOrderCreated{
		ID:         data["id"].(string),
		Status:     data["status"].(string),
		Restaurant: r,
		Items:      items,
	}
}

func (e *EventOrderCreated) EventCategory() string { return "order" }
func (e *EventOrderCreated) EventType() string     { return "created" }
func (e *EventOrderCreated) AggregateID() string   { return e.ID }
func (e *EventOrderCreated) Data() bson.M {
	return bson.M{
		"id":         e.ID,
		"status":     e.Status,
		"restaurant": e.Restaurant,
		"items":      e.Items,
	}
}
func (e *EventOrderCreated) ToProto() proto.Message {
	items := make([]*restaurant_pb.MenuItem, len(e.Items))
	for i, mi := range e.Items {
		items[i] = mi.ToProto()
	}

	return &orders_pb.OrderCreated{
		Id:         e.ID,
		Restaurant: e.Restaurant.ToProto(),
		Items:      items,
		Status:     orders_pb.OrderStatus(orders_pb.OrderStatus_value[e.Status]),
	}
}

type EventStatusUpdated struct {
	ID     string
	Status string
}

func EventStatusUpdatedFromProto(cmd *orders_pb.StatusUpdated) *EventStatusUpdated {
	return &EventStatusUpdated{
		ID:     cmd.Id,
		Status: cmd.Status.String(),
	}
}

func EventStatusUpdatedFromData(data bson.M) *EventStatusUpdated {
	return &EventStatusUpdated{
		ID:     data["id"].(string),
		Status: data["status"].(string),
	}
}

func (e *EventStatusUpdated) EventCategory() string { return "order" }
func (e *EventStatusUpdated) EventType() string     { return "statusupdated" }
func (e *EventStatusUpdated) AggregateID() string   { return e.ID }
func (e *EventStatusUpdated) Data() bson.M {
	return bson.M{
		"id":     e.ID,
		"status": e.Status,
	}
}
func (e *EventStatusUpdated) ToProto() proto.Message {
	return &orders_pb.StatusUpdated{
		Id:     e.ID,
		Status: orders_pb.OrderStatus(orders_pb.OrderStatus_value[e.Status]),
	}
}
