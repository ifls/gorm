package gorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Config GORM config
type Config struct {
	// GORM perform single create, update, delete operations in transactions by default to ensure database data integrity
	// You can disable it by setting `SkipDefaultTransaction` to true
	SkipDefaultTransaction bool //禁用每条更新语句都开一个事务, 读语句不许开事务
	// NamingStrategy tables, columns naming strategy
	NamingStrategy schema.Namer //命名策略
	// Logger
	Logger logger.Interface
	// NowFunc the function to be used when creating a new timestamp
	NowFunc func() time.Time
	// DryRun generate sql without execute
	DryRun bool //预检
	// PrepareStmt executes the given query in cached statement
	PrepareStmt bool //使用预处理语句执行
	// DisableAutomaticPing
	DisableAutomaticPing bool // 不自动 ping
	// DisableForeignKeyConstraintWhenMigrating
	DisableForeignKeyConstraintWhenMigrating bool

	// ClauseBuilders clause builder
	ClauseBuilders map[string]clause.ClauseBuilder //子句构建器
	// ConnPool db conn pool
	ConnPool ConnPool //可能就是 *sql.DB, 也可能是一个 struct 里面内嵌 *sql.DB
	// Dialector database dialector
	Dialector
	// Plugins registered plugins
	Plugins map[string]Plugin

	callbacks  *callbacks
	cacheStore *sync.Map
}

// DB GORM DB definition
type DB struct {
	*Config      //内嵌一个指针
	Error        error
	RowsAffected int64
	Statement    *Statement //sql 相关数据状态
	clone        int        //只在 getInstance() 时可能发生拷贝一个 DB 的情况
}

// Session session config when create session with db.Session() method
// 这些字段 只在 db.Session() 函数里用到过
type Session struct {
	DryRun         bool //会话的配置
	PrepareStmt    bool //只在 db.Session 里面被调用过
	WithConditions bool // true 则db.clone = 2
	Context        context.Context
	Logger         logger.Interface
	NowFunc        func() time.Time //用于拷贝给 db.Config
}

// Open initialize db session based on dialector
func Open(dialector Dialector, config *Config) (db *DB, err error) {
	if config == nil {
		config = &Config{}
	}

	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{}
	}

	if config.Logger == nil {
		config.Logger = logger.Default
	}

	if config.NowFunc == nil {
		config.NowFunc = func() time.Time { return time.Now().Local() }
	}

	if dialector != nil {
		config.Dialector = dialector
	}

	if config.Plugins == nil {
		config.Plugins = map[string]Plugin{}
	}

	if config.cacheStore == nil {
		config.cacheStore = &sync.Map{}
	}

	db = &DB{Config: config, clone: 1}

	db.callbacks = initializeCallbacks(db)

	if config.ClauseBuilders == nil {
		config.ClauseBuilders = map[string]clause.ClauseBuilder{}
	}

	if config.Dialector != nil {
		// 数据库方言,  真正初始化处理函数链
		err = config.Dialector.Initialize(db)
	}

	if config.PrepareStmt {
		db.ConnPool = &PreparedStmtDB{
			ConnPool: db.ConnPool,
			Stmts:    map[string]*sql.Stmt{},
		}
	}

	db.Statement = &Statement{
		DB:       db,
		ConnPool: db.ConnPool,
		Context:  context.Background(),
		Clauses:  map[string]clause.Clause{},
	}

	if err == nil && !config.DisableAutomaticPing {
		if pinger, ok := db.ConnPool.(interface{ Ping() error }); ok {
			err = pinger.Ping()
		}
	}

	if err != nil {
		config.Logger.Error(context.Background(), "failed to initialize database, got error %v", err)
	}

	return
}

// Session create new db session
// session 也就是一些配置
func (db *DB) Session(config *Session) *DB {
	var (
		txConfig = *db.Config
		tx       = &DB{
			Config:    &txConfig,
			Statement: db.Statement,
			clone:     1,
		}
	)

	if config.Context != nil {
		tx.Statement = tx.Statement.clone()
		tx.Statement.DB = tx
		tx.Statement.Context = config.Context
	}

	if config.PrepareStmt {
		tx.Statement.ConnPool = &PreparedStmtDB{
			ConnPool: db.Config.ConnPool,
			Stmts:    map[string]*sql.Stmt{},
		}
	}

	if config.WithConditions {
		tx.clone = 2
	}

	if config.DryRun {
		tx.Config.DryRun = true
	}

	if config.Logger != nil {
		tx.Config.Logger = config.Logger
	}

	if config.NowFunc != nil {
		tx.Config.NowFunc = config.NowFunc
	}

	return tx
}

