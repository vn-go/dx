package sql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func TestSelfJoin(t *testing.T) {
	str := `from(
				department dept, 
				department parent,
				left(dept.parentId = parent.id),
			)`
	sql, err := smartier.simple(str)
	if err != nil {
		panic(err)
	}
	expect := "SELECT * FROM department dept LEFT JOIN department parent ON dept.parentId = parent.id"
	assert.Equal(t, expect, sql)
}
func BenchmarkTestSelfJoin(b *testing.B) {
	str := `from(
				department dept, 
				department parent,
				left(dept.parentId = parent.id and sum(dept.numberOfEmp)>100),
			)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sql, err := smartier.simpleCache(str)
		if err != nil {
			panic(err)
		}
		expect := "SELECT * FROM department dept LEFT JOIN department parent ON dept.parentId = parent.id"
		assert.Equal(b, expect, sql)
	}
}
func TestSubsets(t *testing.T) {
	str := `
		subSets(employee.code,where(employee.id = 1)) qr1,
		subSets(employee.code,where(employee.id = 2)) qr2,
		subSets(
					from(department dept, department parent,dept.parentId = parent.id)
				) dept,
		from(qr1.id=qr2.id),
		where(qr1.code = 'abc' and qr2.code = 'def'),
		qr1.name`

	sql, err := smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)

}
func Test0001(t *testing.T) {
	str := `from(
					employee emp,
					user usr,
					department dept,
					emp.department_id like department.id,
					emp.userId = usr.id,
			),emp.code+dept.code Code, where(emp.id = select(max(emp.id)))`
	str = `max(emp.id), from(employee emp)`
	sql, err := smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)

}
func BenchmarkTest0001(b *testing.B) {
	str := `from(
					employee emp,
					user usr,
					department dept,
					emp.department_id = department.id,
					emp.userId = usr.id,
			),emp.code+dept.code Code`
	str, _ = internal.Helper.QuoteExpression2(str)

	tk := sqlparser.NewStringTokenizer("select " + str)
	stm, err := sqlparser.ParseNext(tk)

	if err != nil {
		panic(err)
	}
	selectStm := stm.(*sqlparser.Select)
	for i := 0; i < b.N; i++ {
		smartier.from(selectStm)
	}

}
