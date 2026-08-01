package task

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

type TaskQuery struct {
	Status TaskStatus
	Search string
	Page   int
	Limit  int
	Sort   string
	Order  Order
}
