package mssql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

var castFuncMSSQL = map[string]string{
	// Date & Time
	"day":    "DAY",
	"month":  "MONTH",
	"year":   "YEAR",
	"hour":   "HOUR",
	"minute": "MINUTE", // MSSQL có
	"second": "SECOND",
	//"microsecond": "",     // ❌ MSSQL không có MICROSECOND (chỉ hỗ trợ đến MILLISECOND)
	"date": "CAST", // ❌ không có hàm DATE(); phải CAST(expr AS DATE)

	// String
	"len":       "LEN",    // MSSQL dùng LEN thay vì LENGTH
	"isnull":    "ISNULL", // MSSQL có ISNULL(expr, value)
	"concat":    "CONCAT", // MSSQL 2012+ có CONCAT()
	"upper":     "UPPER",
	"lower":     "LOWER",
	"left":      "LEFT",
	"right":     "RIGHT",
	"substring": "SUBSTRING",

	// Arithmetic / Numeric
	"abs":      "ABS",
	"ceil":     "CEILING", // MSSQL không có CEIL, chỉ CEILING
	"ceiling":  "CEILING",
	"floor":    "FLOOR",
	"round":    "ROUND",
	"truncate": "ROUND",     // ❌ MSSQL không có TRUNCATE(); dùng ROUND(x, n, 1)
	"mod":      "{$1}%{$2}", // ❌ không có MOD(); dùng toán tử %
	"sign":     "SIGN",
	"pi":       "PI",

	// Power & Logarithm
	"pow":   "POWER",
	"power": "POWER",
	"sqrt":  "SQRT",
	"exp":   "EXP",
	"ln":    "LOG", // ❌ MSSQL không có LN(); LOG(expr) = ln(expr)
	"log":   "LOG", // LOG(expr) = ln(expr), LOG(expr, base) cũng có
	"log10": "LOG10",
	"log2":  "LOG({$1}, 2)", // ❌ MSSQL không có LOG2(), phải dùng LOG(expr, 2)

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
	"atan2": "ATN2", // MSSQL dùng ATN2(y, x) thay vì ATAN2(y, x)
}

var aggregateFuncMSSQL = map[string]string{
	// Core (có đầy đủ)
	"count": "COUNT",
	"sum":   "SUM",
	"avg":   "AVG",
	"min":   "MIN",
	"max":   "MAX",

	// Statistical (SQL Server dùng STDEV/STDEVP, VAR/VARP)
	"std":         "STDEVP", // gần giống STD (population)
	"stddev":      "STDEVP", // alias của STDDEV_POP
	"stddev_pop":  "STDEVP", // population standard deviation
	"stddev_samp": "STDEV",  // sample standard deviation
	"var_pop":     "VARP",   // population variance
	"var_samp":    "VAR",    // sample variance
	"variance":    "VARP",   // gần giống VARIANCE (population)

	// // Bitwise (SQL Server không có aggregate BIT_*)
	// // -> Để trống hoặc sau này có thể fake bằng SUM/CASE
	// "bit_and": "",
	// "bit_or":  "",
	// "bit_xor": "",

	// String aggregate
	"group_concat": "STRING_AGG", // từ SQL Server 2017 trở lên

	// JSON aggregate (SQL Server dùng FOR JSON PATH, không phải hàm aggregate)
	"json_arrayagg":  "",
	"json_objectagg": "",

	// // Special
	// "any_value": "", // SQL Server không có, có thể dùng MIN/MAX thay thế
}

func (d *mssqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {

	if !d.isReleaseMode {
		panic(fmt.Sprintf("%s not implement at mssqlDialect.SqlFunction, see %s", delegator.FuncName, `dialect\mssql\dialect.mssql.SqlFunction.go`))
	} else {
		return "", fmt.Errorf("%s is not function", delegator.FuncName)
	}
	// //delegator.Approved = true
	// delegator.FuncName = strings.ToUpper(delegator.FuncName)
	// return "", nil
}
