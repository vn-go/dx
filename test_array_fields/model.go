package testarrayfields

import (
	"time"

	"github.com/vn-go/dx"
)

type Department struct {
	ID         int    `db:"pk;auto"`
	Name       string `db:"size:100;uk:uq_dept_name"`
	Code       string `db:"size:50;uk:uq_dept_code"`
	ChildrenID []int  `db:"idx"`
	ParentID   *int
	BaseModel
}
type BaseModel struct {
	RecordID    string     `db:"uk;size:36;default:uuid()"`
	CreatedAt   time.Time  `db:"default:now();idx"`
	UpdatedAt   *time.Time `db:"default:now();idx"`
	Description *string    `db:"size:255"`
}

func init() {
	dx.AddForeignKey[Department]("ParentID", &Department{}, "ID", &dx.FkOpt{
		OnDelete: false,
		OnUpdate: false,
	})

}
