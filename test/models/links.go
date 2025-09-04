package models

import (
	"github.com/vn-go/dx"
)

func init() {
	dx.AddModels(
		&Contract{},
		&User{},

		&Department{},
		&Position{},
		&Contract{},
		&User{},
	)
}
