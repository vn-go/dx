package postgres

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

var pgFunc = map[string]string{
	// Date & Time
	"day":    "EXTRACT(DAY FROM $1)", // dùng EXTRACT
	"month":  "EXTRACT(MONTH FROM $1)",
	"year":   "EXTRACT(YEAR FROM $1)",
	"hour":   "EXTRACT(HOUR FROM $1)",
	"minute": "EXTRACT(MINUTE FROM $1)",
	"second": "EXTRACT(SECOND FROM $1)",
	//"microsecond": "EXTRACT(MICROSECONDS FROM %s)", // PG có microseconds
	"date": "DATE(%s)",

	// String
	"len":    "LENGTH",
	"isnull": "COALESCE", // PG dùng COALESCE thay IFNULL
	//"concat":    "CONCAT",
	"upper":     "UPPER",
	"lower":     "LOWER",
	"left":      "LEFT",  // có
	"right":     "RIGHT", // có
	"substring": "SUBSTRING",

	// Arithmetic / Numeric
	"abs":      "ABS",
	"ceil":     "CEIL", // có alias CEILING
	"ceiling":  "CEILING",
	"floor":    "FLOOR",
	"round":    "ROUND",
	"truncate": "TRUNC", // PG dùng TRUNC
	"mod":      "MOD",   // có
	"sign":     "SIGN",
	"pi":       "PI",

	// Power & Logarithm
	"pow":   "POW", // hoặc POWER
	"power": "POWER",
	"sqrt":  "SQRT",
	"exp":   "EXP",
	"ln":    "LN",
	"log":   "LOG", // log base n (LOG(b, x)), khác MySQL
	"log10": "LOG(10, $1)",
	"log2":  "LOG(2, $1)",

	// Random
	"rand": "RANDOM()", // PG dùng RANDOM() thay RAND()

	// Trigonometric
	"sin":   "SIN",
	"cos":   "COS",
	"tan":   "TAN",
	"cot":   "COT", // PG không có trực tiếp -> 1/TAN(x), nếu muốn alias thì bỏ
	"asin":  "ASIN",
	"acos":  "ACOS",
	"atan":  "ATAN",
	"atan2": "ATAN2",
}
var pgAggregateFunc = map[string]string{
	// Core
	"count": "COUNT",
	"sum":   "SUM",
	"avg":   "AVG",
	"min":   "MIN",
	"max":   "MAX",

	// Statistical
	"std":         "STDDEV_POP", // PG không có alias STD, chuẩn là STDDEV_POP
	"stddev":      "STDDEV",     // alias chuẩn trong PG
	"stddev_pop":  "STDDEV_POP",
	"stddev_samp": "STDDEV_SAMP",
	"var_pop":     "VAR_POP",
	"var_samp":    "VAR_SAMP",
	"variance":    "VARIANCE", // có

	// Bitwise
	// "bit_and": "BIT_AND", // PG có
	// "bit_or":  "BIT_OR",  // PG có
	// "bit_xor": "",      // PG không có BIT_XOR

	// String aggregate
	"group_concat": "STRING_AGG", // thay GROUP_CONCAT bằng STRING_AGG(expr, delimiter)

	// JSON aggregate
	"json_arrayagg":  "JSON_AGG",        // PG dùng JSON_AGG
	"json_objectagg": "JSON_OBJECT_AGG", // PG >= 9.4 có

	// Special
	//"any_value": "ANY_VALUE", // PG 9.4+ có
}

func (d *postgresDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {
	fnName := strings.ToLower(delegator.FuncName)
	if fnName == "left" {
		return d.sqlLeftFunc(delegator)
	}
	if ret, ok := pgFunc[fnName]; ok {
		delegator.FuncName = ret
		for i := 0; i < len(delegator.Args); i++ {
			placeHilder := fmt.Sprintf("$%d", i+1)
			if strings.Contains(ret, placeHilder) {
				ret = strings.ReplaceAll(ret, fmt.Sprintf("$%d", i+1), delegator.Args[i])
				delegator.HandledByDialect = true
			}

		}

		return ret, nil
	}
	if ret, ok := pgAggregateFunc[fnName]; ok {
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
	switch strings.ToLower(delegator.FuncName) {
	case "len":
		delegator.FuncName = "LENGTH"
		delegator.HandledByDialect = true
		return "LENGTH" + "(" + strings.Join(delegator.Args, ", ") + ")", nil
	case "concat":
		delegator.HandledByDialect = true
		castArgs := make([]string, len(delegator.Args))
		for i, x := range delegator.Args {
			if x == "?" || x[0] == '$' {
				castArgs[i] = x + "::text"
			} else {
				castArgs[i] = x
			}
		}
		return "CONCAT" + "(" + strings.Join(castArgs, ", ") + ")", nil
	case "if":
		return d.SqlFunctionIf(delegator)
	default:

		if !d.isReleaseMode {
			panic(fmt.Sprintf("%s not implement at postgresDialect.SqlFunction, see %s", delegator.FuncName, `dialect\postgres\dialect.Postgres.SqlFunction.go`))
		} else {
			return "", fmt.Errorf("%s is not function", delegator.FuncName)
		}
	}
}
