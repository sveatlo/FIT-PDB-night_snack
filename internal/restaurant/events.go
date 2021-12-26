package restaurant

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	bson_primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"

	"github.com/sveatlo/night_snack/internal/events"
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

var (
	_ events.Event = &EventCreated{}
	_ events.Event = &EventUpdated{}
	_ events.Event = &EventDeleted{}

	_ events.Event = &EventMenuCategoryCreated{}
	_ events.Event = &EventMenuCategoryUpdated{}
	_ events.Event = &EventMenuCategoryDeleted{}

	_ events.Event = &EventMenuItemCreated{}
	_ events.Event = &EventMenuItemUpdated{}
	_ events.Event = &EventMenuItemDeleted{}
)

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
func (e *EventCreated) Data() bson.M {
	return bson.M{
		"id":   e.ID,
		"name": e.Name,
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
func (e *EventUpdated) Data() bson.M {
	return bson.M{
		"id":   e.ID,
		"name": e.Name,
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
		ID: cmd.GetId(),
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
func (e *EventDeleted) Data() bson.M {
	return bson.M{
		"id":         e.ID,
		"deleted_at": e.DeletedAt,
	}
}
func (e *EventDeleted) ToProto() proto.Message {
	return &restaurant_pb.RestaurantDeleted{
		Id: e.ID,
	}
}

type EventMenuCategoryCreated struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	RestaurantID string `bson:"restaurant_id,omitempty" json:"restaurant_id,omitempty"`
	Name         string `bson:"name,omitempty" json:"name,omitempty"`
}

func EventMenuCategoryCreatedFromProto(cmd *restaurant_pb.MenuCategoryCreated) *EventMenuCategoryCreated {
	return &EventMenuCategoryCreated{
		ID:           cmd.GetId(),
		RestaurantID: cmd.GetRestaurantId(),
		Name:         cmd.GetName(),
	}
}

func EventMenuCategoryCreatedFromData(data bson.M) *EventMenuCategoryCreated {
	return &EventMenuCategoryCreated{
		ID:           data["id"].(string),
		RestaurantID: data["restaurant_id"].(string),
		Name:         data["name"].(string),
	}
}

func (e *EventMenuCategoryCreated) EventCategory() string { return "restaurant" }
func (e *EventMenuCategoryCreated) EventType() string     { return "menucategorycreated" }
func (e *EventMenuCategoryCreated) AggregateID() string   { return e.RestaurantID }
func (e *EventMenuCategoryCreated) Data() bson.M {
	return bson.M{
		"id":            e.ID,
		"restaurant_id": e.RestaurantID,
		"name":          e.Name,
	}
}
func (e *EventMenuCategoryCreated) ToProto() proto.Message {
	return &restaurant_pb.MenuCategoryCreated{
		Id:           e.ID,
		RestaurantId: e.RestaurantID,
		Name:         e.Name,
	}
}

type EventMenuCategoryUpdated struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	RestaurantID string `bson:"restaurant_id,omitempty" json:"restaurant_id,omitempty"`
	Name         string `bson:"name,omitempty" json:"name,omitempty"`
}

func EventMenuCategoryUpdatedFromProto(cmd *restaurant_pb.MenuCategoryUpdated) *EventMenuCategoryUpdated {
	return &EventMenuCategoryUpdated{
		ID:           cmd.GetId(),
		RestaurantID: cmd.GetRestaurantId(),
		Name:         cmd.GetName(),
	}
}

func EventMenuCategoryUpdatedFromData(data bson.M) *EventMenuCategoryUpdated {
	return &EventMenuCategoryUpdated{
		ID:           data["id"].(string),
		RestaurantID: data["restaurant_id"].(string),
		Name:         data["name"].(string),
	}
}

func (e *EventMenuCategoryUpdated) EventCategory() string { return "restaurant" }
func (e *EventMenuCategoryUpdated) EventType() string     { return "menucategoryupdated" }
func (e *EventMenuCategoryUpdated) AggregateID() string   { return e.RestaurantID }
func (e *EventMenuCategoryUpdated) Data() bson.M {
	return bson.M{
		"id":            e.ID,
		"restaurant_id": e.RestaurantID,
		"name":          e.Name,
	}
}
func (e *EventMenuCategoryUpdated) ToProto() proto.Message {
	return &restaurant_pb.MenuCategoryUpdated{
		Id:           e.ID,
		RestaurantId: e.RestaurantID,
		Name:         e.Name,
	}
}

type EventMenuCategoryDeleted struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	RestaurantID string `bson:"restaurant_id,omitempty" json:"restaurant_id,omitempty"`
}

func EventMenuCategoryDeletedFromProto(cmd *restaurant_pb.MenuCategoryDeleted) *EventMenuCategoryDeleted {
	return &EventMenuCategoryDeleted{
		ID:           cmd.GetId(),
		RestaurantID: cmd.GetRestaurantId(),
	}
}

func EventMenuCategoryDeletedFromData(data bson.M) *EventMenuCategoryDeleted {
	return &EventMenuCategoryDeleted{
		ID:           data["id"].(string),
		RestaurantID: data["restaurant_id"].(string),
	}
}

func (e *EventMenuCategoryDeleted) EventCategory() string { return "restaurant" }
func (e *EventMenuCategoryDeleted) EventType() string     { return "menucategorydeleted" }
func (e *EventMenuCategoryDeleted) AggregateID() string   { return e.RestaurantID }
func (e *EventMenuCategoryDeleted) Data() bson.M {
	return bson.M{
		"id":            e.ID,
		"restaurant_id": e.RestaurantID,
	}
}
func (e *EventMenuCategoryDeleted) ToProto() proto.Message {
	return &restaurant_pb.MenuCategoryDeleted{
		Id:           e.ID,
		RestaurantId: e.RestaurantID,
	}
}

type EventMenuItemCreated struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	RestaurantID string `bson:"restaurant_id,omitempty" json:"restaurant_id,omitempty"`
	CategoryID   string `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name         string `bson:"name,omitempty" json:"name,omitempty"`
	Description  string `bson:"description,omitempty" json:"description,omitempty"`
}

func EventMenuItemCreatedFromProto(cmd *restaurant_pb.MenuItemCreated) *EventMenuItemCreated {
	return &EventMenuItemCreated{
		ID:           cmd.GetId(),
		RestaurantID: cmd.GetRestaurantId(),
		CategoryID:   cmd.GetCategoryId(),
		Name:         cmd.GetName(),
		Description:  cmd.GetDescription(),
	}
}

func EventMenuItemCreatedFromData(data bson.M) *EventMenuItemCreated {
	return &EventMenuItemCreated{
		ID:           data["id"].(string),
		RestaurantID: data["restaurant_id"].(string),
		CategoryID:   data["category_id"].(string),
		Name:         data["name"].(string),
		Description:  data["description"].(string),
	}
}

func (e *EventMenuItemCreated) EventCategory() string { return "restaurant" }
func (e *EventMenuItemCreated) EventType() string     { return "menuitemcreated" }
func (e *EventMenuItemCreated) AggregateID() string   { return e.RestaurantID }
func (e *EventMenuItemCreated) Data() bson.M {
	return bson.M{
		"id":            e.ID,
		"restaurant_id": e.RestaurantID,
		"category_id":   e.CategoryID,
		"name":          e.Name,
		"description":   e.Description,
	}
}
func (e *EventMenuItemCreated) ToProto() proto.Message {
	return &restaurant_pb.MenuItemCreated{
		Id:           e.ID,
		RestaurantId: e.RestaurantID,
		CategoryId:   e.CategoryID,
		Name:         e.Name,
		Description:  e.Description,
	}
}

type EventMenuItemUpdated struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	RestaurantID string `bson:"restaurant_id,omitempty" json:"restaurant_id,omitempty"`
	CategoryID   string `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name         string `bson:"name,omitempty" json:"name,omitempty"`
	Description  string `bson:"description,omitempty" json:"description,omitempty"`
}

