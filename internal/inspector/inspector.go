package inspector

import (
	"github.com/ekobayong/gorm-model-generator/internal/model"
)

// Inspector defines the interface for database inspection
type Inspector interface {
	GetTables(schema string, whitelist, blacklist []string) ([]model.Table, error)
	GetColumns(tableName string) ([]model.Column, error)
	GetForeignKeys(tableName string) ([]model.ForeignKey, error)
}
