package expr

import (
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) FuncExpr(context *exprCompileContext, expr *sqlparser.FuncExpr) (string, error) {
	strArgs := []string{}
	backup_purpose := context.Purpose
	context.Purpose = BUILD_FUNC
	defer func() {
		context.Purpose = backup_purpose
	}()

	for _, arg := range expr.Exprs {
		argStr, err := compiler.compile(context, arg)
		if err != nil {
			context.Purpose = backup_purpose
			return "", err
		}
		strArgs = append(strArgs, argStr)
	}
	dialectDelegateFunction := internal.DialectDelegateFunction{
		FuncName:         expr.Name.String(),
		Args:             strArgs,
		HandledByDialect: false,
	}

	ret, err := context.dialect.SqlFunction(&dialectDelegateFunction)
	if err != nil {

		return "", err
	}
	if dialectDelegateFunction.HandledByDialect {

		return ret, nil
	}

	retTxt := dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")"
	if aliasField, ok := context.stackAliasFields.Pop(); ok {
		retTxt = retTxt + " AS " + context.dialect.Quote(aliasField)

	}

	return retTxt, nil

}
