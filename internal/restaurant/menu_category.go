package restaurant

import (
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type MenuCategory struct {
	ID           string `gorm:"primaryKey" bson:"_id"`
	RestaurantID string `gorm:"not null" bson:"restaurant_id"`

	Name string `gorm:"not null;uniqueIndex" bson:"name"`

	Items []MenuItem `bson:"items"`
}

func (mc *MenuCategory) ToProto() *restaurant_pb.MenuCategory {
	items := make([]*restaurant_pb.MenuItem, len(mc.Items))
	for i, mi := range mc.Items {
		items[i] = mi.ToProto()
	}

	return &restaurant_pb.MenuCategory{
		Id:   mc.ID,
		Name: mc.Name,

		Items: items,
	}
}
