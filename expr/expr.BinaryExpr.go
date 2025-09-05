package expr

import (
	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) BinaryExpr(context *exprCompileContext, expr *sqlparser.BinaryExpr) (string, error) {
	fieldAlias := ""
	if context.Purpose == BUILD_SELECT {
		if _fieldAlias, ok := context.stackAliasFields.Pop(); ok {
			fieldAlias = _fieldAlias
		}
	}
	backupPurpose := context.Purpose
	context.Purpose = BUILD_FUNC

	defer func() {
		context.Purpose = backupPurpose
	}()
	left, err := compiler.compile(context, expr.Left)
	if err != nil {
		return "", err
	}
	right, err := compiler.compile(context, expr.Right)
	if err != nil {
		return "", err
	}

	ret := left + " " + expr.Operator + " " + right
	if fieldAlias != "" {
		ret = ret + " AS " + context.Dialect.Quote(fieldAlias)
	}
	return ret, nil

}
