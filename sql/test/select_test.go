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

func TestSelectOneTable(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, "select item.id from item where id = ? and name ='admin'", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)

}
func BenchmarkSelectOneTable(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(b, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sql.Compiler.Resolve(dialect, "select item.id from item where id = ? and name ='admin'", 1)
		if err != nil {
			panic(err)
		}
	}

}
func TestSelect(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, "select * from item where id = ?", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)

}
func TestSelect2Table(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, "select * from item, decrementDetail where item.id = ?", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)

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
	fmt.Println(sqlCompiled.Query)

}
func TestSelectSubQuery(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `
	select qr1.ItemD  from (
			select id ItemD from item where id = ?
			) qr1
	`, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)

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
	fmt.Println(sqlCompiled.Query)

}
