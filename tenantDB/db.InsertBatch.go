package tenantDB

func (db *TenantDB) InsertBatch(data interface{}) error {
	return OnDbInsertBatchFunc(db, data)
	// return OnDbInsertBatchFunc(db, data)

}

type OnDbInsertBatchFuncType func(db *TenantDB, data interface{}) error

var OnDbInsertBatchFunc OnDbInsertBatchFuncType
