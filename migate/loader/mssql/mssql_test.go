package mssql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/db"
)

var sqlServerDns = "sqlserver://sa:123456@localhost?database=a0001&fetchSize=10000&encrypt=disable"

func TestMssql(t *testing.T) {
	Db, err := db.Open("sqlserver", sqlServerDns)
	assert.NoError(t, err)
	m := NewMssqlSchemaLoader()
	schema, err := m.LoadFullSchema(Db)
	assert.NoError(t, err)
	assert.NotNil(t, schema)
}
