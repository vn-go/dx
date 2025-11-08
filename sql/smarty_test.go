package sql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelfJoin(t *testing.T) {
	str := `from(
				department dept, 
				department parent,
				left(dept.parentId = parent.id),
			)`
	sql, _, _, err := smartier.simple(str)
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
		sql, _, _, err := smartier.simpleCache(str)
		if err != nil {
			panic(err)
		}
		expect := "SELECT * FROM department dept LEFT JOIN department parent ON dept.parentId = parent.id"
		assert.Equal(b, expect, sql)
	}
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
	sql, _, _, err := smartier.simple(str)
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
	sql, _, _, err := smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)

}
