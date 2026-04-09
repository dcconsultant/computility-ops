package config

import "os"

type Config struct {
	Addr          string
	StorageDriver string
	MySQLDSN      string
}

func Load() Config {
	return Config{
		Addr:          getenv("APP_ADDR", ":8080"),
		StorageDriver: getenv("STORAGE_DRIVER", "memory"),
		MySQLDSN:      os.Getenv("MYSQL_DSN"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
