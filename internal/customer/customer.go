package customer

type Customer struct {
	ID        string `gorm:"primaryKey"`
	FirstName string
	LastName  string
}
