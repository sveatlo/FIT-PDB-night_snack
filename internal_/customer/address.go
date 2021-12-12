package customer

type Address struct {
	ID         string `gorm:"primaryKey"`
	CustomerID string

	Name     string
	City     string
	Street   string
	StreetNo string
}
