package model

import (
	"reflect"
	"sync"

	"github.com/vn-go/dx/entity"
)

type modelRegistryInfo struct {
	TableName string
	ModelType reflect.Type
	Entity    *entity.Entity
}
type modelRegister struct {
	cacheModelRegistry  map[reflect.Type]*modelRegistryInfo
	cacheGetModelByType sync.Map
}
