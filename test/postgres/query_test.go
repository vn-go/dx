package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
	_ "github.com/vn-go/dx/test/models"
)

var pgDsn = "postgres://postgres:123456@localhost:5432/hrm?sslmode=disable&"

func TestQuery(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	userInfos := []struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
	}{}
	err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id),take(?)", 1, 10000)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 847, len(userInfos))

}
func TestQuery2(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	users := []models.User{}
	err = db.From(&models.User{}).Where("id>1").Order("id").Limit(10000).Offset(1).Find(&users)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 847, len(users))

}
func TestQueryDls2(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	users := []models.User{}
	err = db.DslQuery(&users, "user(),where(id>=1),sort(id),skip(?),take(?)", 1, 10000)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 847, len(users))

}
func BenchmarkSelectAllUsers(b *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	type count struct {
		Total int
	}
	var countResult []count
	err = db.DslQuery(&countResult, "user(count(id) Total),where(id>=1)", 0, 0)
	if err != nil {
		panic(err)
	}
	assert.Equal(b, 848, countResult[0].Total)
	countResult[0].Total = countResult[0].Total - 1 // gaim di 1 o test co skip 1

	b.Run("dsl", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			users := []models.User{}
			err = db.DslQuery(&users, "user(),where(id>=1),sort(id),skip(?),take(?)", 1, 10000)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, countResult[0].Total, len(users))
		}
	})
	b.Run("dsl-parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				users := []models.User{}
				err = db.DslQuery(&users, "user(),where(id>=1),sort(id),skip(?),take(?)", 1, 10000)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, countResult[0].Total, len(users))
			}
		})
	})
	b.Run("model", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			users := []models.User{}
			err = db.From(&models.User{}).Where("id>=1").Order("id").Limit(10000).Offset(1).Find(&users)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, countResult[0].Total, len(users))
		}
	})
	b.Run("model-parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				users := []models.User{}
				err = db.From(&models.User{}).Where("id>=1").Order("id").Limit(10000).Offset(1).Find(&users)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, countResult[0].Total, len(users))
			}
		})
	})
}
func TestGetFirst(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var user models.User
	err = db.DslFirstRow(&user, "user(),where(username=?)", "admin")
	if err != nil {
		panic(err)
	}
	t.Log(user)
}
func BenchmarkGetFirst(b *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var user models.User
	b.Run("use-dsl", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = db.DslFirstRow(&user, "user(),where(username=?)", "admin")
			if err != nil {
				panic(err)
			}
			assert.Equal(b, "admin", user.Username)
		}
	})
	b.Run("use-dsl-parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				err = db.DslFirstRow(&user, "user(),where(username=?)", "admin")
				if err != nil {
					panic(err)
				}
				assert.Equal(b, "admin", user.Username)
			}
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = db.DslFirstRow(&user, "user(),where(username=?)", "admin")
			if err != nil {
				panic(err)
			}
			assert.Equal(b, "admin", user.Username)
		}
	})
	b.Run("use-model", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := db.First(&user, "username = ?", "admin")

			if err != nil {
				panic(err)
			}
			assert.Equal(b, "admin", user.Username)
		}
	})
	b.Run("use-model-parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				err := db.First(&user, "username = ?", "admin")

				if err != nil {
					panic(err)
				}
				assert.Equal(b, "admin", user.Username)
			}
		})
	})

}
func BenchmarkQuery(t *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	t.Run("dsl-get-847-rows", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userInfos := []struct {
				Id       uint64 `db:"pk;auto" json:"id"`
				Username string `db:"size:50;uk" json:"username"`
			}{}
			err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id),skip(?),take(?)", 1, 10000)
			if err != nil {
				panic(err)
			}
			assert.Equal(t, 847, len(userInfos))
		}
	})
	t.Run("dsl-get-10000-rows-parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				userInfos := []struct {
					Id       uint64 `db:"pk;auto" json:"id"`
					Username string `db:"size:50;uk" json:"username"`
				}{}
				err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id),skip(?),take(?)", 1, 10000)
				if err != nil {
					panic(err)
				}
				assert.Equal(t, 847, len(userInfos))
			}
		})
	})
}
