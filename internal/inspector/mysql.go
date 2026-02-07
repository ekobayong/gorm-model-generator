package inspector

import (
	"fmt"

	"github.com/ekobayong/gorm-model-generator/internal/model"

	"gorm.io/gorm"
)

type MySQLInspector struct {
	db *gorm.DB
}

func NewMySQLInspector(db *gorm.DB) *MySQLInspector {
	return &MySQLInspector{db: db}
}

func (i *MySQLInspector) GetTables(schema string, whitelist, blacklist []string) ([]model.Table, error) {
	var tableNames []string
	
    // Default to current database if schema is empty
    query := "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()"
    if schema != "" {
        query = fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s'", schema)
    }

	if err := i.db.Raw(query).Scan(&tableNames).Error; err != nil {
		return nil, err
	}

	var tables []model.Table
	for _, name := range tableNames {
		// Filter whitelist
		if len(whitelist) > 0 {
			found := false
			for _, w := range whitelist {
				if w == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter blacklist
		ignored := false
		for _, b := range blacklist {
			if b == name {
				ignored = true
				break
			}
		}
		if ignored {
			continue
		}

		tables = append(tables, model.Table{Name: name})
	}
	return tables, nil
}

func (i *MySQLInspector) GetColumns(tableName string) ([]model.Column, error) {
	var columns []model.Column
	
    // We use information_schema for more details
	rows, err := i.db.Raw(`
		SELECT column_name, data_type, is_nullable, column_key, column_comment
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() AND table_name = ?
        ORDER BY ordinal_position
	`, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var colName, dataType, isNullable, colKey, comment string
		if err := rows.Scan(&colName, &dataType, &isNullable, &colKey, &comment); err != nil {
			return nil, err
		}

		col := model.Column{
			Name:         colName,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			IsPrimaryKey: colKey == "PRI",
			Comment:      comment,
		}
        
        // Map to Go types (simple mapping for now)
        col.GoType = i.mapType(dataType)
        col.GormTag = i.buildGormTag(col)
        col.JsonTag = colName
        
		columns = append(columns, col)
	}
	return columns, nil
}

func (i *MySQLInspector) GetForeignKeys(tableName string) ([]model.ForeignKey, error) {
    var fks []model.ForeignKey

    rows, err := i.db.Raw(`
        SELECT 
            k.constraint_name, 
            k.column_name, 
            k.referenced_table_name, 
            k.referenced_column_name,
            r.update_rule,
            r.delete_rule
        FROM information_schema.key_column_usage k
        JOIN information_schema.referential_constraints r 
            ON k.constraint_name = r.constraint_name 
            AND k.table_schema = r.constraint_schema
        WHERE k.referenced_table_name IS NOT NULL 
          AND k.table_schema = DATABASE() 
          AND k.table_name = ?
    `, tableName).Rows()
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var fk model.ForeignKey
        if err := rows.Scan(&fk.ConstraintName, &fk.ColumnName, &fk.RefTableName, &fk.RefColumnName, &fk.OnUpdate, &fk.OnDelete); err != nil {
            return nil, err
        }
        fks = append(fks, fk)
    }

    return fks, nil
}

func (i *MySQLInspector) mapType(dataType string) string {
    switch dataType {
    case "int", "integer", "smallint", "tinyint", "mediumint":
        return "int64"
    case "bigint":
        return "int64"
    case "float", "double", "decimal":
        return "float64"
    case "char", "varchar", "text", "longtext", "mediumtext", "tinytext":
        return "string"
    case "date", "datetime", "timestamp":
        return "time.Time"
    case "blob", "longblob", "mediumblob", "tinyblob", "binary", "varbinary":
        return "[]byte"
    case "tinyint(1)", "boolean", "bool": // Handle MySQL boolean
         return "bool"
    case "json":
        return "datatypes.JSON" // Requires datatypes package
    default:
        return "string"
    }
}

func (i *MySQLInspector) buildGormTag(col model.Column) string {
    tag := fmt.Sprintf("column:%s", col.Name)
    if col.IsPrimaryKey {
        tag += ";primaryKey"
    }
    if !col.IsNullable {
        tag += ";not null"
    }
    // AutoIncrement logic usually hidden in type, but for INT PK it's default in GORM
    // We can add type:%s if stricter control needed
    tag += fmt.Sprintf(";type:%s", col.DataType)
    return tag
}
