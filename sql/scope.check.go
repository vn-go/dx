package sql

import (
	"fmt"
	"strings"
)

type DataScope struct {
	IsFull bool
	Scope  map[string]string
}
type queryScope struct {
	Fields []string
	IsFull bool
}
type QueryScopes map[string]queryScope

func (q *QueryScopes) Validate(AccessScope accessScopes, AccaccessScopesHash256 string) error {
	// check is allow acess to entity
	for k, v := range *q {
		if _, ok := AccessScope[k]; !ok {
			return fmt.Errorf("access denied to dataset %s", k)
		}
		// check is allow access to fields
		for _, f := range v.Fields {
			if _, ok := AccessScope[strings.ToLower(k)][strings.ToLower(f)]; !ok {
				return fmt.Errorf("access denied to field %s.%s", k, f)
			}
		}
	}
	return nil
}
