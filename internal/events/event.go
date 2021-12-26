package events

import (
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

type Event interface {
	EventCategory() string
	EventType() string
	AggregateID() string
	Data() bson.M
	ToProto() proto.Message
}
