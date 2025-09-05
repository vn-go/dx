package models

import "time"

type BaseModel struct {
	RecordID    string     `db:"uk;size:36;default:uuid()"`
	CreatedAt   time.Time  `db:"default:now();idx"`
	UpdatedAt   *time.Time `db:"default:now();idx"`
	Description *string    `db:"size:255"`
}
