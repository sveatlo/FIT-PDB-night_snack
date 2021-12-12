package restaurant

import (
	"time"
)

type MenuCategory struct {
	ID           string    `gorm:"primaryKey" bson:"_id"`
	RestaurantID string    `gorm:"not null" bson:"restaurant_id"`
	DeletedAt    time.Time `gorm:"null" bson:"deleted_at"`

	Name string `gorm:"not null;uniqueIndex" bson:"name"`

	Items []MenuItem `bson:"items"`
}
