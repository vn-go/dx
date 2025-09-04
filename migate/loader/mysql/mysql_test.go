package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/db"
)

var mySqlDsn = "root:123456@tcp(127.0.0.1:3306)/a001"

func TestMisql(t *testing.T) {
	Db, err := db.Open("mysql", mySqlDsn)
	assert.NoError(t, err)
	m := NewMysqlSchemaLoader(Db)
	schema, err := m.LoadFullSchema()
	assert.NoError(t, err)
	assert.NotNil(t, schema)
}
