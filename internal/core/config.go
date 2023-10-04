package core

import "os"

// SensordConf configuration of the SensorD
type SensordConf struct {
	SensorApiListenHttp string
	DatabaseUrl         string
	DatabaseLog         bool
	AdminApiListenHttp  string
}

func LoadConfig() *SensordConf {
	// create config from envs
	conf := &SensordConf{
		SensorApiListenHttp: os.Getenv("LISTEN_HTTP"),
		AdminApiListenHttp:  os.Getenv("ADMIN_LISTEN_HTTP"),
		DatabaseUrl:         os.Getenv("DB_URL"),
		DatabaseLog:         os.Getenv("DB_LOG") == "true",
	}
	return conf
}
