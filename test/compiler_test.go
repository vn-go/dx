package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/compiler"
	_ "github.com/vn-go/dx/compiler"
	_ "github.com/vn-go/dx/test/models" // registe all model
)

func TestCompilerSelectAllOneTable(t *testing.T) {
	ret, err := compiler.Compile("select * from User", "mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}
func BenchmarkCompilerSelectAllOneTable(t *testing.B) {
	for i := 0; i < t.N; i++ {
		compiler.Compile("select * from User", "mysql")
	}

}
func TestCompilerSelectAll2Table(t *testing.T) {
	ret, err := compiler.Compile("select concat(userId,?,code) as FName, username  from user", "mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}

func TestCompilerSelect1Table(t *testing.T) {
	ret, err := compiler.Compile("select User.*,Department.* from User,Department", "mysql")
	// expect := "`T1`.`id` `ID`,`T1`.`record_id` `RecordID`,`T1`.`created_at` `CreatedAt`,`T1`.`updated_at` `UpdatedAt`,`T1`.`user_id` `UserId`,`T1`.`email` `Email`,`T1`.`phone` `Phone`,`T1`.`username` `Username`,`T1`.`hash_password` `HashPassword`,`T1`.`description` `Description`"
	assert.NoError(t, err)
	t.Log(ret)
}
