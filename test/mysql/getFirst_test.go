package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"

func TestGetFirst(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	var user models.User
	err = db.First(&user)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
	err = db.First(&user, "userid=?", user.UserId)
	t.Log(user)
	if err != nil {
		t.Error(err)
	}

}
func TestDataSourceFromModel(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	ds := db.ModelDatasource("user").Select("count(id) as Count,sum(id) Sum")
	ds = ds.Where("Count*Sum >? and username='admin'", 100)
	sql, err := ds.ToDict()
	if err != nil {
		t.Error(err)
	}
	t.Log(sql)
	ds1 := db.ModelDatasource("xxx").Select("count(id) as Count,sum(id) Sum")
	ds1 = ds1.Where("Count*Sum >? and username='admin'", 100)
	sq2, err := ds1.ToDict()
	if err != nil {
		t.Error(err)
	}
	t.Log(sq2)
}
func TestDataSourceFromModelToSql(t *testing.T) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	ds := db.ModelDatasource("user").Select("count(id) as Count,sum(id) Sum")
	ds = ds.Where("Count*Sum >? and username='admin'", 100)
	sql, err := ds.ToSql()
	if err != nil {
		t.Error(err)
	}
	t.Log(sql.Sql)

}
func TestSmartySql(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := db.Smart(`
		select count(id)  Count,sum(id) Sum
		from user 
		
		where Count*Sum >? and username='admin'
		group by username
	`, 100)
	if err != nil {
		panic(err)
	}
	t.Log(sql)
	println(sql.Query)
}
func TestSmartySqlLevel2(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := db.Smart(`
		from(user u), 
		where(Count*Sum >? and username='admin'),
		group(username),
		count(u.id) Count, 
		sum(u.id) Sum, 
	`, 100)
	if err != nil {
		panic(err)
	}
	t.Log(sql)
	println(sql.Query)
}
func BenchmarkCompare(b *testing.B) {

	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	smartLele2ExpectedSql := "SELECT count(`u`.`id`) `Count`, sum(`u`.`id`) `Sum` FROM `sys_users` `u` WHERE `u`.`username` = {2} GROUP BY `u`.`username` HAVING count(`u`.`id`) * sum(`u`.`id`) > {1}"
	smartExpectedSql := "SELECT count(`T1`.`id`) `Count`, sum(`T1`.`id`) `Sum` FROM `sys_users` `T1` WHERE `T1`.`username` = {2} GROUP BY `T1`.`username` HAVING count(`T1`.`id`) * sum(`T1`.`id`) > {1}"
	classicExpectedSql := "SELECT COUNT(`T1`.`id`) `Count`,SUM(`T1`.`id`) `Sum` FROM `sys_users` `T1` GROUP BY `T1`.`username` HAVING COUNT(`T1`.`id`) * SUM(`T1`.`id`) > {1} AND `T1`.`username` = {2}"
	b.Run("Classic", func(b *testing.B) {

		ds := db.ModelDatasource("user").Select("count(id) as Count,sum(id) Sum")
		ds = ds.Where("Count*Sum >? and username='admin'", 100) // raw sql string builder was completed
		/*
			Raw sql is:
				select count(id)  Count,sum(id) Sum
				from user
				where Count*Sum >? and username='admin'
		*/
		b.ResetTimer()
		for i := 0; i < b.N; i++ {

			sql, err := ds.ToSql()
			if err != nil {
				panic(err)
			}

			assert.Equal(b, classicExpectedSql, sql.Sql)
		}

	})
	b.Run("Smarty", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql, err := db.Smart(`
			select count(id)  Count,sum(id) Sum
			from user 
			
			where Count*Sum >? and username='admin'
			group by username
		`, 100)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, smartExpectedSql, sql.Query)
		}

	})
	b.Run("Smarty-leve2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql, err := db.Smart(`
			from(user u), 
			where(Count*Sum >? and username='admin'),
			group(username),
			count(u.id) Count, 
			sum(u.id) Sum, 
			`, 100)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, smartLele2ExpectedSql, sql.Query)
		}

	})
	b.Run("Classic-Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ds := db.ModelDatasource("user").Select("count(id) as Count,sum(id) Sum")
				ds = ds.Where("Count*Sum >? and username='admin'", 100) // raw sql string builder was completed
				sql, err := ds.ToSql()
				if err != nil {
					panic(err)
				}
				assert.Equal(b, classicExpectedSql, sql.Sql)
			}
		})
	})
	b.Run("Smarty-Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart(`
				select count(id)  Count,sum(id) Sum
				from user 
				
				where Count*Sum >? and username='admin'
				group by username
			`, 100)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, smartExpectedSql, sql.Query)
			}
		})
	})
	b.Run("Smarty-level2-Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart(`
			from(user u), 
			where(Count*Sum >? and username='admin'),
			group(username),
			count(u.id) Count, 
			sum(u.id) Sum, 
			`, 100)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, smartLele2ExpectedSql, sql.Query)
			}
		})
	})
}
func TestSmartySimplest(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := db.Smart("user.username, where(username='admin')")
	if err != nil {
		t.Error(err)
	}
	exprectedSql := "SELECT `T1`.`username` `Username` FROM `sys_users` `T1` WHERE `T1`.`username` = {1}"
	assert.Equal(t, exprectedSql, sql.Query)
}
func BenchmarkSimpliest(b *testing.B) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	ds := db.ModelDatasource("user").Select("username")
	ds = ds.Where("username='admin'")
	classicExpectedSql := "SELECT `T1`.`username` `username` FROM `sys_users` `T1` WHERE `T1`.`username` = {1}"
	smartSipliestExpectedSql := "SELECT `T1`.`username` `Username` FROM `sys_users` `T1` WHERE `T1`.`username` = {1}"
	b.Run("classic", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {

			sql, err := ds.ToSql()
			if err != nil {
				panic(err)
			}

			assert.Equal(b, classicExpectedSql, sql.Sql)
		}
	})
	b.Run("ssmart-simplest", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql, err := db.Smart("user.username, where(username='admin')")
			if err != nil {
				panic(err)
			}
			assert.Equal(b, smartSipliestExpectedSql, sql.Query)
		}
	})
	b.Run("classic-parallel", func(b *testing.B) {
		ds := db.ModelDatasource("user").Select("username")
		ds = ds.Where("username='admin'")
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {

				sql, err := ds.ToSql()
				if err != nil {
					panic(err)
				}
				assert.Equal(b, classicExpectedSql, sql.Sql)
			}
		})
	})
	b.Run("smart-simplest-parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart("user.username, where(username='admin')")
				if err != nil {
					panic(err)
				}
				assert.Equal(b, smartSipliestExpectedSql, sql.Query)
			}
		})
	})
}
