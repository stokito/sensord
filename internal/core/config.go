package core

import "os"

// SensordConf configuration of the SensorD
type SensordConf struct {
	ApiListenHttp string
	DatabaseUrl   string
	DatabaseLog   bool
}

func LoadConfig() *SensordConf {
	// create config from envs
	conf := &SensordConf{
		ApiListenHttp: os.Getenv("LISTEN_HTTP"),
		DatabaseUrl:   os.Getenv("DB_URL"),
		DatabaseLog:   os.Getenv("DB_LOG") == "true",
	}
	return conf
}
