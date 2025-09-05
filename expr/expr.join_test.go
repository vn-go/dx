package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/dialect/factory"
)

func TestEormJoin(t *testing.T) {
	ej := &exprCompiler{

		Context: &exprCompileContext{
			Tables: []string{},
			Alias:  map[string]string{},
			schema: &map[string]bool{
				//"User": true,
			},
			Dialect: factory.DialectFactory.Create("mssql"),
			Purpose: BUILD_JOIN,
		},
	}
	err := ej.Build("Departments INNER JOIN User ON User.Code = Departments.Code INNER JOIN Check ON Check.Name = 'John'")
	assert.NoError(t, err)
	assert.Equal(t, "[departments] AS [T1] INNER JOIN [User] AS [T2] ON [T2].[code] = [T1].[code] INNER JOIN [checks] AS [T3] ON [T3].[name] = N'John'", ej.Content)

}
func BenchmarkEormJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ej := &exprCompiler{

			Context: &exprCompileContext{
				Tables: []string{},
				Alias:  map[string]string{},
				schema: &map[string]bool{
					"User": true,
				},
				Dialect: factory.DialectFactory.Create("mssql"),
				Purpose: BUILD_JOIN,
			},
		}
		err := ej.Build("Departments INNER JOIN User ON User.Code = Departments.Code INNER JOIN Check ON Check.Name = 'John'")
		if err != nil {
			b.Fail()
		}
	}
}
