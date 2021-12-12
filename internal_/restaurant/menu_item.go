package restaurant

import (
	"time"
)

type MenuItem struct {
	ID             string    `gorm:"primaryKey" bson:"_id"`
	MenuCategoryID string    `gorm:"not null" bson:"restaurant_id"`
	DeletedAt      time.Time `gorm:"null" bson:"deleted_at"`

	Name        string `gorm:"not null;uniqueIndex" bson:"name"`
	Description string `gorm:"null" bson:"description"`
}
