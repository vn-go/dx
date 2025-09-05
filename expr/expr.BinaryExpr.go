package expr

import (
	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) BinaryExpr(context *exprCompileContext, expr *sqlparser.BinaryExpr) (string, error) {
	fieldAlias := ""
	if context.purpose == build_purpose_select {
		if _fieldAlias, ok := context.stackAliasFields.Pop(); ok {
			fieldAlias = _fieldAlias
		}
	}
	backupPurpose := context.purpose
	context.purpose = build_purpose_for_function

	defer func() {
		context.purpose = backupPurpose
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
		ret = ret + " AS " + context.dialect.Quote(fieldAlias)
	}
	return ret, nil

}
