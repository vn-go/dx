package models

import (
	"github.com/vn-go/dx"
)

type Employee struct {
	ID           int    `json:"id" db:"pk;auto;"`
	FirstName    string `json:"name" db:"size(50);idx"`
	LastName     string `json:"lastName" db:"size(50);idx"`
	DepartmentID int    `json:"departmentId"`
	PositionID   int    `json:"positionId"`
	UserID       uint64 `json:"userId"`
	BaseModel
}

func init() {
	dx.AddModels(Employee{}, &Department{})
	dx.AddForeignKey[Employee](
		"DepartmentID",
		&Department{},
		"ID", nil)

	dx.AddForeignKey[Employee](
		"PositionID",
		&Position{},
		"ID", nil)
	// dx.AddForeignKey[Employee](
	// 	"UserID",
	// 	&User{},
	// 	"ID", nil,
	// )

}
