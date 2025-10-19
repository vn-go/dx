package quicky

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

func (s *SqlParser) Parse(data *QueryItem, dialect types.Dialect, textParams []string, args ...any) error {
	dict := Dictionanry{}
	argsInspects := &ArgInspects{}
	field := &FieldInspects{}
	jonClause := ""
	selectStr := ""
	var err error
	// determine Form clause
	if dataInspect, ok := data.InspectData["from"]; ok {
		jonClause, err = dataInspect.JoinClauseInFrom(dialect, textParams, args, argsInspects, field, &dict)
		if err != nil {
			return err
		}
		fmt.Println(jonClause, args)
	} else if dataInspect, ok := data.InspectData["select"]; ok { // determine Form clause in select clause

		jonClause, err = dataInspect.JoinClause(dialect, textParams, args, argsInspects, field, &dict)
		if err != nil {
			return err
		}
		fmt.Println(jonClause, args)
	} else {
		panic("Form clause not found,see SqlParser.Parse")
	}
	selectStr, err = data.InspectData["select"].BuilSelect(dialect, textParams, args, argsInspects, field, &dict)
	if err != nil {
		return err
	}
	fmt.Println(selectStr, args)
	return nil

}
