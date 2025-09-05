package dbutils

type dbUtilsType struct {
	Insert       *inserter
	ModelFactory *modelFacoryType
}

var DbUtils = &dbUtilsType{
	Insert:       &inserter{},
	ModelFactory: &modelFacoryType{},
}
