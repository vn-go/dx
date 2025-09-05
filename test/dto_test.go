package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"

	"github.com/vn-go/dx/test/models"
)

func TestNewDTO(t *testing.T) {
	d, err := dx.NewDTO[models.Department]()
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}
func BenchmarkNewDTO(t *testing.B) {
	dx.NewDTO[models.Department]()
	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		dx.NewDTO[models.Department]()
	}

}
