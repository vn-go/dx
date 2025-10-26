package models

import (
	"github.com/vn-go/dx"
)

type Department struct {
	ID        int    `db:"pk;auto"`
	Name      string `db:"size:100;uk:uq_dept_name"`
	Code      string `db:"size:50;uk:uq_dept_code"`
	ChidrenID []int  `db:"idx"`
	ParentID  *int
	BaseModel
}

func init() {
	dx.AddForeignKey[Department]("ParentID", &Department{}, "ID", &dx.FkOpt{
		OnDelete: false,
		OnUpdate: false,
	})

}
