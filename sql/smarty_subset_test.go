package sql

import (
	"fmt"
	"testing"
)

func TestSubsets(t *testing.T) {
	str := `
		subSets(employee.code,where(employee.id = 1)) qr1,
		subSets(user.username,where(user.id = 2)) qr2,
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
