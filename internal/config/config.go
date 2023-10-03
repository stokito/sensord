package config

import "os"

// SensordConf configuration of the SensorD
type SensordConf struct {
	ApiListenHttp string
	DatabaseUrl   string
	DatabaseLog   bool
}

var Conf *SensordConf

func LoadConfig() {
	// create config from envs
	conf := &SensordConf{
		ApiListenHttp: os.Getenv("LISTEN_HTTP"),
		DatabaseUrl:   os.Getenv("DB_URL"),
		DatabaseLog:   os.Getenv("DB_LOG") == "true",
	}
	Conf = conf
}
