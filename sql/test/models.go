package test

import (
	"github.com/vn-go/dx"
)

type Increment struct {
	ID int `db:"pk;auto"`
}
type IncrementDetail struct {
	ID          int `db:"pk;auto"`
	IncrementID int
	ItemID      int
	Amount      float64
}
type Decrement struct {
	ID int `db:"pk;auto"`
}
type DecrementDetail struct {
	ID          int `db:"pk;auto"`
	DecrementID int
	ItemID      int
	Amount      float64
}
type Item struct {
	ID    int `db:"pk;auto"`
	Name  string
	Price float64
}
type Customers struct {
	ID   int `db:"pk;auto"`
	Name string
	Age  int
}

func init() {
	dx.AddModels(
		&Increment{},
		&IncrementDetail{},
		&Decrement{},
		&DecrementDetail{},
		&Item{},
		&Customers{},
	)

}
