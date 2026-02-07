package config

// Config holds the configuration for the generator
type Config struct {
	DatabaseURL     string
	OutputDirectory string
	PackageName     string
	Tables          []string
	IgnoreTables    []string
	Driver          string // mysql, postgres, sqlite
}

// GlobalConfig stores the parsed configuration
var GlobalConfig Config
