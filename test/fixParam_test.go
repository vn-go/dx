package test

import (
	"fmt"
	"strings"
	"testing"
)

// InspectStringParam scans an SQL statement and extracts all string literals
// enclosed in single quotes ('...'). Each detected literal is replaced by a
// parameter placeholder `?`, and all extracted literal values are returned
// as a slice of strings.
//
// Example:
//
//	sql := "SELECT * FROM a='hello ''jony''' AND b='ok'"
//	query, params := InspectStringParam(sql)
//
//	// Result:
//	// query  -> "SELECT * FROM a=? AND b=?"
//	// params -> []string{"hello 'jony'", "ok"}
//
// Rules:
//  1. A literal string starts and ends with a single quote (').
//  2. Two consecutive single quotes (”) inside a literal represent an escaped quote (').
//  3. The function scans linearly from left to right (O(n) complexity).
//  4. Only literal strings are replaced — numbers, NULLs, or identifiers remain unchanged.
//
// Notes:
//   - This is not a full SQL parser, but it correctly handles common SQL literal patterns.
//   - Works reliably for most SQL dialects (MySQL, PostgreSQL, SQL Server, etc.).
func InspectStringParam(s string, args ...any) (string, []string) {
	var out strings.Builder // Holds the transformed SQL with ? placeholders
	var params []string     // Stores extracted string literal values
	inString := false       // Tracks whether the parser is currently inside a string literal
	var buf strings.Builder // Temporary buffer for building the current string literal
	index := len(args)
	for i := 0; i < len(s); i++ {
		ch := s[i]

		if ch == '\'' { // Found a single quote
			if !inString {
				// Entering a string literal
				inString = true
				buf.Reset()
			} else {
				// Already inside a literal
				if i+1 < len(s) && s[i+1] == '\'' {
					// Found an escaped single quote ('')
					buf.WriteByte('\'')
					i++ // Skip the next quote
				} else {
					// End of the current literal
					inString = false
					params = append(params, buf.String())
					fnx := fmt.Sprintf("get_params_info(%d)", index)
					out.WriteString(fnx) // Replace literal with placeholder
					index++
				}
			}
			continue
		}

		if inString {
			// Collect characters inside the string literal
			buf.WriteByte(ch)
		} else {
			// Copy non-literal characters to output as-is
			out.WriteByte(ch)
		}
	}

	return out.String(), params
}
func TestExtracArgs(t *testing.T) {
	sqlTest := map[string][]any{
		"select * from a='hello ''jony''' and b='ok' and c='Jack''s test'":                            []any{},
		"select concat(a,'') from a='hello ''jony''' and b='ok' and c='Jack''s test'":                 []any{},
		"select concat(a,'','hello''test''') from a='hello ''jony''' and b='ok' and c='Jack''s test'": []any{},
		"select concat(a,?,'hello''test''') from a='hello ''jony''' and b='ok' and c='Jack''s test'":  []any{"*"},
	}
	i := 0
	for k, v := range sqlTest {
		x, args := InspectStringParam(k, v...)
		fmt.Println(i, x, "\""+strings.Join(args, "\",\"")+"\"")
		i++
	}
}
