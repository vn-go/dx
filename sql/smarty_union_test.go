package sql

import (
	"fmt"
	"testing"
)

func TestSmartyUnion(t *testing.T) {
	str := `
	subSets(employee.code,where(employee.id = 1)) qr1,
	subSets(user.username,where(user.id = 2)) qr2,
	subSets(user.username,where(user.id = 3)) qr3,
	union(qr1+qr2*qr3),
	`

	sql, err := smartier.simple(str)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}
