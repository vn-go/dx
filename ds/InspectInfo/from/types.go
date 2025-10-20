package from

import (
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/sqlparser"
)

type DictionaryItem struct {
	Content string
	DbType  sqlparser.ValType
	Alias   string
}
type Dictionary struct {
	FieldMap        map[string]DictionaryItem
	AliasMap        map[string]string
	AliasMapReverse map[string]string
	Entities        map[string]*entity.Entity
	TableAlias      map[string]string
}
