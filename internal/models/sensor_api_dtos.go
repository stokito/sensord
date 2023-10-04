package models

import "time"

// MeasurementDto used to parse a request from a sensor
type MeasurementDto struct {
	SensorId int       `json:"sensorId"`
	Time     time.Time `json:"time"`
	Value    float64   `json:"value"`
}
