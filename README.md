# GORM Model Generator

[![Go Report Card](https://goreportcard.com/badge/github.com/ekobayong/gorm-model-generator)](https://goreportcard.com/report/github.com/ekobayong/gorm-model-generator)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A CLI tool to automatically generate [GORM](https://gorm.io/) models (structs) from an existing database schema.
Built with Go, intended to be a lightweight alternative to `sea-orm-cli` for the Go ecosystem.

## Features

- 🔍 **Schema Inspection**: Reads tables, columns, and data types from MySQL and PostgreSQL.
- 🔗 **Relationship Inference**: Automatically detects **Foreign Keys** and generates `BelongsTo` / `HasMany` relationships.
- 🏷️ **Smart Tagging**: Generates `gorm` and `json` tags with proper conventions (`camelCase` JSON, `snake_case` DB).
- 📦 **Import Management**: Auto-detects necessary imports (e.g., `time`, `gorm.io/datatypes`).
- 🛠️ **CLI Interface**: Easy to use command-line interface powered by Cobra.

## Installation

### From Source

```bash
go install github.com/ekobayong/gorm-model-generator@latest
```

Or clone and build locally:

```bash
git clone https://github.com/ekobayong/gorm-model-generator.git
cd gorm-model-generator
go build -o gorm-model-generator
```

## Usage

Basic usage to generate models from a database:

```bash
gorm-model-generator generate -u "user:pass@tcp(localhost:3306)/dbname"
```

### Supported Flags

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--database-url` | `-u` | Database Connection String (DSN) **(Required)** | - |
| `--driver` | `-d` | Database driver (`mysql`, `postgres`) | `mysql` |
| `--output-dir` | `-o` | Directory for generated files | `./models` |
| `--package` | `-p` | Go package name for the files | `model` |
| `--tables` | `-t` | Comma-separated list of tables to include | (All) |
| `--ignore-tables`| `-i` | Comma-separated list of tables to exclude | (None) |

### Examples

**MySQL:**
```bash
./gorm-model-generator generate \
  -u "root:secret@tcp(127.0.0.1:3306)/my_app?charset=utf8mb4&parseTime=True&loc=Local" \
  -o ./internal/entity \
  -p entity
```

**PostgreSQL:**
```bash
./gorm-model-generator generate \
  -d postgres \
  -u "host=localhost user=postgres password=secret dbname=my_app port=5432 sslmode=disable" \
  -o ./db/models
```

## How It Works

1.  **Connects** to your database using the provided DSN.
2.  **Inspects** `information_schema` to retrieve table structures and constraints.
3.  **Analyzes** Foreign Keys to build a dependency graph.
4.  **Generates** Go code using `text/template`, mapping SQL types to Go types (e.g., `VARCHAR` -> `string`, `TIMESTAMP` -> `time.Time`).

## Roadmap

- [ ] SQLite Support
- [ ] Many-to-Many relationship inference
- [ ] Custom naming strategies
- [ ] Interactive mode (UI)

## Contributing

Contributions are welcome! Please open an issue or submit a Pull Request.

## License

MIT License. See [LICENSE](LICENSE) file.
