package compiler

import (
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func newCompilerFromSqlNode(node sqlparser.SQLNode, dialect types.Dialect) (*compiler, error) {
	var err error

	ret := &compiler{
		// sql:     sql,
		node:    node,
		dialect: dialect,
		args:    internal.CompilerArgs{},
	}

	tableList := tableExtractor.getTables(node, map[string]bool{})
	isSubQuery := false
	if tableList != nil {
		isSubQuery = tableList.isSubQuery
	}
	ret.returnField, err = FieldExttractor.GetFieldAlais(node, map[string]bool{}, isSubQuery)
	if err != nil {
		return nil, err
	}
	if tableList != nil {
		ret.dict = ret.CreateDictionary(tableList.tables, ret.returnField)
	}

	return ret, nil

}
