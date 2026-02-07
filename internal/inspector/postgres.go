package inspector

import (
	"fmt"

	"github.com/ekobayong/gorm-model-generator/internal/model"

	"gorm.io/gorm"
)

type PostgresInspector struct {
	db *gorm.DB
}

func NewPostgresInspector(db *gorm.DB) *PostgresInspector {
	return &PostgresInspector{db: db}
}

func (i *PostgresInspector) GetTables(schema string, whitelist, blacklist []string) ([]model.Table, error) {
	var tableNames []string
    
    if schema == "" {
        schema = "public"
    }

    query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s'", schema)
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

func (i *PostgresInspector) GetColumns(tableName string) ([]model.Column, error) {
	var columns []model.Column

	rows, err := i.db.Raw(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = ?
        ORDER BY ordinal_position
	`, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

    // Need to check for PKs separately or join
    // For simplicity, let's just get columns first then PKs? 
    // Or just query constraint info.
    
    // Let's get PKs first
    pks := make(map[string]bool)
    pkRows, err := i.db.Raw(`
        SELECT kcu.column_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu
          ON tc.constraint_name = kcu.constraint_name
          AND tc.table_schema = kcu.table_schema
        WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = ?
    `, tableName).Rows()
    if err == nil {
        defer pkRows.Close()
        for pkRows.Next() {
            var pkCol string
            pkRows.Scan(&pkCol)
            pks[pkCol] = true
        }
    }

	for rows.Next() {
		var colName, dataType, isNullable, colDefault string
		if err := rows.Scan(&colName, &dataType, &isNullable, &colDefault); err != nil {
			return nil, err
		}

		col := model.Column{
			Name:         colName,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			IsPrimaryKey: pks[colName],
		}
        
        col.GoType = i.mapType(dataType)
        col.GormTag = i.buildGormTag(col)
        col.JsonTag = colName
        
		columns = append(columns, col)
	}
	return columns, nil
}

func (i *PostgresInspector) GetForeignKeys(tableName string) ([]model.ForeignKey, error) {
    var fks []model.ForeignKey

    rows, err := i.db.Raw(`
        SELECT
            tc.constraint_name, 
            kcu.column_name, 
            ccu.table_name AS foreign_table_name,
            ccu.column_name AS foreign_column_name,
            rc.update_rule,
            rc.delete_rule
        FROM 
            information_schema.table_constraints AS tc 
            JOIN information_schema.key_column_usage AS kcu
              ON tc.constraint_name = kcu.constraint_name
              AND tc.table_schema = kcu.table_schema
            JOIN information_schema.constraint_column_usage AS ccu
              ON ccu.constraint_name = tc.constraint_name
              AND ccu.table_schema = tc.table_schema
            JOIN information_schema.referential_constraints AS rc
                ON rc.constraint_name = tc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name = ?
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

func (i *PostgresInspector) mapType(dataType string) string {
    switch dataType {
    case "smallint", "integer", "bigint":
        return "int64"
    case "boolean":
         return "bool"
    case "real", "double precision", "numeric":
        return "float64"
    case "character varying", "text", "character", "uuid":
        return "string"
    case "date", "timestamp without time zone", "timestamp with time zone":
        return "time.Time"
    case "bytea":
        return "[]byte"
    case "json", "jsonb":
        return "datatypes.JSON"
    default:
        return "string"
    }
}

func (i *PostgresInspector) buildGormTag(col model.Column) string {
    tag := fmt.Sprintf("column:%s", col.Name)
    if col.IsPrimaryKey {
        tag += ";primaryKey"
    }
    if !col.IsNullable {
        tag += ";not null"
    }
    tag += fmt.Sprintf(";type:%s", col.DataType)
    return tag
}
