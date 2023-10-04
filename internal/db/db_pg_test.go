package db

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sensord/internal/config"
	"sync"
	"testing"
	"time"
)

// The integration test
func Test_Db(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig()

	dbConn := NewPostgresDb(config.Conf.DatabaseUrl, config.Conf.DatabaseLog)
	dbErr := dbConn.Connect(ctx)
	if !assert.NoError(t, dbErr) {
		return
	}
	defer dbConn.Close()

	utc, _ := time.LoadLocation("UTC")
	day := time.Date(2023, 1, 1, 0, 0, 0, 0, utc)
	sensorId := 1

	// check that for the day we don't have any measurements
	measurement, sqlErr := dbConn.GetMeasurementStatsForDay(ctx, day, sensorId)
	assert.NoError(t, sqlErr)
	expected := &MeasurementRec{
		Day:      day,
		SensorId: sensorId,
	}
	assert.Equal(t, expected, measurement)

	// Insert the first record for a day
	dbConn.StoreMeasurement(ctx, day, sensorId, 1.0)
	measurement, sqlErr = dbConn.GetMeasurementStatsForDay(ctx, day, sensorId)
	assert.NoError(t, sqlErr)
	expected = &MeasurementRec{
		Day:        day,
		SensorId:   1,
		TotalCount: 1,
		TotalSum:   1,
		AvgValue:   1,
		MinValue:   1,
		MaxValue:   1,
	}
	assert.Equal(t, expected, measurement)
	dbConn.StoreMeasurement(ctx, day, sensorId, 2.0)
	dbConn.StoreMeasurement(ctx, day, sensorId, 3.0)
	// add a record for tomorrow
	nextDay := day.AddDate(0, 0, 1)
	dbConn.StoreMeasurement(ctx, nextDay, sensorId, 4.0)
	measurement, sqlErr = dbConn.GetMeasurementStatsForDay(ctx, day, sensorId)
	assert.NoError(t, sqlErr)
	expected = &MeasurementRec{
		Day:        day,
		SensorId:   1,
		TotalCount: 3,
		TotalSum:   6,
		AvgValue:   2,
		MinValue:   1,
		MaxValue:   3,
	}
	assert.Equal(t, expected, measurement)

	// Imitate crazy day with heavy load
	crazyDay := nextDay.AddDate(0, 1, 0)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbConn.StoreMeasurement(ctx, crazyDay, sensorId, 1.0)
		}()
	}

	wg.Wait()
	measurement, sqlErr = dbConn.GetMeasurementStatsForDay(ctx, crazyDay, sensorId)
	assert.NoError(t, sqlErr)
	expected = &MeasurementRec{
		Day:        crazyDay,
		SensorId:   1,
		TotalCount: 10000,
		TotalSum:   10000,
		AvgValue:   1,
		MinValue:   1,
		MaxValue:   1,
	}
	assert.Equal(t, expected, measurement)

	dbConn.Cleanup(ctx)
}
