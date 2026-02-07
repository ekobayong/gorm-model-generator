package generator

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ekobayong/gorm-model-generator/internal/config"
	"github.com/ekobayong/gorm-model-generator/internal/model"

	"github.com/iancoleman/strcase"
)

type Generator struct {
	Config config.Config
}

func NewGenerator(cfg config.Config) *Generator {
	return &Generator{Config: cfg}
}

func (g *Generator) Generate(tables []model.Table) error {
	// 1. Preprocess: Enrich model with Go types and relationships
	enrichedTables := g.processTables(tables)

	// 2. Prepare output directory
	if err := os.MkdirAll(g.Config.OutputDirectory, 0755); err != nil {
		return err
	}

	// 3. Generate files
	tmpl, err := template.New("model").Parse(modelTemplate)
	if err != nil {
		return err
	}

	for _, table := range enrichedTables {
		filename := filepath.Join(g.Config.OutputDirectory, table.Name+".go")
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		data := struct {
			PackageName string
			Table       model.Table
			Imports     []string
		}{
			PackageName: g.Config.PackageName,
			Table:       table,
			Imports:     detectImports(table),
		}

		if err := tmpl.Execute(f, data); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) processTables(tables []model.Table) []model.Table {
	tableMap := make(map[string]*model.Table)
	for i := range tables {
		tableMap[tables[i].Name] = &tables[i]
	}

	for i := range tables {
		t := &tables[i]
		t.GoName = strcase.ToCamel(t.Name)

		// Linking Foreign Keys
		for _, fk := range t.ForeignKeys {
			refTable, ok := tableMap[fk.RefTableName]
			if !ok {
				continue // Referenced table not in generation list
			}

			// Add "Belongs To" relation to current table
			// e.g. User belongs to Company
			relationName := strcase.ToCamel(fk.RefTableName)
            // Trim "s" if plural? naive approach for now
            if strings.HasSuffix(relationName, "s") {
                relationName = strings.TrimSuffix(relationName, "s")
            }
            
			t.Relations = append(t.Relations, model.Relation{
				Name:         relationName,
				Type:         "BelongsTo",
				OtherType:    strcase.ToCamel(refTable.Name),
				ForeignKey:   strcase.ToCamel(fk.ColumnName),
				References:   strcase.ToCamel(fk.RefColumnName),
			})

			// Add "Has Many" relation to referenced table
			// e.g. Company has many Users
            hasManyName := strcase.ToCamel(t.Name)
            // Ensure plural for HasMany
             if !strings.HasSuffix(hasManyName, "s") {
                hasManyName += "s"
            }

			refTable.Relations = append(refTable.Relations, model.Relation{
				Name:       hasManyName,
				Type:       "HasMany",
				OtherType:  t.GoName,
				ForeignKey: strcase.ToCamel(fk.ColumnName),
				References: strcase.ToCamel(fk.RefColumnName),
			})
		}
        
        // Fix up columns Go definitions
        for j := range t.Columns {
            col := &t.Columns[j]
            col.GoName = strcase.ToCamel(col.Name)
            if col.Name == "id" {
                col.GoName = "ID"
            }
            // ID fix for FKs e.g. company_id -> CompanyID
             if strings.HasSuffix(col.GoName, "Id") {
                col.GoName = strings.TrimSuffix(col.GoName, "Id") + "ID"
            }
        }
	}
    
    return tables
}

func detectImports(t model.Table) []string {
    imports := make(map[string]bool)
    for _, col := range t.Columns {
        if col.GoType == "time.Time" {
            imports["time"] = true
        }
        if strings.Contains(col.GoType, "datatypes") {
            imports["gorm.io/datatypes"] = true
        }
    }
    var result []string
    for k := range imports {
        result = append(result, k)
    }
    return result
}

const modelTemplate = `package {{.PackageName}}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)

// {{.Table.GoName}} mapped from table <{{.Table.Name}}>
type {{.Table.GoName}} struct {
{{- range .Table.Columns}}
	{{.GoName}} {{.GoType}} ` + "`gorm:\"{{.GormTag}}\" json:\"{{.JsonTag}}\"`" + `
{{- end}}
{{- if .Table.Relations}}

    // Relationships
{{- range .Table.Relations}}
    {{.Name}} {{if eq .Type "HasMany"}}[]{{end}}{{.OtherType}} ` + "`gorm:\"foreignKey:{{.ForeignKey}};references:{{.References}}\" json:\"{{.Name}},omitempty\"`" + `
{{- end}}
{{- end}}
}

func ({{.Table.GoName}}) TableName() string {
	return "{{.Table.Name}}"
}
`
