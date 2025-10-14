package oracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/db"
)

var dsn = "oracle://system:123456@localhost:1521/FREEPDB1"

func TestPostgres(t *testing.T) {
	Db, err := db.Open("oracle", dsn)
	assert.NoError(t, err)
	m := NewOracleSchemaLoader()
	schema, err := m.LoadFullSchema(Db, "app")
	assert.NoError(t, err)
	assert.NotNil(t, schema)
}
