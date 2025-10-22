package sql

import (
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

type dictionaryBuildItem struct {
	dbTableName string
	alias       string
}

func (dictBuildItem *dictionaryBuildItem) backtick(dialect types.Dialect) string {

	return dialect.Quote(dictBuildItem.dbTableName) + " " + dialect.Quote(dictBuildItem.alias)
}

// dictionary.Build.go
func (d *dictionary) Build(alias string, datasetName string, dialect types.Dialect) (*dictionaryBuildItem, error) {
	ent := model.ModelRegister.FindEntityByName(datasetName)
	if ent == nil {
		return nil, newCompilerError("dataset %s not found", datasetName)
	}
	d.tableAlias[strings.ToLower(datasetName)] = alias
	d.tableAlias[strings.ToLower(ent.TableName)] = alias
	for _, col := range ent.Cols {
		key := strings.ToLower(alias + "." + col.Name)
		fieldExpr := dialect.Quote(alias, col.Name)
		d.fields[key] = dictionaryField{
			Expr: fieldExpr,
			Typ:  internal.Helper.GetSqlTypeFfromGoType(col.Field.Type),
		}
	}
	return &dictionaryBuildItem{
		dbTableName: ent.TableName,
		alias:       alias,
	}, nil
}
