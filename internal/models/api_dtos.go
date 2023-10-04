package models

import "time"

type MeasurementDto struct {
	SensorId int       `json:"sensorId"`
	Time     time.Time `json:"time"`
	Value    float64   `json:"value"`
}
