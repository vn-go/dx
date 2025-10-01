package mysql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

var castFunc = map[string]string{
	// Date & Time
	"day":         "DAY",
	"month":       "MONTH",
	"year":        "YEAR",
	"hour":        "HOUR",
	"minute":      "MINUTE", // sửa "minutes" -> "minute" chuẩn hơn
	"second":      "SECOND",
	"microsecond": "MICROSECOND",
	"date":        "DATE",

	// String
	"len":       "LENGTH",
	"isnull":    "IFNULL",
	"concat":    "CONCAT",
	"upper":     "UPPER",
	"lower":     "LOWER",
	"left":      "LEFT",
	"right":     "RIGHT",
	"substring": "SUBSTRING",

	// Arithmetic / Numeric
	"abs":      "ABS",
	"ceil":     "CEIL", // hoặc CEILING
	"ceiling":  "CEILING",
	"floor":    "FLOOR",
	"round":    "ROUND",
	"truncate": "TRUNCATE",
	"mod":      "MOD",
	"sign":     "SIGN",
	"pi":       "PI",

	// Power & Logarithm
	"pow":   "POW",
	"power": "POWER",
	"sqrt":  "SQRT",
	"exp":   "EXP",
	"ln":    "LN",
	"log":   "LOG",
	"log10": "LOG10",
	"log2":  "LOG2",

	// Random
	"rand": "RAND",

	// Trigonometric
	"sin":   "SIN",
	"cos":   "COS",
	"tan":   "TAN",
	"cot":   "COT",
	"asin":  "ASIN",
	"acos":  "ACOS",
	"atan":  "ATAN",
	"atan2": "ATAN2",
}

func (d *mySqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {
	fnName := strings.ToLower(delegator.FuncName)
	if ret, ok := castFunc[fnName]; ok {
		delegator.FuncName = ret
		return ret, nil
	}
	switch fnName {
	case "now":
		delegator.HandledByDialect = true
		return "NOW()", nil

	default:
		if !d.isReleaseMode {
			panic(fmt.Sprintf("%s not implement at mySqlDialect.SqlFunction, see %s", delegator.FuncName, `dialect\mysql\dialect.mysql.SqlFunction.go`))
		} else {
			return "", fmt.Errorf("%s is not function", delegator.FuncName)
		}
		//return "", nil
	}

}
