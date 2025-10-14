package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/db"
)

var pgDsn = "postgres://postgres:123456@localhost:5432/a001?sslmode=disable"

func TestPostgres(t *testing.T) {
	Db, err := db.Open("postgres", pgDsn)
	assert.NoError(t, err)
	m := NewPosgresSchemaLoader()
	schema, err := m.LoadFullSchema(Db, m.GetDefaultSchema())
	assert.NoError(t, err)
	assert.NotNil(t, schema)
}
