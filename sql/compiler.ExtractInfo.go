package sql

import (
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type extractInfoReqiureAcessScope struct {
	Field   refFields
	Hash256 string
}
type extractInfoOutputField struct {
	OutputFields   outputFields
	Hash256        string
	OutputFieldMap map[string]OutputField
}
type extractInfoAccessScopeCheck struct {
	Scope   accessScopes
	Hash256 string
}
type ExtractInfo struct {
	SelectStatement   types.SelectStatement
	RequireAcessScope ExtractInfoReqiureAcessScope
	OuputInfo         extractInfoOutputField
	ScopeCheck        extractInfoAccessScopeCheck
	Args              arguments
}
type ExtractInfoReqiureAcessScope extractInfoReqiureAcessScope
type ExtractInfoOutputField extractInfoOutputField

func (e *ExtractInfoOutputField) NewOutputFields() outputFields {
	return outputFields{}
}

var extractInfoPool = sync.Pool{
	New: func() any {
		return new(ExtractInfo)
	},
}

func (e *ExtractInfo) Clone() *ExtractInfo {
	ret := extractInfoPool.Get().(*ExtractInfo)
	ret.Args = make(arguments, len(e.Args))
	copy(ret.Args, e.Args)
	ret.SelectStatement = types.SelectStatement{
		Source:   e.SelectStatement.Source,
		Selector: e.SelectStatement.Selector,
		Filter:   e.SelectStatement.Filter,
		Sort:     e.SelectStatement.Sort,
		Having:   e.SelectStatement.Having,
		GroupBy:  e.SelectStatement.GroupBy,
		Offset:   e.SelectStatement.Offset,
		Limit:    e.SelectStatement.Limit,
	}
	ret.RequireAcessScope = e.RequireAcessScope
	ret.OuputInfo = e.OuputInfo
	ret.ScopeCheck = e.ScopeCheck
	return ret
}

// compiler.ExtractInfo.go
type initExtractInfo struct {
	val  *ExtractInfo
	err  error
	once sync.Once
}

var initExtractInfoMap sync.Map

func (c compiler) ExtractInfo(dialect types.Dialect, dlsQuery string, args []any) (*ExtractInfo, error) {
	a, _ := initExtractInfoMap.LoadOrStore(dlsQuery, &initExtractInfo{})
	i := a.(*initExtractInfo)
	i.once.Do(func() {
		i.val, i.err = c.extractInfo(dialect, dlsQuery)
	})
	return i.val, i.err
}

func (c compiler) extractInfo(dialect types.Dialect, query string) (*ExtractInfo, error) {
	var err error
	//var node sqlparser.SQLNode
	var sqlStm sqlparser.Statement
	isNotStartWithSelect := !c.startWithSelectKeyword(query)
	reIndex := []int{}
	querySimple := ""
	textParams := []string{}
	if isNotStartWithSelect {
		querySimple, reIndex, textParams, err = smartier.simple(query)
		if err != nil {
			return nil, err
		}

		query = querySimple

		// offset = querySimple.offset
		// limit = querySimple.limit
	}
	inputSql := internal.Helper.ReplaceQuestionMarks(query, GET_PARAMS_FUNC)

	queryCompiling, _ := internal.Helper.InspectStringParam(inputSql)
	injector := newInjector(dialect, textParams)
	if !isNotStartWithSelect {
		queryCompiling, err = internal.Helper.QuoteExpression(queryCompiling)
	}
	//
	if err != nil {
		return nil, err
	}
	sqlStm, err = sqlparser.Parse(queryCompiling)
	if err != nil {
		return nil, err
	}
	selectStatementResult, err := froms.getSelectStatement(sqlStm, injector, CMP_SELECT)
	if err != nil {
		return nil, err
	}
	ret := selectStatementResult.compilerResult
	ret.reIndex = reIndex

	// for j := 0; j < len(ret.Args); j++ {
	// 	if j < len(ret.reIndex) {
	// 		selectStatementResult.compilerResult.Args[j].index = ret.reIndex[j] + 1
	// 	}

	// }
	ret.AccessScope = accessScopes{}
	itemsForHash256Key := []string{}
	for k, v := range ret.Fields {
		if _, ok := ret.AccessScope[strings.ToLower(v.EntityName)]; !ok {
			ret.AccessScope[strings.ToLower(v.EntityName)] = map[string]string{}
		}
		ret.AccessScope[strings.ToLower(v.EntityName)][strings.ToLower(v.EntityFieldName)] = k
		itemsForHash256Key = append(itemsForHash256Key, k)
	}
	ret.Hash256AccessScope = internal.Helper.Hash256(strings.Join(itemsForHash256Key, ""))
	for i := 0; i < len(ret.OutputFields); i++ {
		ret.OutputFields[i].DbType = internal.Helper.GetSqlTypeFfromGoType(ret.OutputFields[i].FieldType)
	}
	ret.Hash256OutputFields = ret.OutputFields.ToHas256Key()
	ouputInfoMap := map[string]OutputField{}
	for i := 0; i < len(ret.OutputFields); i++ {
		ouputInfoMap[strings.ToLower(ret.OutputFields[i].Name)] = ret.OutputFields[i]
	}
	retExtractInfo := &ExtractInfo{
		SelectStatement: selectStatementResult.selectStatement,

		RequireAcessScope: ExtractInfoReqiureAcessScope{
			Field:   ret.Fields,
			Hash256: ret.Hash256AccessScope,
		},
		OuputInfo: extractInfoOutputField{
			OutputFields:   ret.OutputFields,
			Hash256:        ret.Hash256OutputFields,
			OutputFieldMap: ouputInfoMap,
		},
		ScopeCheck: extractInfoAccessScopeCheck{
			Scope:   ret.AccessScope,
			Hash256: ret.Hash256AccessScope,
		},
		Args: ret.Args,
	}
	return retExtractInfo, nil
}
