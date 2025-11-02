package dx

import "github.com/vn-go/dx/sql"

type ds struct {
}

type scopeAccessor struct {
}

func (s *scopeAccessor) New(tableName string, fields ...string) *sql.QueryScopes {
	if len(fields) == 0 {
		return &sql.QueryScopes{
			tableName: {
				IsFull: true,
			},
		}
	}
	return &sql.QueryScopes{
		tableName: {
			Fields: fields,
		},
	}
}

var ScopeAccess = &scopeAccessor{}

var Data = ds{}
