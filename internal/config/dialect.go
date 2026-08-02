package config

import (
	"fmt"

	"gorm.io/gorm"
)

// IsPostgres reports whether db uses PostgreSQL dialect.
func IsPostgres(db *gorm.DB) bool {
	if db == nil {
		db = GetDB()
	}
	return db != nil && db.Dialector.Name() == "postgres"
}

// QuoteIdent 按方言引用标识符（列名/表名）：postgres 用双引号，其余用反引号。
func QuoteIdent(db *gorm.DB, name string) string {
	if IsPostgres(db) {
		return `"` + name + `"`
	}
	return "`" + name + "`"
}

// JSONIntExpr builds a SQL expression to extract an integer from a JSON column.
func JSONIntExpr(db *gorm.DB, column, key string) string {
	if IsPostgres(db) {
		return fmt.Sprintf("(%s::json->>'%s')::bigint", column, key)
	}
	return fmt.Sprintf("CAST(JSON_EXTRACT(%s, '$.%s') AS INTEGER)", column, key)
}
