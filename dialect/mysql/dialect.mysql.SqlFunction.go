package mysql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

var castFunc = map[string]string{
	// Date & Time
	"day":    "DAY",
	"month":  "MONTH",
	"year":   "YEAR",
	"hour":   "HOUR",
	"minute": "MINUTE", // sửa "minutes" -> "minute" chuẩn hơn
	"second": "SECOND",
	//"microsecond": "MICROSECOND",
	"date": "DATE",

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
var aggregateFunc = map[string]string{
	// Core
	"count": "COUNT",
	"sum":   "SUM",
	"avg":   "AVG",
	"min":   "MIN",
	"max":   "MAX",

	// Statistical
	"std":         "STD",    // alias của STDDEV_POP
	"stddev":      "STDDEV", // alias của STDDEV_POP
	"stddev_pop":  "STDDEV_POP",
	"stddev_samp": "STDDEV_SAMP",
	"var_pop":     "VAR_POP",
	"var_samp":    "VAR_SAMP",
	"variance":    "VARIANCE", // alias của VAR_POP

	// Bitwise
	// "bit_and": "BIT_AND",
	// "bit_or":  "BIT_OR",
	// "bit_xor": "BIT_XOR",

	// String aggregate
	"group_concat": "GROUP_CONCAT",

	// JSON aggregate
	"json_arrayagg":  "JSON_ARRAYAGG",
	"json_objectagg": "JSON_OBJECTAGG",

	// // Special
	// "any_value": "ANY_VALUE",
}

func (d *mySqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {
	fnName := strings.ToLower(delegator.FuncName)
	if fnName == "if" {
		return d.SqlFunctionIf(delegator)
	}
	if strings.Contains(fnName, ".") {
		items := strings.Split(fnName, ".")
		if items[0] == "list" {
			return d.SqlFunctionResolveArrayFunctions(items[1], delegator)
		}
	}
	if fnName == "text" {
		return d.SqlFunctionText(delegator)
	}
	if ret, ok := castFunc[fnName]; ok {
		delegator.FuncName = ret
		return ret, nil
	}
	if ret, ok := aggregateFunc[fnName]; ok {
		delegator.FuncName = ret
		delegator.IsAggregate = true
		return ret, nil
	}
	if fnName == "countall" {
		delegator.FuncName = "count(*)"
		delegator.IsAggregate = true
		delegator.HandledByDialect = true
		return "count(*)", nil
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


