package models

import "time"

type BaseModel struct {
	CreatedAt   *time.Time `db:"default:now();idx"`
	UpdatedAt   *time.Time `db:"default:now();idx"`
	Description *string    `db:"size:255"`
}
