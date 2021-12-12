package restaurant

import (
	"fmt"
	"log"
	"time"

	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type Restaurant struct {
	ID        string    `gorm:"primaryKey" bson:"_id"`
	Name      string    `gorm:"not null;uniqueIndex" bson:"name"`
	DeletedAt time.Time `gorm:"-" bson:"deleted_at"`

	MenuCategories []MenuCategory `bson:"menu_categories"`
}

func NewFromEvents(events []Event) (r *Restaurant, err error) {
	r = &Restaurant{
		MenuCategories: []MenuCategory{},
	}

	for _, event := range events {
		r.ApplyEvent(event)
	}

	return
}

func (r *Restaurant) ApplyEvent(event Event) {
	if r == nil {
		*r = Restaurant{}
	}

	switch e := event.(type) {
	case *EventCreated:
		r.ID = e.ID
		r.Name = e.Name
	case *EventUpdated:
		r.Name = e.Name
	case *EventDeleted:
		r.DeletedAt = e.DeletedAt

	case *EventMenuCategoryCreated:
		r.MenuCategories = append(r.MenuCategories, MenuCategory{
			ID:    e.ID,
			Name:  e.Name,
			Items: []MenuItem{},
		})
	case *EventMenuCategoryUpdated:
		for i, category := range r.MenuCategories {
			if category.ID == e.ID {
				r.MenuCategories[i].Name = e.Name
				break
			}
		}
	case *EventMenuCategoryDeleted:
		var (
			i        int
			category MenuCategory
		)
		for i, category = range r.MenuCategories {
			if category.ID == e.ID {
				break
			}
		}
		r.MenuCategories = append(r.MenuCategories[:i], r.MenuCategories[i+1:]...)

	case *EventMenuItemCreated:
		for i, category := range r.MenuCategories {
			if category.ID == e.CategoryID {
				r.MenuCategories[i].Items = append(r.MenuCategories[i].Items, MenuItem{
					ID:             e.ID,
					MenuCategoryID: e.CategoryID,
					Name:           e.Name,
					Description:    e.Description,
				})
				break
			}
		}
	case *EventMenuItemUpdated:
	outerU:
		for i, category := range r.MenuCategories {
			if category.ID == e.CategoryID {
				for j, item := range category.Items {
					if item.ID == e.ID {
						item.Name = e.Name
						item.Description = e.Description
						r.MenuCategories[i].Items[j] = item
						break outerU
					}
				}
			}
		}
	case *EventMenuItemDeleted:
		log.Printf("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB %+v", r.MenuCategories, e.CategoryID)
	outerD:
		for i, category := range r.MenuCategories {
			if category.ID == e.CategoryID {
				for j, item := range category.Items {
					fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAA 0", item.ID)
					if item.ID == e.ID {
						fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAA 1", len(r.MenuCategories[i].Items))
						r.MenuCategories[i].Items = append(r.MenuCategories[i].Items[:j], r.MenuCategories[i].Items[j+1:]...)
						fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAA 2", len(r.MenuCategories[i].Items))
						break outerD
					}
				}
			}
		}
	}
}

func (r *Restaurant) ToProto() *restaurant_pb.Restaurant {
	categories := make([]*restaurant_pb.MenuCategory, len(r.MenuCategories))
	for i, c := range r.MenuCategories {
		categories[i] = c.ToProto()
	}

	return &restaurant_pb.Restaurant{
		Id:         r.ID,
		Name:       r.Name,
		Categories: categories,
	}
}
