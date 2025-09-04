package model

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Table1 struct {
	Id   int    `db:"auto;pk";`
	Code string `db:"size:50`
}

func TestREgModle(t *testing.T) {
	ModelRegister.RegisterType(reflect.TypeFor[Table1]())
	ret, err := GetModel[Table1]()
	assert.NoError(t, err)
	assert.NotEmpty(t, ret)
}
