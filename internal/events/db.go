package events

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AggregateDB struct {
	ID       string `bson:"_id"`
	Category string
	Version  int       `bson:"version"`
	Events   []EventDB `bson:"events"`
}
type EventDB struct {
	AggregateID string    `bson:"_id"`
	Timestamp   time.Time `bson:"timestamp"`
	Category    string    `bson:"category"`
	Type        string    `bson:"type"`
	Data        bson.M    `bson:"data"`
}
