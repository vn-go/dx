package shorttest

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (fx *InspectStatement) ResolveColName(t *sqlparser.ColName, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	fieldName := t.Name.String()
	qualifierName := ""
	if t.Qualifier.IsEmpty() {
		if len(fx.AliasEntityName) > 0 && len(fx.Tables) > 1 {
			return nil, NewQueryTypeError("'%s' is require qualify with dataset name", fieldName)
		}
		qualifierName = fx.AliasEntityNameRevert[fx.Tables[0]]
	} else {
		qualifierName = t.Qualifier.Name.String()
	}
	key := strings.ToLower(fmt.Sprintf("%s.%s", qualifierName, fieldName))
	if check, ok := fx.ColummsInDictToColumnsScope[key]; ok {
		if datasetName, ok := fx.ColumnsScope[check]; !ok {
			return nil, NewQueryTypeError("'%s' can not access to this column in dataset'%s'", fieldName, datasetName)
		}
	}
	field, ok := fx.ColumnsDict[key]
	if !ok {
		return nil, NewQueryTypeError("'%s' can not access to this column in dataset'%s'", fieldName, t.Qualifier.Name)
	}
	if fx.ColumnsDictRevert == nil {
		fx.ColumnsDictRevert = make(map[string]string)
	}
	fx.ColumnsDictRevert[strings.ToLower(field)] = fmt.Sprintf("%s.%s", qualifierName, fieldName)
	if fieldInfo, ok := fx.ColumnsFieldMap[strings.ToLower(fmt.Sprintf("%s.%s", qualifierName, fieldName))]; ok {
		return &ResolveInfo{
			Content:  field,
			IsColumn: true,
			Typ:      internal.Helper.GetSqlTypeFfromGoType(fieldInfo.Field.Type),
		}, nil
	}

	return &ResolveInfo{
		Content:  field,
		IsColumn: true,
		Typ:      -1,
	}, nil
}
