package quicky

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

// sqlNode.BuildDictionaryWithAlias
func (s sqlNode) BuildDictionaryWithAlias(node *sqlparser.ColName, dialect types.Dialect, dict *Dictionanry, alias string) error {
	if dict.AliasMap == nil {
		dict.AliasMap = map[string]string{}
	}
	if dict.AliasMapReverse == nil {
		dict.AliasMapReverse = map[string]string{}
	}
	if dict.FieldMap == nil {
		dict.FieldMap = map[string]DictionanryItem{}
	}
	if dict.Entities == nil {
		dict.Entities = map[string]*entity.Entity{}
	}
	indexOfAlais := len(dict.AliasMap) + 1
	if alias == "" {
		alias = fmt.Sprintf("T%d", indexOfAlais)
	}
	tableName := node.Name.String()

	ent := model.ModelRegister.FindEntityByName(tableName)
	if ent == nil {
		return newParseError("dataset '%s' was not found", tableName)
	}
	dict.AliasMap[strings.ToLower(tableName)] = ent.TableName
	dict.AliasMap[strings.ToLower(alias)] = ent.TableName
	dict.AliasMapReverse[strings.ToLower(ent.TableName)] = alias
	dict.AliasMapReverse[strings.ToLower(tableName)] = alias
	dict.Entities[ent.TableName] = ent
	for _, col := range ent.Cols {
		key := fmt.Sprintf("%s.%s", tableName, col.Field.Name)
		dict.FieldMap[strings.ToLower(key)] = DictionanryItem{
			Content: dialect.Quote(ent.TableName, col.Name),
			DbType:  internal.Helper.GetSqlTypeFfromGoType(col.Field.Type),
			Alias:   col.Field.Name,
		}

	}
	return nil
}
