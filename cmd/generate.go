package cmd

import (
	"fmt"
	"log"

	"github.com/ekobayong/gorm-model-generator/internal/config"
	"github.com/ekobayong/gorm-model-generator/internal/generator"
	"github.com/ekobayong/gorm-model-generator/internal/inspector"

	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate GORM models",
	Long:  `Connects to the database and generates Go structs based on the schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		runGenerate()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&config.GlobalConfig.DatabaseURL, "database-url", "u", "", "Database URL (DSN)")
	generateCmd.Flags().StringVarP(&config.GlobalConfig.OutputDirectory, "output-dir", "o", "./models", "Output directory for generated models")
	generateCmd.Flags().StringVarP(&config.GlobalConfig.PackageName, "package", "p", "model", "Package name for generated models")
	generateCmd.Flags().StringSliceVarP(&config.GlobalConfig.Tables, "tables", "t", []string{}, "List of tables to generate (comma separated)")
	generateCmd.Flags().StringSliceVarP(&config.GlobalConfig.IgnoreTables, "ignore-tables", "i", []string{}, "List of tables to ignore (comma separated)")
	generateCmd.Flags().StringVarP(&config.GlobalConfig.Driver, "driver", "d", "mysql", "Database driver (mysql, postgres, sqlite)")

	generateCmd.MarkFlagRequired("database-url")
}

func runGenerate() {
	var db *gorm.DB
	var err error
	var insp inspector.Inspector

	switch config.GlobalConfig.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(config.GlobalConfig.DatabaseURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		insp = inspector.NewMySQLInspector(db)
	case "postgres":
		db, err = gorm.Open(postgres.Open(config.GlobalConfig.DatabaseURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		insp = inspector.NewPostgresInspector(db) // TO BE IMPLEMENTED
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.GlobalConfig.DatabaseURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		// insp = inspector.NewSQLiteInspector(db) // TODO
		log.Fatal("SQLite inspector not yet implemented")
	default:
		log.Fatalf("unsupported driver: %s", config.GlobalConfig.Driver)
	}

	fmt.Printf("Connected to %s database. Inspecting schema...\n", config.GlobalConfig.Driver)

	// 1. Get Tables
	tables, err := insp.GetTables("", config.GlobalConfig.Tables, config.GlobalConfig.IgnoreTables)
	if err != nil {
		log.Fatalf("failed to get tables: %v", err)
	}

	for i, table := range tables {
		fmt.Printf("Processing table: %s\n", table.Name)

		// 2. Get Columns
		cols, err := insp.GetColumns(table.Name)
		if err != nil {
			log.Fatalf("failed to get columns for table %s: %v", table.Name, err)
		}
		tables[i].Columns = cols

		// 3. Get Foreign Keys
		fks, err := insp.GetForeignKeys(table.Name)
		if err != nil {
			log.Printf("warning: failed to get foreign keys for table %s: %v", table.Name, err)
		}
		tables[i].ForeignKeys = fks
	}

	// 4. Generate Code
	gen := generator.NewGenerator(config.GlobalConfig)
	if err := gen.Generate(tables); err != nil {
		log.Fatalf("failed to generate code: %v", err)
	}

	fmt.Println("Done.")
}
