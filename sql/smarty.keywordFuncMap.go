package sql

var sqlFuncWhitelist = map[string]bool{
	// --- Numeric functions ---
	"abs":     true,
	"ceil":    true,
	"ceiling": true,
	"floor":   true,
	"round":   true,
	"power":   true,
	"mod":     true,
	"sqrt":    true,

	// --- String functions ---
	"lower":       true,
	"upper":       true,
	"length":      true,
	"char_length": true,
	"trim":        true,
	"substring":   true,
	"concat":      true,
	"replace":     true,

	// --- Date/Time functions ---
	"current_date":      true,
	"current_timestamp": true,
	"extract":           true,
	"date_part":         true, // PostgreSQL compatible
	"year":              true,
	"month":             true,
	"day":               true,

	// --- Conditional / Null handling ---
	"coalesce": true,
	"nullif":   true,

	// --- Type conversion ---
	"cast": true,

	// --- Safe aggregates ---
	"sum":   true,
	"avg":   true,
	"min":   true,
	"max":   true,
	"count": true,
}

var keywordFuncMap = map[string]bool{
	"from":    true,
	"where":   true,
	"sort":    true,
	"limit":   true,
	"offset":  true,
	"group":   true,
	"subsets": true,
	"rowset":  true,

	"union": true,
	//"count": true, "sum": true, "avg": true, "min": true, "max": true,
}
var customKeyworMap = map[string]bool{
	"if":     true,
	"left":   true,
	"right":  true,
	"concat": true,
}
var keywordSQLStandard = map[string]bool{
	// Query structure
	"from": true, "where": true, "group": true, "having": true,
	"order": true, "by": true, "limit": true, "offset": true, "distinct": true,
	"union": true, "intersect": true, "except": true,

	// DDL
	"create": true, "alter": true, "drop": true, "table": true, "index": true,
	"view": true, "schema": true, "database": true,

	// DML
	"insert": true, "into": true, "values": true, "update": true,
	"set": true, "delete": true,

	// Conditions
	"and": true, "or": true, "not": true, "in": true, "like": true,
	"between": true, "is": true, "exists": true,

	// Joins
	"join": true, "inner": true, "left": true, "right": true,
	"full": true, "outer": true, "on": true, "using": true, "as": true,

	// Aggregates
	"count": true, "sum": true, "avg": true, "min": true, "max": true,

	// Others
	"case": true, "when": true, "then": true, "else": true, "end": true,
	"asc": true, "desc": true, "all": true, "any": true, "some": true,
	"cast": true, "convert": true, "null": true, "true": true, "false": true,
}

// func init() {
// 	keywordFuncMap = internal.UnionMap(keywordFuncMap, sqlFuncWhitelist)
// 	keywordFuncMap = internal.UnionMap(keywordFuncMap, customKeyworMap)
// }
