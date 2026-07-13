package expense

import "time"

type Category string

const (
	Food      Category = "food"
	Transport Category = "transport"
	Shopping  Category = "shopping"
	Other     Category = "other"
)

type Expense struct {
	ID        int       `json:"id"`
	Amount    int       `json:"amount"`
	Category  Category  `json:"category"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Category) IsValid() bool {
	switch c {
	case Food, Transport, Shopping, Other:
		return true
	default:
		return false
	}
}