func EventMenuItemUpdatedFromProto(cmd *restaurant_pb.MenuItemUpdated) *EventMenuItemUpdated {
	return &EventMenuItemUpdated{
		ID:           cmd.GetId(),
		RestaurantID: cmd.GetRestaurantId(),
		CategoryID:   cmd.GetCategoryId(),
		Name:         cmd.GetName(),
		Description:  cmd.GetDescription(),
	}
}

func EventMenuItemUpdatedFromData(data bson.M) *EventMenuItemUpdated {
	return &EventMenuItemUpdated{
		ID:           data["id"].(string),
		RestaurantID: data["restaurant_id"].(string),
		CategoryID:   data["category_id"].(string),
		Name:         data["name"].(string),
		Description:  data["description"].(string),
	}
}

func (e *EventMenuItemUpdated) EventCategory() string { return "restaurant" }
func (e *EventMenuItemUpdated) EventType() string     { return "menuitemupdated" }
func (e *EventMenuItemUpdated) AggregateID() string   { return e.RestaurantID }
func (e *EventMenuItemUpdated) Data() bson.M {
	return bson.M{
		"id":            e.ID,
		"restaurant_id": e.RestaurantID,
		"category_id":   e.CategoryID,
		"name":          e.Name,
		"description":   e.Description,
	}
}
func (e *EventMenuItemUpdated) ToProto() proto.Message {
	return &restaurant_pb.MenuItemUpdated{
		Id:           e.ID,
		RestaurantId: e.RestaurantID,
		CategoryId:   e.CategoryID,
		Name:         e.Name,
		Description:  e.Description,
	}
}

type EventMenuItemDeleted struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	RestaurantID string `bson:"restaurant_id,omitempty" json:"restaurant_id,omitempty"`
	CategoryID   string `bson:"category_id,omitempty" json:"category_id,omitempty"`
}

func EventMenuItemDeletedFromProto(cmd *restaurant_pb.MenuItemDeleted) *EventMenuItemDeleted {
	return &EventMenuItemDeleted{
		ID:           cmd.GetId(),
		RestaurantID: cmd.GetRestaurantId(),
		CategoryID:   cmd.GetCategoryId(),
	}
}

func EventMenuItemDeletedFromData(data bson.M) *EventMenuItemDeleted {
	return &EventMenuItemDeleted{
		ID:           data["id"].(string),
		RestaurantID: data["restaurant_id"].(string),
		CategoryID:   data["category_id"].(string),
	}
}

func (e *EventMenuItemDeleted) EventCategory() string { return "restaurant" }
func (e *EventMenuItemDeleted) EventType() string     { return "menuitemdeleted" }
func (e *EventMenuItemDeleted) AggregateID() string   { return e.RestaurantID }
func (e *EventMenuItemDeleted) Data() bson.M {
	return bson.M{
		"id":            e.ID,
		"restaurant_id": e.RestaurantID,
		"category_id":   e.CategoryID,
	}
}
func (e *EventMenuItemDeleted) ToProto() proto.Message {
	return &restaurant_pb.MenuItemDeleted{
		Id:           e.ID,
		RestaurantId: e.RestaurantID,
		CategoryId:   e.CategoryID,
	}
}
