package sqlserver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestSelectCrossTab(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	query := `
				from(user u),
				crossTab(for(day(u.createdOn) Day ,1,5),select(count(u.id) Total))
				`
	expectQuery := "SELECT count(CASE WHEN day([u].[created_on]) = @p1 THEN [u].[id] END) [TotalDay1], count(CASE WHEN day([u].[created_on]) = @p2 THEN [u].[id] END) [TotalDay2], count(CASE WHEN day([u].[created_on]) = @p3 THEN [u].[id] END) [TotalDay3], count(CASE WHEN day([u].[created_on]) = @p4 THEN [u].[id] END) [TotalDay4], count(CASE WHEN day([u].[created_on]) = @p5 THEN [u].[id] END) [TotalDay5] FROM [sys_users] [u]"
	expectArgs := `[
   1,
   2,
   3,
   4,
   5
 ]`
	expectedScopeAccess := `{
  "user.createdon": {
    "EntityName": "User",
    "EntityFieldName": "CreatedOn"
  },
  "user.id": {
    "EntityName": "User",
    "EntityFieldName": "Id"
  }
}`
	sql, err := db.Smart(query)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectQuery, sql.Query)
	assert.Equal(t, expectArgs, sql.Args.String())

	assert.Equal(t, expectedScopeAccess, sql.ScopeAccess.String())
}
func BenchmarkSelectCrossTab(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	query := `
				from(user u),
				crossTab(for(day(u.createdOn) Day ,1,5),select(count(u.id) Total))
				`
	expectQuery := "SELECT count(CASE WHEN day([u].[created_on]) = @p1 THEN [u].[id] END) [TotalDay1], count(CASE WHEN day([u].[created_on]) = @p2 THEN [u].[id] END) [TotalDay2], count(CASE WHEN day([u].[created_on]) = @p3 THEN [u].[id] END) [TotalDay3], count(CASE WHEN day([u].[created_on]) = @p4 THEN [u].[id] END) [TotalDay4], count(CASE WHEN day([u].[created_on]) = @p5 THEN [u].[id] END) [TotalDay5] FROM [sys_users] [u]"
	expectArgs := `[
   1,
   2,
   3,
   4,
   5
 ]`
	expectedScopeAccess := `{
  "user.createdon": {
    "EntityName": "User",
    "EntityFieldName": "CreatedOn"
  },
  "user.id": {
    "EntityName": "User",
    "EntityFieldName": "Id"
  }
}`

	b.Run("no-parallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql, err := db.Smart(query)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, expectQuery, sql.Query)
			assert.Equal(b, expectArgs, sql.Args.String())

			assert.Equal(b, expectedScopeAccess, sql.ScopeAccess.String())
		}
	})
	b.Run("parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart(query)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, expectQuery, sql.Query)
				assert.Equal(b, expectArgs, sql.Args.String())

				assert.Equal(b, expectedScopeAccess, sql.ScopeAccess.String())
			}
		})
	})
}
func TestSelectCrossTab002(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	expectedQuery := `SELECT count(CASE WHEN day([u].[created_on]) = @p1 THEN [u].[id] END) [TotalDay1], count(CASE WHEN day([u].[created_on]) = @p2 THEN [u].[id] END) [TotalDay2], count(CASE WHEN day([u].[created_on]) = @p3 THEN [u].[id] END) [TotalDay3], count(CASE WHEN day([u].[created_on]) = @p4 THEN [u].[id] END) [TotalDay4], count(CASE WHEN day([u].[created_on]) = @p5 THEN [u].[id] END) [TotalDay5], count(CASE WHEN day([u].[created_on]) = @p6 THEN [u].[id] END) [Day1], count(CASE WHEN day([u].[created_on]) = @p7 THEN [u].[id] END) [Day2], count(CASE WHEN day([u].[created_on]) = @p8 THEN [u].[id] END) [Day3], count(CASE WHEN day([u].[created_on]) = @p9 THEN [u].[id] END) [Day4], count(CASE WHEN day([u].[created_on]) = @p10 THEN [u].[id] END) [Day5] FROM [sys_users] [u] left join  [sys_roles] [r] ON [u].[role_id] = [r].[id] WHERE year([u].[created_on]) = @p11 OR year([r].[created_on]) = @p12
`
	query := `
	from(user u, role r,left(u.roleId=r.id)),
	crossTab(for(day(u.createdOn) Day ,1,5),select(count(u.id) user,sum(r.id) Role) role),
	where (year(u.createdOn)=2021 or year(r.createdOn)=2022)
	`
	sql, err := db.Smart(query)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedQuery, sql.Query)
}
func TestSelectCrossTab003(t *testing.T) {
	query := `
	subsets(
			from(user u, role r,left(u.roleId=r.id)),
			crossTab(
							for(day(u.createdOn) Day ,1,5),
							select(	count(u.id) user,
									sum(r.id) Role) role
					)
	) a,
	subsets(
			from(user u, role r,right(u.roleId=r.id)),
			crossTab(
							for(day(u.createdOn) Day ,1,5),
							select(count(u.id) user,
							sum(r.id) Role) role
					)
	) a,
	uion(a*b)
	`
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := db.Smart(query)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)

}
