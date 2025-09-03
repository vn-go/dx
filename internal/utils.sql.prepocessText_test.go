package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_preprocessText(t *testing.T) {
	txt1 := "LEN(Code+' '+Name)"
	eTxt1 := Utils.EXPR.QuoteExpression(txt1)
	assert.Equal(t, "LEN(`Code`+' '+`Name`)", eTxt1)
	txt2 := "aaa/bbb*cc+1-ddd7"
	eTxt2 := Utils.EXPR.QuoteExpression(txt2)
	assert.Equal(t, "`aaa`/`bbb`*`cc`+1-`ddd7`", eTxt2)

}
