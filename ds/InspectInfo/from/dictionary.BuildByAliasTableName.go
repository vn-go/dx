package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/errors"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func NewDictionary() *Dictionary {
	return &Dictionary{
		Entities:        make(map[string]*entity.Entity),
		AliasMap:        make(map[string]string),
		AliasMapReverse: make(map[string]string),
		TableAlias:      make(map[string]string),
		FieldMap:        make(map[string]DictionaryItem),
	}
}

func (d *Dictionary) BuildByAliasTableName(dialect types.Dialect, alias, table string) error {
	ent := model.ModelRegister.FindEntityByName(table)
	if ent == nil {
		return errors.NewParseError("dataset %s not found", table)
	}
	d.Entities[strings.ToLower(table)] = ent
	d.AliasMap[strings.ToLower(table)] = ent.TableName
	d.AliasMapReverse[strings.ToLower(ent.TableName)] = table
	d.TableAlias[ent.TableName] = strings.ToLower(alias)
	if alias != "" {
		d.Entities[strings.ToLower(alias)] = d.Entities[strings.ToLower(table)]
		d.AliasMap[strings.ToLower(alias)] = ent.TableName
		d.AliasMapReverse[strings.ToLower(ent.TableName)] = strings.ToLower(alias)
	}
	for _, col := range ent.Cols {
		key := strings.ToLower(fmt.Sprintf("%s.%s", table, col.Field.Name))
		d.FieldMap[key] = DictionaryItem{
			Content: fmt.Sprintf("%s.%s", dialect.Quote(ent.TableName), dialect.Quote(col.Name)),
			DbType:  internal.Helper.GetSqlTypeFfromGoType(col.Field.Type),
			Alias:   col.Field.Name,
		}
		key2 := strings.ToLower(fmt.Sprintf("%s.%s", alias, col.Field.Name))
		d.FieldMap[key2] = DictionaryItem{
			Content: fmt.Sprintf("%s.%s", dialect.Quote(strings.ToLower(alias)), dialect.Quote(col.Name)),
			DbType:  internal.Helper.GetSqlTypeFfromGoType(col.Field.Type),
			Alias:   col.Field.Name,
		}
	}
	return nil
}
