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
	if t.Qualifier.Name.IsEmpty() {
		// check alias table in dict
		alias := "T1" // if not found, use default alias
		// determine database tabel name
		var ent *entity.Entity
		if entFind, ok := injector.dict.aliasToEntity[strings.ToLower(alias)]; ok {
			ent = entFind
		} else {
			return nil, newCompilerError("dataset was not found")
		}

		key := strings.ToLower(fmt.Sprintf("%s.%s", alias, t.Name.String()))

		if field, ok := injector.dict.fields[key]; ok {
			refFieldKey := strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), field.EntityField.Name))
			return &compilerResult{
				Content: field.Expr,
				Args:    nil,
				Fields: refFields{
					refFieldKey: refFieldInfo{
						EntityName:      ent.EntityType.Name(),
						EntityFieldName: field.EntityField.Name,
					},
				},
				selectedExprs: dictionaryFields{
					strings.ToLower(field.Expr): dictionaryField{
						Expr:  field.Expr,
						Typ:   field.Typ,
						Alias: field.EntityField.Name,
					},
				},
				selectedExprsReverse: dictionaryFields{
					field.EntityField.Name: dictionaryField{
						Expr:  field.Expr,
						Typ:   field.Typ,
						Alias: field.EntityField.Name,
					},
				},
			}, nil
		}

	}

	panic("unimplemented")
}
