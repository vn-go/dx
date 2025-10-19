package quicky

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

func (s *SqlParser) Parse(data *QueryItem, dialect types.Dialect, textParams []string, args ...any) error {
	dict := Dictionanry{}
	argsInspects := &ArgInspects{}
	field := &FieldInspects{}
	// determine Form clause
	if expr, ok := data.InspectData["from"]; ok {
		var onClause sqlparser.SQLNode
		for _, node := range expr.nodes {
			switch node.(type) {
			case *sqlparser.ComparisonExpr:
				onClause = node.(sqlparser.SQLNode)
			case *QueryItem:
				rSql := &SqlParser{}
				err := rSql.Parse(node.(*QueryItem), dialect, textParams, args...)
				if err != nil {
					return err
				}

			default:
				panic("Unsupported form clause")
			}

		}
		fmt.Println(onClause)
	} else if data, ok := data.InspectData["select"]; ok { // determine Form clause in select clause

		jonClause, err := data.JoinClause(dialect, textParams, args, argsInspects, field, &dict)
		if err != nil {
			return err
		}
		fmt.Println(jonClause, args)
	} else {
		panic("Form clause not found,see SqlParser.Parse")
	}
	return nil
}
