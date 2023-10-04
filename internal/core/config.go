package core

import "os"

// SensordConf configuration of the SensorD
type SensordConf struct {
	// Sensor HTTP API listen address. You can specify `hostname:port` or just `:port`
	// Env: SENSOR_LISTEN_HTTP
	SensorApiListenHttp string

	// Admin HTTP API listen address
	// Env: ADMIN_LISTEN_HTTP
	AdminApiListenHttp string

	// DatabaseUrl PostgreSQL connection string.
	// postgres://postgres:postgres@localhost:5432/sensorsdb?sslmode=disable&search_path=sensors
	// Env: DB_LOG
	DatabaseUrl string

	// DatabaseLog if `true` then log SQL queries and args. Useful for testing and debug.
	// Env: DB_LOG
	DatabaseLog bool
}

// LoadConfig from environment variables
func LoadConfig() *SensordConf {
	// create config from envs
	conf := &SensordConf{
		SensorApiListenHttp: os.Getenv("SENSOR_LISTEN_HTTP"),
		AdminApiListenHttp:  os.Getenv("ADMIN_LISTEN_HTTP"),
		DatabaseUrl:         os.Getenv("DB_URL"),
		DatabaseLog:         os.Getenv("DB_LOG") == "true",
	}
	return conf
}
