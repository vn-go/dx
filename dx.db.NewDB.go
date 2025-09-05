package dx

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/dialect/factory"
)

type initCreateDBNoCache struct {
	once sync.Once

	err error
	db  *DB
}

var cacheCreateDB = sync.Map{}

func (db *DB) NewDB(dbName string) (*DB, error) {
	key := fmt.Sprintf("%s:%s", dbName, db.DriverName)
	actual, _ := cacheCreateDB.LoadOrStore(key, &initCreateDBNoCache{})
	init := actual.(*initCreateDBNoCache)
	init.once.Do(func() {

		dialect := factory.DialectFactory.Create(db.DriverName)
		dsn, err := dialect.NewDataBase(db.DB, dbName)
		if err != nil {
			init.err = err
			return
		}

		ret, err := Open(db.DriverName, dsn)
		if err != nil {
			init.err = err
			return
		}

		init.db = ret

	})
	return init.db, init.err

}
