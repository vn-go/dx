package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sql"
	"github.com/vn-go/dx/sqlparser"
)

var cnn = "sqlserver://sa:123456@localhost:1433?database=hrm"
var dsnMySql = "root:123456@tcp(127.0.0.1:3306)/hrm"

func TestSmartQuery(t *testing.T) {
	db, err := dx.Open("mysql", dsnMySql)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `
														max(item.price) MaxPrice, 
														where (MaxPrice > 1000 and item.name='admin'),
														from(
														itm.id=incrementDetail.itemId,
														item itm
														)`)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)
}
func BenchmarkSmartQuery(b *testing.B) {
	db, err := dx.Open("mysql", dsnMySql)
	assert.NoError(b, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)
	b.ResetTimer()
	query := `
			max(item.price) MaxPrice, 
			where (MaxPrice > 1000),
			from(item.id=incrementDetail.itemId)`
	expectedQuery := "SELECT max(`T1`.`price`) `MaxPrice` FROM `items` `T1` join  `increment_details` `T2` ON `T1`.`id` = `T2`.`item_id` HAVING max(`T1`.`price`) > {1}"
	for i := 0; i < b.N; i++ {
		sqlCompiled, err := sql.Compiler.Resolve(dialect, query)
		if err != nil {
			panic(err)
		}
		assert.Equal(b, expectedQuery, sqlCompiled.Query)
	}

	b.Run("parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sqlCompiled, err := sql.Compiler.Resolve(dialect, query)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, expectedQuery, sqlCompiled.Query)
			}
		})
	})
}
func TestSelect2TableJoin(t *testing.T) {

	//db, err := dx.Open("sqlserver", cnn)
	db, err := dx.Open("mysql", dsnMySql)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `
		select text(item.id)  Code
		from 
		item 	left join (select id,price from item where id>1000) qr1
					on item.id = qr1.itemId
				left join incrementDetail 
					on item.id = incrementDetail.itemId
	`)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)

}
func BenchmarkSelect2TableJoin(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(b, err)
	defer db.Close()

	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
		SELECT item.* 
		FROM item 
		LEFT JOIN (SELECT id, price FROM item WHERE id > 1000) qr1 ON item.id = qr1.itemId 
		LEFT JOIN incrementDetail ON item.id = incrementDetail.itemId
	`

	b.Run("parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := sql.Compiler.Resolve(dialect, query)
				if err != nil {
					panic(err)
				}
			}
		})
	})

	b.Run("sequential", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := sql.Compiler.Resolve(dialect, query)
			if err != nil {
				panic(err)
			}
		}
	})
}

func TestSelectOneTableWithSum(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `select sum(item.id)+min(price) Total  from item 
														where total>1000 or price>100 order by total desc`, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlCompiled.Query)

}
func BenchmarkSelectOneTableWithSum(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(b, err)
	defer db.Close()

	dialect := factory.DialectFactory.Create(db.DriverName)
	query := `
		SELECT SUM(item.id + 1) + MIN(price) AS Total 
		FROM item
		HAVING SUM(item.id + 1) + MIN(price) > 1000 OR price > 100
		ORDER BY price DESC
	`

	b.Run("parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := sql.Compiler.Resolve(dialect, query)
				if err != nil {
					panic(err)
				}
			}
		})
	})

	b.Run("sequential", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := sql.Compiler.Resolve(dialect, query)
			if err != nil {
				panic(err)
			}
		}
	})
}

func TestSelectOneTable(t *testing.T) {

	db, err := dx.Open("sqlserver", cnn)
	assert.NoError(t, err)
	defer db.Close()
	dialect := factory.DialectFactory.Create(db.DriverName)

	sqlCompiled, err := sql.Compiler.Resolve(dialect, `select name, sum(item.id+1) Total from item 
														where (id = ? and name ='admin') and (item.id+1>0)`, 1)
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
		_, err := sql.Compiler.ResolveNoCache(dialect, "select name, sum(item.id+1) Total from item where id = ? and name ='admin'")
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
	`, 1, 2, 3)
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
func TestCTE(t *testing.T) {
	sql := `WITH 
		CTE_BangTam1 AS (
			SELECT Col1 FROM BangA
		),
		CTE_BangTam2 AS (
			
			SELECT Col2 FROM BangB 
		),
		CTE_BangTam3 AS (
			
			SELECT Col3 FROM BangC
		)
			SELECT * FROM CTE_BangTam2`
	sql, err := internal.Helper.QuoteExpression(sql)
	if err != nil {
		panic(err)
	}
	node, err := sqlparser.Parse(sql)
	if err != nil {
		panic(err)
	}
	fmt.Println(node)
}
