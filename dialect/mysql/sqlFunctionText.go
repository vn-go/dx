package mysql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

// sqlFunctionText.go

func (mysql *mySqlDialect) SqlFunctionText(delegator *types.DialectDelegateFunction) (string, error) {
	if len(delegator.Args) == 1 {
		return mysql.SqlFunctionTextOneArg(delegator)
	} else if len(delegator.Args) == 2 {
		/*
			MySQL:
			SELECT FORMAT(price, 2) AS FormattedPrice FROM items;
			hoặc: FORMAT(price, 'de_DE') cho định dạng vùng miền
		*/
		field := delegator.Args[0]
		formatStr := delegator.Args[1]
		delegator.HandledByDialect = true
		return fmt.Sprintf("FORMAT(%s, %s)", field, formatStr), nil
	} else {
		return "", types.NewCompilerError("TEXT() function requires one or two arguments.")
	}
}

func (mysql *mySqlDialect) SqlFunctionTextOneArg(delegator *types.DialectDelegateFunction) (string, error) {
	delegator.HandledByDialect = true
	// MySQL: CAST(expr AS CHAR) hoặc CAST(expr AS CHAR(N)) hoặc CAST(expr AS BINARY)
	return fmt.Sprintf("CAST(%s AS CHAR)", delegator.Args[0]), nil
}
