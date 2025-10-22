package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/sql"
)

var cnn = "sqlserver://sa:123456@localhost:1433?database=hrm"

func TestSelect(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, "select * from item where id = ?", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Content)

}
func TestSelectUnion(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `
	select * from increment where id = ?
	union all
	select * from decrement where id = ?
	union all
	select * from incrementDetail where id = ?
	`, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Content)

}
func TestSelectSubQuery(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `
	select * from (
			select * from hrm.employee where id = ?
			) qr1
	`, 1)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sqlCompiled.Content)

}
func TestSelectSubQueryUnion(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `
	select * from (
				select * from hrm.employee where id = ?
				union all
				select * from hrm.employee where id = ?
			) qr1
	`, 1)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sqlCompiled.Content)

}
