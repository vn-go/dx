package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestFindbyWherePg(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	total := uint64(0)
	err = db.Model(&models.User{}).Where("username!=?", "admin").Count(&total)
	assert.NoError(t, err)
	err = db.Where("username!=?", "admin").Limit(100).Order("Id desc").Find(&user)

	assert.NoError(t, err)
}
func BenchmarkFindbyWhereNoCacheV1PG(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	total := uint64(0)
	err = db.Model(&models.User{}).Where("username!=?", 25).Count(&total)
	assert.NoError(t, err)
	for i := 0; i < t.N; i++ {
		db.Where("username!=?", "admin").Limit(100).Order("Id desc").Find(&user)
	}

}
func TestFindbyWhereLimit100RowsMysqlV2PG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	total := uint64(0)
	err = db.Model(&models.User{}).Where("username!=?", 25).Count(&total)
	assert.NoError(t, err)
	err = db.Where("username!=?", "admin").Limit(100).Order("Id desc").Find(&user)

	assert.NoError(t, err)
}
func BenchmarkFindbyWhereLimit100RowsMysqlV1PG(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.WithContext(t.Context()).Where("username!=?", "admin").Limit(100).Order("Id desc").Find(&user)

	}

}
func BenchmarkFindbyWhereLimit100RowsMysqlV2PG(t *testing.B) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.WithContext(t.Context()).Where("username!=?", "admin").Limit(100).Order("Id desc").Find(&user)

	}

}
func TestFindbyWhereMysqlWithTransPG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	tx := db.WithContext(t.Context()).Begin()
	if tx.Error != nil {
		t.Log(tx.Error)
	}

	var users []models.User
	w := tx.Where("username != ?", "admin")
	if err := w.Find(&users); err != nil {
		// Rollback nếu có lỗi
		err := tx.Rollback()
		t.Log(err)
	}

	// 4. Commit transaction sau khi tất cả các thao tác đã thành công
	if err := tx.Commit(); err != nil {
		t.Log(err)
	}

}
func BenchmarkFindbyWhereMysqlWithTransPG(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	tx := db.WithContext(t.Context()).Begin()
	if tx.Error != nil {
		t.Log(tx.Error)
	}
	for i := 0; i < t.N; i++ {
		var users []models.User
		w := tx.Where("username != ?", "admin").Limit(100).Order("Id desc")
		if err := w.Find(&users); err != nil {
			// Rollback nếu có lỗi
			tx.Rollback()

		} else {
			tx.Commit()
		}
	}

}
func TestDbLimitMysqlPG(t *testing.T) {
	dx.Options.ShowSql = true
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User
	err = db.Offset(1000).Limit(100).Order("userId desc").Find(&users)
	t.Log(users)
	assert.NoError(t, err)
}
func BenchmarkDbLimitPG(t *testing.B) {

	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.Offset(1000).Limit(100).Order("userId desc").Find(&users)
	}

	//t.Log(users)
}
func TestSelectPG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users = []models.User{}
	fx := db.Select(
		"username",
		"email",
	).Find(&users)
	t.Log(fx)
}
func BenchmarkSelecPG(t *testing.B) {

	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	//fx := db.Select("amount+?+size", "concat(firstName,?,lastName) FullName", "code", 1, " ")
	var users = []models.User{}
	for i := 0; i < t.N; i++ {
		db.Select(
			"username",
			"email",
		).Find(&users)
	}

}
func TestSelectUserAndEmailPG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users = &[]struct {
		Username string
		Email    *string
	}{}
	fxx := db.Model(&models.User{}).Select(
		"concat(username,?,username) username",
		"email", " ",
	)
	sql, _, err := fxx.GetSQL(*fxx.GetModelType())
	t.Log(err)
	fmt.Print(sql)
	fx := db.Model(&models.User{}).Select(
		"concat(username,?,username) username",
		"email", " ",
	).Find(users)
	t.Log(fx)
}
func BenchmarkSelectUserAndEmailPG(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users = &[]struct {
		Username string
		Email    *string
	}{}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.Model(&models.User{}).Limit(100).Select(
			"concat(username,?,username) username",
			"email", " ",
		).Find(users)
	}

}
func TestWithContextSelectUserAndEmailPG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users = &[]struct {
		Username string
		Email    *string
	}{}

	fx := db.WithContext(t.Context()).Model(&models.User{}).Select(
		"concat(username,?,username) username",
		"email", " ",
	).Find(users)
	t.Log(fx)
}
func BenchmarkWithContextSelectUserAndEmailPG(t *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users = &[]struct {
		Username string
		Email    *string
	}{}
	for i := 0; i < t.N; i++ {
		db.WithContext(t.Context()).Model(&models.User{}).Select(
			"concat(username,?,username) username",
			"email", " ",
		).Find(users)
	}

}
func TestTransSelectUserAndEmailPG(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	tx := db.WithContext(t.Context()).Begin()
	assert.NoError(t, tx.Error)
	var users = &[]struct {
		Username string
		Email    *string
	}{}

	fx := tx.Model(&models.User{}).Select(
		"concat(username,?,username) username",
		"email", " ",
	).Find(users)
	t.Log(fx)
}
func TestTransFirstUserAndEmailPG(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	tx := db.WithContext(t.Context()).Begin()
	assert.NoError(t, tx.Error)
	var user = &struct {
		Username string
		Email    *string
	}{}

	fx := tx.Model(&models.User{}).Select(
		"concat(username,?,username) username",
		"email", " ",
	).First(user)
	t.Log(fx)
}
func TestFnTransFirstUserAndEmailPG(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	tx := db.WithContext(t.Context()).Begin()
	assert.NoError(t, tx.Error)
	var user = &struct {
		Username string
		Email    *string
	}{}
	
	err = db.WithContext(t.Context()).Transaction(nil, func(tx *dx.Tx) error {
		fx := tx.Model(&models.User{}).Select(
			"concat(username,?,username) username",
			"email", " ",
		).First(user)
		t.Log(fx)
		return fx
	})
	assert.NoError(t, err)

}
