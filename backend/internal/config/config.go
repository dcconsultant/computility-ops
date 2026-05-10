package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr                 string
	StorageDriver        string
	MySQLDSN             string
	MetaImportCleanDays  int
	MetaImportKeepLatest int
}

func Load() Config {
	return Config{
		Addr:                 getenv("APP_ADDR", ":8080"),
		StorageDriver:        getenv("STORAGE_DRIVER", "memory"),
		MySQLDSN:             os.Getenv("MYSQL_DSN"),
		MetaImportCleanDays:  getenvInt("META_IMPORT_CLEAN_DAYS", 7),
		MetaImportKeepLatest: getenvInt("META_IMPORT_KEEP_LATEST", 200),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
