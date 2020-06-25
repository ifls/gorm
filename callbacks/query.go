package callbacks

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// query 处理函数
func Query(db *gorm.DB) {
	if db.Error == nil {
		if db.Statement.Schema != nil && !db.Statement.Unscoped {
			for _, c := range db.Statement.Schema.QueryClauses {
				db.Statement.AddClause(c)
			}
		}

		if db.Statement.SQL.String() == "" {
			BuildQuerySQL(db)
		}

		if !db.DryRun {
			//查询
			rows, err := db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if err != nil {
				db.AddError(err)
				return
			}
			defer rows.Close()
			//读取解析
			gorm.Scan(rows, db, false)
		}
	}
}

// 处理函数 构建 select 语句
func BuildQuerySQL(db *gorm.DB) {
	db.Statement.SQL.Grow(100)
	clauseSelect := clause.Select{Distinct: db.Statement.Distinct}

	if db.Statement.ReflectValue.Kind() == reflect.Struct {
		var conds []clause.Expression
		for _, primaryField := range db.Statement.Schema.PrimaryFields {
			if v, isZero := primaryField.ValueOf(db.Statement.ReflectValue); !isZero {
				conds = append(conds, clause.Eq{Column: clause.Column{Table: db.Statement.Table, Name: primaryField.DBName}, Value: v})
			}
		}

		if len(conds) > 0 {
			db.Statement.AddClause(clause.Where{Exprs: conds})
		}
	}

	if len(db.Statement.Selects) > 0 {
		clauseSelect.Columns = make([]clause.Column, len(db.Statement.Selects))
		for idx, name := range db.Statement.Selects {
			if db.Statement.Schema == nil {
				clauseSelect.Columns[idx] = clause.Column{Name: name, Raw: true}
			} else if f := db.Statement.Schema.LookUpField(name); f != nil {
				clauseSelect.Columns[idx] = clause.Column{Name: f.DBName}
			} else {
				clauseSelect.Columns[idx] = clause.Column{Name: name, Raw: true}
			}
		}
	}

	// inline joins
	if len(db.Statement.Joins) != 0 {
		if len(db.Statement.Selects) == 0 {
			clauseSelect.Columns = make([]clause.Column, len(db.Statement.Schema.DBNames))
			for idx, dbName := range db.Statement.Schema.DBNames {
				clauseSelect.Columns[idx] = clause.Column{Table: db.Statement.Table, Name: dbName}
			}
		}

		joins := []clause.Join{}
		for name, conds := range db.Statement.Joins {
			if db.Statement.Schema == nil {
				joins = append(joins, clause.Join{
					Expression: clause.Expr{SQL: name, Vars: conds},
				})
			} else if relation, ok := db.Statement.Schema.Relationships.Relations[name]; ok {
				tableAliasName := relation.Name

				for _, s := range relation.FieldSchema.DBNames {
					clauseSelect.Columns = append(clauseSelect.Columns, clause.Column{
						Table: tableAliasName,
						Name:  s,
						Alias: tableAliasName + "__" + s,
					})
				}

				exprs := make([]clause.Expression, len(relation.References))
				for idx, ref := range relation.References {
					if ref.OwnPrimaryKey {
						exprs[idx] = clause.Eq{
							Column: clause.Column{Table: db.Statement.Schema.Table, Name: ref.PrimaryKey.DBName},
							Value:  clause.Column{Table: tableAliasName, Name: ref.ForeignKey.DBName},
						}
					} else {
						if ref.PrimaryValue == "" {
							exprs[idx] = clause.Eq{
								Column: clause.Column{Table: db.Statement.Schema.Table, Name: ref.ForeignKey.DBName},
								Value:  clause.Column{Table: tableAliasName, Name: ref.PrimaryKey.DBName},
							}
						} else {
							exprs[idx] = clause.Eq{
								Column: clause.Column{Table: tableAliasName, Name: ref.ForeignKey.DBName},
								Value:  ref.PrimaryValue,
							}
						}
					}
				}

				joins = append(joins, clause.Join{
					Type:  clause.LeftJoin,
					Table: clause.Table{Name: relation.FieldSchema.Table, Alias: tableAliasName},
					ON:    clause.Where{Exprs: exprs},
				})
			} else {
				joins = append(joins, clause.Join{
					Expression: clause.Expr{SQL: name, Vars: conds},
				})
			}
		}

		db.Statement.AddClause(clause.From{Joins: joins})
	} else {
		db.Statement.AddClauseIfNotExists(clause.From{})
	}

	db.Statement.AddClauseIfNotExists(clauseSelect)

	db.Statement.Build("SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT", "FOR")
}

//处理函数
func Preload(db *gorm.DB) {
	if db.Error == nil && len(db.Statement.Preloads) > 0 {
		preloadMap := map[string][]string{}
		for name := range db.Statement.Preloads {
			if name == clause.Associations {
				for _, rel := range db.Statement.Schema.Relationships.Relations {
					if rel.Schema == db.Statement.Schema {
						preloadMap[rel.Name] = []string{rel.Name}
					}
				}
			} else {
				preloadFields := strings.Split(name, ".")
				for idx := range preloadFields {
					preloadMap[strings.Join(preloadFields[:idx+1], ".")] = preloadFields[:idx+1]
				}
			}
		}

		preloadNames := make([]string, len(preloadMap))
		idx := 0
		for key := range preloadMap {
			preloadNames[idx] = key
			idx++
		}
		sort.Strings(preloadNames)

		for _, name := range preloadNames {
			var (
				curSchema     = db.Statement.Schema
				preloadFields = preloadMap[name]
				rels          = make([]*schema.Relationship, len(preloadFields))
			)

			for idx, preloadField := range preloadFields {
				if rel := curSchema.Relationships.Relations[preloadField]; rel != nil {
					rels[idx] = rel
					curSchema = rel.FieldSchema
				} else {
					db.AddError(fmt.Errorf("%v: %w", name, gorm.ErrUnsupportedRelation))
				}
			}

			preload(db, rels, db.Statement.Preloads[name])
		}
	}
}
// 处理函数
func AfterQuery(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && db.Statement.Schema.AfterFind {
		callMethod(db, func(value interface{}, tx *gorm.DB) bool {
			if i, ok := value.(gorm.AfterFindInterface); ok {
				db.AddError(i.AfterFind(tx))
				return true
			}
			return false
		})
	}
}
