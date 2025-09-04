package models

import (
	"time"

	"github.com/vn-go/dx"
)

type LeaveRequest struct {
	BaseModel
	ID         int `db:"pk;auto"`
	EmployeeId int `db:"idx"`
	StartDate  time.Time
	EndDate    time.Time
	Reason     string `db:"size:255"`
	Status     string `db:"size:20"` // pending, approved, rejected
}

func init() {
	dx.AddModels(&LeaveRequest{})

	dx.AddForeignKey[LeaveRequest]("EmployeeId", &Employee{}, "ID", nil)

}
