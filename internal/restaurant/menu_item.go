package restaurant

import (
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type MenuItem struct {
	ID             string `gorm:"primaryKey" bson:"_id"`
	MenuCategoryID string `gorm:"not null" bson:"category_id"`

	Name        string `gorm:"not null;uniqueIndex" bson:"name"`
	Description string `gorm:"null" bson:"description"`
}

func NewMenuItemFromProto(mi *restaurant_pb.MenuItem) *MenuItem {
	return &MenuItem{
		ID:          mi.Id,
		Name:        mi.Name,
		Description: mi.Description,
	}
}

func (mi *MenuItem) ToProto() *restaurant_pb.MenuItem {
	return &restaurant_pb.MenuItem{
		Id:          mi.ID,
		Name:        mi.Name,
		Description: mi.Description,
	}
}
