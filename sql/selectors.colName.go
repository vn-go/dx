package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/sqlparser"
)

// selectors.colName.go
func (s selectors) colName(t *sqlparser.ColName, injector *injector, cmpTyp CMP_TYP, selectedExprsReverse dictionaryFields) (*compilerResult, error) {

	if len(injector.dict.entities) > 1 && t.Qualifier.Name.IsEmpty() {

		return nil, newCompilerError(ERR_AMBIGUOUS_FIELD_NAME, "'%s' is ambiguous, specify dataset name", t.Name.String())
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
		} else if len(injector.dict.aliasToEntity) == 1 && cmpTyp != CMP_SELECT && cmpTyp != CMP_JOIN {
			// If selecting from only one table, and the SELECT clause hasn't been compiled yet,
			// the current field is probably an alias of an expression in the SELECT clause.
			for k, v := range injector.dict.aliasToEntity {
				ent = v
				alias = k
				break
			}
		} else if subqueryEntity, ok := injector.dict.subqueryEntites[strings.ToLower(alias)]; ok {
			return s.colNameInSubquery(t, injector, subqueryEntity)

		} else {
			if cmpTyp == CMP_WHERE || cmpTyp == CMP_ORDER_BY {
				if cmpField, ok := selectedExprsReverse[t.Name.Lowered()]; ok {
					return &compilerResult{
						Content:           cmpField.Expr,
						OriginalContent:   originalContent,
						AliasOfContent:    cmpField.Alias,
						IsInAggregateFunc: cmpField.IsInAggregateFunc,
						IsExpression:      true,
					}, nil
				}
				return nil, newCompilerError(ERR_FIELD_NOT_FOUND, "Field '%s' was not found", t.Name.String())
			}
			if t.Qualifier.IsEmpty() {
				return nil, newCompilerError(ERR_AMBIGUOUS_FIELD_NAME, "'%s' is ambiguous, specify dataset name", t.Name.String())
			}
			return nil, newCompilerError(ERR_DATASET_NOT_FOUND, "Dataset '%s' was not found", t.Qualifier.Name.String())
		}
	}

	key := strings.ToLower(fmt.Sprintf("%s.%s", alias, t.Name.String()))
	fr := refFields{}
	if field, ok := injector.dict.fields[key]; ok {
		refFieldKey := fmt.Sprintf("%s.%s", alias, t.Name.String())
		retAlias := alias
		if ent.EntityType != nil {
			refFieldKey = strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), field.EntityField.Field.Name))
			retAlias = field.EntityField.Field.Name
			fr = refFields{ // add ref field for permission check
				refFieldKey: refFieldInfo{
					EntityName:      ent.EntityType.Name(),
					EntityFieldName: field.EntityField.Field.Name,
				},
			}
		}

		cmpField := &dictionaryField{
			Expr:  field.Expr,
			Typ:   field.Typ,
			Alias: retAlias,
		}
		return &compilerResult{
			Content:         field.Expr,
			OriginalContent: originalContent,
			Args:            nil,
			Fields:          fr,
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
	} else if cmpTyp == CMP_WHERE || cmpTyp == CMP_ORDER_BY {
		if cmpField, ok := selectedExprsReverse[t.Name.Lowered()]; ok {
			return &compilerResult{
				Content:           cmpField.Expr,
				OriginalContent:   originalContent,
				AliasOfContent:    cmpField.Alias,
				IsInAggregateFunc: cmpField.IsInAggregateFunc,
				IsExpression:      true,
			}, nil
		}
		return nil, newCompilerError(ERR_FIELD_NOT_FOUND, "field '%s' was not found", t.Name.String())
	}
	if ent != nil {
		return nil, newCompilerError(ERR_FIELD_NOT_FOUND, "field '%s' was not found in dataset '%s'", t.Name.String(), ent.EntityType.Name())
	}

	panic("unimplemented, see selectors.colName")
}
