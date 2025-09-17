package dx

// type DbContext[TModelSet any] struct {
// }

// type dbModelContext[T any] struct {
// 	db     *DB
// 	Models T
// }

// type initRegisterAll struct {
// 	once sync.Once
// }

// var cacheRegisterAll sync.Map
// var mapRegisterAll map[reflect.Type]bool = map[reflect.Type]bool{}

// func (r *DbContext[T]) registerAll() {
// 	typ := reflect.TypeFor[T]()
// 	if _, ok := mapRegisterAll[typ]; ok {
// 		return
// 	}
// 	actually, _ := cacheRegisterAll.LoadOrStore(typ, &initRegisterAll{})
// 	init := actually.(*initRegisterAll)
// 	init.once.Do(func() {
// 		for i := 0; i < typ.NumField(); i++ {
// 			field := typ.Field(i)
// 			if field.Anonymous {
// 				continue
// 			}
// 			modelRegistry.RegisterType(field.Type)
// 		}
// 		mapRegisterAll[typ] = true
// 	})
// }

// type dbContextGetKey struct {
// 	typ      reflect.Type
// 	dbName   string
// 	dbDriver string
// }
// type initDbContextGet[T any] struct {
// 	val  *dbModelContext[T]
// 	once sync.Once
// }

// var cacheDbContextGet sync.Map

// func (r *DbContext[T]) Get(db *DB) *dbModelContext[T] {
// 	typ := reflect.TypeFor[T]()
// 	key := dbContextGetKey{
// 		typ:      typ,
// 		dbName:   db.DbName,
// 		dbDriver: db.DriverName,
// 	}
// 	actually, _ := cacheDbContextGet.LoadOrStore(key, &initDbContextGet[T]{})
// 	init := actually.(*initDbContextGet[T])
// 	init.once.Do(func() {
// 		ret := &dbModelContext[T]{
// 			db: db,
// 		}
// 		init.val = ret
// 	})
// 	return init.val
// }
// func (ctx *dbModelContext[T]) First(model any, args ...interface{}) error {
// 	return ctx.db.First(model, args...)
// }
