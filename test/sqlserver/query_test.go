package sqlserver

import (
	"testing"

	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

func TestQuery(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	db.Select()
	userInfos := []struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
	}{}
	err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id),skip(?),take(?)", 10, 10)
	if err != nil {
		panic(err)
	}
	t.Log(userInfos)

}
func BenchmarkQuery(t *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	t.Run("Query", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userInfos := []struct {
				Id       uint64 `db:"pk;auto" json:"id"`
				Username string `db:"size:50;uk" json:"username"`
			}{}

			err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id),skip(?),take(?)", 10, 1000)
			if err != nil {
				panic(err)
			}

		}
	})
	t.Run("Query-Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				userInfos := []struct {
					Id       uint64 `db:"pk;auto" json:"id"`
					Username string `db:"size:50;uk" json:"username"`
				}{}

				err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id),skip(?),take(?)", 10, 1000)
				if err != nil {
					panic(err)
				}

			}
		})
	})
}
