package restaurant

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	bson_primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

var (
	_ Event = &EventCreated{}
	_ Event = &EventUpdated{}
	_ Event = &EventDeleted{}
)

type Event interface {
	EventCategory() string
	EventType() string
	AggregateID() string
	Data() bson.D
	ToProto() proto.Message
}

type EventCreated struct {
	ID   string `bson:"id,omitempty" json:"id,omitempty"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}

func EventCreatedFromProto(cmd *restaurant_pb.RestaurantCreated) *EventCreated {
	return &EventCreated{
		ID:   cmd.GetId(),
		Name: cmd.GetName(),
	}
}

func EventCreatedFromData(data bson.M) *EventCreated {
	return &EventCreated{
		ID:   data["id"].(string),
		Name: data["name"].(string),
	}
}

func (e *EventCreated) EventCategory() string { return "restaurant" }
func (e *EventCreated) EventType() string     { return "created" }
func (e *EventCreated) AggregateID() string   { return e.ID }
func (e *EventCreated) Data() bson.D {
	return bson.D{
		{Key: "id", Value: e.ID},
		{Key: "name", Value: e.Name},
	}
}
func (e *EventCreated) ToProto() proto.Message {
	return &restaurant_pb.RestaurantCreated{
		Id:   e.ID,
		Name: e.Name,
	}
}

type EventUpdated struct {
	ID   string `bson:"id,omitempty" json:"id,omitempty"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}

func EventUpdatedFromProto(cmd *restaurant_pb.RestaurantUpdated) *EventUpdated {
	return &EventUpdated{
		ID:   cmd.GetId(),
		Name: cmd.GetName(),
	}
}

func EventUpdatedFromData(data bson.M) *EventUpdated {
	return &EventUpdated{
		ID:   data["id"].(string),
		Name: data["name"].(string),
	}
}

func (e *EventUpdated) EventCategory() string { return "restaurant" }
func (e *EventUpdated) EventType() string     { return "updated" }
func (e *EventUpdated) AggregateID() string   { return e.ID }
func (e *EventUpdated) Data() bson.D {
	return bson.D{
		{Key: "id", Value: e.ID},
		{Key: "name", Value: e.Name},
	}
}
func (e *EventUpdated) ToProto() proto.Message {
	return &restaurant_pb.RestaurantUpdated{
		Id:   e.ID,
		Name: e.Name,
	}
}

type EventDeleted struct {
	ID        string    `bson:"id,omitempty" json:"id,omitempty"`
	DeletedAt time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func EventDeletedFromProto(cmd *restaurant_pb.RestaurantDeleted) *EventDeleted {
	return &EventDeleted{
		ID:        cmd.GetId(),
		DeletedAt: cmd.GetDeletedAt().AsTime(),
	}
}

func EventDeletedFromData(data bson.M) *EventDeleted {
	return &EventDeleted{
		ID:        data["id"].(string),
		DeletedAt: (data["deleted_at"].(bson_primitive.DateTime)).Time(),
	}
}

func (e *EventDeleted) EventCategory() string { return "restaurant" }
func (e *EventDeleted) EventType() string     { return "deleted" }
func (e *EventDeleted) AggregateID() string   { return e.ID }
func (e *EventDeleted) Data() bson.D {
	return bson.D{
		{Key: "id", Value: e.ID},
		{Key: "deleted_at", Value: e.DeletedAt},
	}
}
func (e *EventDeleted) ToProto() proto.Message {
	return &restaurant_pb.RestaurantDeleted{
		Id:        e.ID,
		DeletedAt: timestamppb.New(e.DeletedAt),
	}
}
