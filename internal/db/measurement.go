package db

import "time"

type MeasurementRec struct {
	Day        time.Time `db:"measurement_day"`
	SensorId   int       `db:"sensor_id"`
	TotalCount int64     `db:"total_count"`
	TotalSum   float64   `db:"total_sum"`
	AvgValue   float64   `db:"avg_value"`
	MinValue   float64   `db:"min_value"`
	MaxValue   float64   `db:"max_value"`
}
