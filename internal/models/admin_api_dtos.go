package models

import "time"

// MeasurementRec used in stats reporting
type MeasurementRec struct {
	// PeriodStart Day at midnight
	PeriodStart time.Time
	// PeriodEnd when the stats period ends. Exclusive
	PeriodEnd  time.Time
	SensorId   int
	TotalCount int64
	TotalSum   float64
	// Average temperature
	AvgValue float64
	// Minimal temperature
	MinValue float64
	// Maximum temperature
	MaxValue float64
}
