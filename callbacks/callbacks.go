package callbacks

import (
	"gorm.io/gorm"
)

type Config struct {
	LastInsertIDReversed bool
	WithReturning        bool
}

// 注册回调到mysql 驱动
func RegisterDefaultCallbacks(db *gorm.DB, config *Config) {
	enableTransaction := func(db *gorm.DB) bool {
		return !db.SkipDefaultTransaction
	}

	//insert into
	createCallback := db.Callback().Create()
	createCallback.Match(enableTransaction).Register("gorm:begin_transaction", BeginTransaction)
	createCallback.Register("gorm:before_create", BeforeCreate)
	createCallback.Register("gorm:save_before_associations", SaveBeforeAssociations)
	createCallback.Register("gorm:create", Create(config))
	createCallback.Register("gorm:save_after_associations", SaveAfterAssociations)
	createCallback.Register("gorm:after_create", AfterCreate)
	createCallback.Match(enableTransaction).Register("gorm:commit_or_rollback_transaction", CommitOrRollbackTransaction)

	//select
	queryCallback := db.Callback().Query()
	queryCallback.Register("gorm:query", Query)
	queryCallback.Register("gorm:preload", Preload) //预加载后续数据
	queryCallback.Register("gorm:after_query", AfterQuery)

	//delete
	deleteCallback := db.Callback().Delete()
	deleteCallback.Match(enableTransaction).Register("gorm:begin_transaction", BeginTransaction)
	deleteCallback.Register("gorm:before_delete", BeforeDelete)
	deleteCallback.Register("gorm:delete", Delete)
	deleteCallback.Register("gorm:after_delete", AfterDelete)
	deleteCallback.Match(enableTransaction).Register("gorm:commit_or_rollback_transaction", CommitOrRollbackTransaction)

	// update
	updateCallback := db.Callback().Update()
	updateCallback.Match(enableTransaction).Register("gorm:begin_transaction", BeginTransaction)
	updateCallback.Register("gorm:setup_reflect_value", SetupUpdateReflectValue)
	updateCallback.Register("gorm:before_update", BeforeUpdate)
	updateCallback.Register("gorm:save_before_associations", SaveBeforeAssociations)
	updateCallback.Register("gorm:update", Update)
	updateCallback.Register("gorm:save_after_associations", SaveAfterAssociations)
	updateCallback.Register("gorm:after_update", AfterUpdate)
	updateCallback.Match(enableTransaction).Register("gorm:commit_or_rollback_transaction", CommitOrRollbackTransaction)

	//r
	db.Callback().Row().Register("gorm:raw", RowQuery)

	db.Callback().Raw().Register("gorm:raw", RawExec)
}
