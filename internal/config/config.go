package config

// SensordConf configuration of the SensorD
type SensordConf struct {
	ApiListenHttp string
	DatabaseUrl   string
	DatabaseLog   bool
}

var Conf *SensordConf
