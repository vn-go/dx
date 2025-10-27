package sqlserver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestSelectInOneTable(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	query := `user(username,hashpassword),where(username='admin')`
	expectedQuery := "SELECT [T1].[username] [Username], [T1].[hash_password] [HashPassword] FROM [sys_users] [T1] WHERE [T1].[username] = @p1"
	expectedArg := `[
   "admin"
 ]`
	scopeAccessExpected := `{
  "user.hashpassword": {
    "EntityName": "User",
    "EntityFieldName": "HashPassword"
  },
  "user.username": {
    "EntityName": "User",
    "EntityFieldName": "Username"
  }
}`
	sql, err := db.Smart(query)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
	assert.Equal(t, expectedQuery, sql.Query)
	assert.Equal(t, expectedArg, sql.Args.String())
	assert.Equal(t, scopeAccessExpected, sql.ScopeAccess.String())
	fmt.Println("------------------------------------------------")
	fmt.Println(sql.Args.String())
	fmt.Println("------------------------------------------------")
	fmt.Println(sql.ScopeAccess.String())
	fmt.Println("------------------------------------------------")
}
func BenchmarkSelectInOneTable(t *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	query := `user(username,hashpassword),where(username='admin')`
	expectedQuery := "SELECT [T1].[username] [Username], [T1].[hash_password] [HashPassword] FROM [sys_users] [T1] WHERE [T1].[username] = @p1"
	expectedArg := `[
   "admin"
 ]`
	scopeAccessExpected := `{
  "user.hashpassword": {
    "EntityName": "User",
    "EntityFieldName": "HashPassword"
  },
  "user.username": {
    "EntityName": "User",
    "EntityFieldName": "Username"
  }
}`
	t.Run("Smart", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			sql, err := db.Smart(query)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, expectedQuery, sql.Query)
			assert.Equal(t, expectedArg, sql.Args.String())
			assert.Equal(t, scopeAccessExpected, sql.ScopeAccess.String())
		}
	})
	t.Run("parallels", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sql, err := db.Smart(query)
				if err != nil {
					panic(err)
				}

				assert.Equal(t, expectedQuery, sql.Query)
				assert.Equal(t, expectedArg, sql.Args.String())
				assert.Equal(t, scopeAccessExpected, sql.ScopeAccess.String())
			}
		})
	})

}
func TestStatYearOfCreateUserAndRole(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	expectedquery := "SELECT count([user].[id]) [TotalUser], year([user].[created_on]) [Year], count([role].[id]) [TotalRole] FROM [sys_users] [user] join  [sys_roles] [role] ON [user].[role_id] = [role].[id] WHERE year([user].[created_on]) = @p1 GROUP BY year([user].[created_on]) HAVING count([user].[id]) > @p2"
	expectedArgs := `[
   2025,
   0
 ]`
	expectedOuputFields := `[
  {
    "Name": "TotalUser",
    "IsCalculated": true
  },
  {
    "Name": "Year",
    "IsCalculated": true
  },
  {
    "Name": "TotalRole",
    "IsCalculated": true
  }
]`
	expectedScopeAccess := `{
  "role.id": {
    "EntityName": "Role",
    "EntityFieldName": "Id"
  },
  "user.createdon": {
    "EntityName": "User",
    "EntityFieldName": "CreatedOn"
  },
  "user.id": {
    "EntityName": "User",
    "EntityFieldName": "Id"
  },
  "user.roleid": {
    "EntityName": "User",
    "EntityFieldName": "RoleId"
  }
}`
	query := `user(count(id) TotalUser,year(createdOn)+1 Year),
			  role(count(id) TotalRole),
			  
			  from(user,role,user.roleId=role.id),
			  where( year=? and totalUser>0)`

	sql, err := db.Smart(query, 2025)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
	assert.Equal(t, expectedquery, sql.Query)
	assert.Equal(t, expectedArgs, sql.Args.String())
	assert.Equal(t, expectedOuputFields, sql.OutputFields.String())
	assert.Equal(t, expectedScopeAccess, sql.ScopeAccess.String())

}
func BenchmarkStatYearOfCreateUserAndRolev1(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		b.Error(err)
	}
	defer db.Close()
	expectedquery := "SELECT count([user].[id]) [TotalUser], year([user].[created_on]) [Year], count([role].[id]) [TotalRole] FROM [sys_users] [user] join  [sys_roles] [role] ON [user].[role_id] = [role].[id] WHERE year([user].[created_on]) = @p1 GROUP BY year([user].[created_on]) HAVING count([user].[id]) > @p2"
	expectedArgs := `[
   2025,
   0
 ]`
	expectedOuputFields := `[
  {
    "Name": "TotalUser",
    "IsCalculated": true
  },
  {
    "Name": "Year",
    "IsCalculated": true
  },
  {
    "Name": "TotalRole",
    "IsCalculated": true
  }
]`
	expectedScopeAccess := `{
  "role.id": {
    "EntityName": "Role",
    "EntityFieldName": "Id"
  },
  "user.createdon": {
    "EntityName": "User",
    "EntityFieldName": "CreatedOn"
  },
  "user.id": {
    "EntityName": "User",
    "EntityFieldName": "Id"
  },
  "user.roleid": {
    "EntityName": "User",
    "EntityFieldName": "RoleId"
  }
}`
	query := `user(count(id) TotalUser,year(createdOn) Year),
			  role(count(id) TotalRole),
			  
			  from(user,role,user.roleId=role.id),
			  where( year=? and totalUser>0)`
	b.Run("Smart", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sql, err := db.Smart(query, 2025)
			if err != nil {
				panic(err)
			}
			assert.Equal(b, expectedquery, sql.Query)
			assert.Equal(b, expectedArgs, sql.Args.String())
			assert.Equal(b, expectedOuputFields, sql.OutputFields.String())
			assert.Equal(b, expectedScopeAccess, sql.ScopeAccess.String())
		}
	})
	b.Run("parallels", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				sql, err := db.Smart(query, 2025)
				if err != nil {
					panic(err)
				}
				assert.Equal(b, expectedquery, sql.Query)
				assert.Equal(b, expectedArgs, sql.Args.String())
				assert.Equal(b, expectedOuputFields, sql.OutputFields.String())
				assert.Equal(b, expectedScopeAccess, sql.ScopeAccess.String())
			}
		})
	})
}
