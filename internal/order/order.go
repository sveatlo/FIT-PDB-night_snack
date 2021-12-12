package order

type Order struct {
	ID                string
	CustomerID        string
	CustomerAddressID string

	MenuItems []string
}
