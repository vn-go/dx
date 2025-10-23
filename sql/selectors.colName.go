package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/sqlparser"
)

// selectors.colName.go
func (s selectors) colName(t *sqlparser.ColName, injector *injector) (*compilerResult, error) {
	if len(injector.dict.entities) > 1 && t.Qualifier.Name.IsEmpty() {
		return nil, newCompilerError("'%s' is ambiguous, specify dataset	 name", t.Name.String())
	}
	originalContent := t.Name.String()
	alias := "T1" // if not found, use default alias
	if !t.Qualifier.Name.IsEmpty() {
		alias = strings.ToLower(t.Qualifier.Name.String())
		originalContent = t.Qualifier.Name.String() + "." + originalContent
	}
	// check alias table in dict

	// determine database table name
	var ent *entity.Entity
	if entFind, ok := injector.dict.aliasToEntity[strings.ToLower(alias)]; ok {
		ent = entFind
	} else {

		if entFind, ok := injector.dict.entities[strings.ToLower(alias)]; ok {
			ent = entFind
		} else {

			return nil, newCompilerError("dataset was not found")
		}
	}

	key := strings.ToLower(fmt.Sprintf("%s.%s", alias, t.Name.String()))

	if field, ok := injector.dict.fields[key]; ok {
		refFieldKey := strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), field.EntityField.Name))
		cmpField := &dictionaryField{
			Expr:  field.Expr,
			Typ:   field.Typ,
			Alias: field.EntityField.Field.Name,
		}
		return &compilerResult{
			Content:         field.Expr,
			OriginalContent: originalContent,
			Args:            nil,
			Fields: refFields{ // add ref field for permission check
				refFieldKey: refFieldInfo{
					EntityName:      ent.EntityType.Name(),
					EntityFieldName: field.EntityField.Field.Name,
				},
			},
			selectedExprs: dictionaryFields{ // add selected expr next phase of compiler
				strings.ToLower(field.Expr): cmpField,
			},
			selectedExprsReverse: dictionaryFields{ // hold reverse of selected exprs for where clause compiler
				field.EntityField.Name: cmpField,
			},
			nonAggregateFields: dictionaryFields{ // hold non aggregate fields for group by clause compiler
				strings.ToLower(field.Expr): cmpField,
			},
			AliasOfContent: field.Alias,
		}, nil
	}
	panic("unimplemented, see selectors.colName")
}
