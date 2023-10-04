package db

import "time"

type MeasurementRec struct {
	MeasureTime time.Time `db:"measure_time"`
	SensorId    int       `db:"sensor_id"`
	Value       float64   `db:"value"`
}