// WithContext change current instance db's context to ctx
func (db *DB) WithContext(ctx context.Context) *DB {
	return db.Session(&Session{WithConditions: true, Context: ctx})
}

// Debug start debug mode
func (db *DB) Debug() (tx *DB) {
	return db.Session(&Session{
		WithConditions: true,
		Logger:         db.Logger.LogMode(logger.Info),
	})
}

// Set -> store value with key into current db instance's context
func (db *DB) Set(key string, value interface{}) *DB {
	//可能就是同一个对象
	tx := db.getInstance()
	tx.Statement.Settings.Store(key, value)
	return tx
}

// Get get value with key from current db instance's context
func (db *DB) Get(key string) (interface{}, bool) {
	return db.Statement.Settings.Load(key)
}

// InstanceSet store value with key into current db instance's context
func (db *DB) InstanceSet(key string, value interface{}) *DB {
	tx := db.getInstance()
	tx.Statement.Settings.Store(fmt.Sprintf("%p", tx.Statement)+key, value)
	return tx
}

// InstanceGet get value with key from current db instance's context
func (db *DB) InstanceGet(key string) (interface{}, bool) {
	return db.Statement.Settings.Load(fmt.Sprintf("%p", db.Statement) + key)
}

// Callback returns callback manager
// 返回所有回调函数
func (db *DB) Callback() *callbacks {
	return db.callbacks
}

// AddError add error to db
// 加错误, 返回最新的错误
func (db *DB) AddError(err error) error {
	if db.Error == nil {
		db.Error = err
	} else if err != nil {
		//包装错误
		db.Error = fmt.Errorf("%v; %w", db.Error, err)
	}
	return db.Error
}

// DB returns `*sql.DB`
// 返回内部的连接池 一般是 *sql.DB 定义的原始通用接口
func (db *DB) DB() (*sql.DB, error) {
	connPool := db.ConnPool

	if stmtDB, ok := connPool.(*PreparedStmtDB); ok {
		connPool = stmtDB.ConnPool
	}

	if sqldb, ok := connPool.(*sql.DB); ok {
		return sqldb, nil
	}

	return nil, errors.New("invalid db")
}

//
func (db *DB) getInstance() *DB {
	if db.clone > 0 {
		tx := &DB{Config: db.Config}

		if db.clone == 1 {
			// clone with new statement
			tx.Statement = &Statement{
				DB:       tx,
				ConnPool: db.Statement.ConnPool,
				Context:  db.Statement.Context,
				Clauses:  map[string]clause.Clause{},
			}
		} else {
			// with clone statement
			tx.Statement = db.Statement.clone()
			tx.Statement.DB = tx
		}

		return tx
	}

	return db
}

func Expr(expr string, args ...interface{}) clause.Expr {
	return clause.Expr{SQL: expr, Vars: args}
}

// 表连接
func (db *DB) SetupJoinTable(model interface{}, field string, joinTable interface{}) error {
	var (
		tx                      = db.getInstance()
		stmt                    = tx.Statement
		modelSchema, joinSchema *schema.Schema
	)

	if err := stmt.Parse(model); err == nil {
		modelSchema = stmt.Schema
	} else {
		return err
	}

	if err := stmt.Parse(joinTable); err == nil {
		joinSchema = stmt.Schema
	} else {
		return err
	}

	if relation, ok := modelSchema.Relationships.Relations[field]; ok && relation.JoinTable != nil {
		for _, ref := range relation.References {
			if f := joinSchema.LookUpField(ref.ForeignKey.DBName); f != nil {
				f.DataType = ref.ForeignKey.DataType
				ref.ForeignKey = f
			} else {
				return fmt.Errorf("missing field %v for join table", ref.ForeignKey.DBName)
			}
		}

		for name, rel := range relation.JoinTable.Relationships.Relations {
			if _, ok := joinSchema.Relationships.Relations[name]; !ok {
				rel.Schema = joinSchema
				joinSchema.Relationships.Relations[name] = rel
			}
		}

		relation.JoinTable = joinSchema
	} else {
		return fmt.Errorf("failed to found relation: %v", field)
	}

	return nil
}

//添加插件
func (db *DB) Use(plugin Plugin) (err error) {
	name := plugin.Name()
	if _, ok := db.Plugins[name]; !ok {
		if err = plugin.Initialize(db); err == nil {
			db.Plugins[name] = plugin
		}
	} else {
		return ErrRegistered
	}

	return err
}
