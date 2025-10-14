package oracle

import (
	"fmt"
	"testing"

	_ "github.com/sijms/go-ora/v2" // đăng ký driver
	"github.com/vn-go/dx"
)

var oracleDsn = "oracle://system:123456@localhost:1521/FREEPDB1"

func TestConnectOracle(t *testing.T) {
	db, err := dx.Open("oracle", oracleDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var sysdate string
	err = db.QueryRow("SELECT TO_CHAR(SYSDATE, 'YYYY-MM-DD HH24:MI:SS') FROM dual").Scan(&sysdate)
	if err != nil {
		panic(err)
	}

	fmt.Println("Current time:", sysdate)
}
