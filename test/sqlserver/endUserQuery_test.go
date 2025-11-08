package sqlserver

import (
	"fmt"
	"testing"

	"github.com/vn-go/dx"
)

func TestEndUserQuery(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	qr := dx.NewEndUserQuery()
	qr = qr.Fields("user(concat(username, '''123'+?) Username),id", 1)
	// qr = qr.From("user")
	qr = qr.Filter("id = ?", 1)

	fe := qr.ToFrontEnd(db)
	fe = fe.Select("id, len(username+'''1234') Len")
	fe = fe.Filter("Len > ? and (count(id)>? or sum(len(username))=0)", 5)
	query, args, err := fe.ToSql()
	if err != nil {
		panic(err)
	}
	fmt.Println(query, args)

}
