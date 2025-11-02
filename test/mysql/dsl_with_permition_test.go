package mysql

import (
	"testing"
	"time"

	"github.com/vn-go/dx"
)

func TestDls2(t *testing.T) {

	type Account struct {
		ID       int64
		Name     string `db:"size:50"`
		CreateOn time.Time
	}
	dx.Options.ShowSql = true
	dx.AddModels(Account{})
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	scopeAcess := dx.ScopeAccess.New("account")
	var accounts []Account
	err = db.DslQueryWithScopeAccess(scopeAcess, &accounts, "account(),where(id = ?)", 1)
	if err != nil {
		panic(err)
	}

}
