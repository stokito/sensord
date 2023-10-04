package db

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"sensord/internal/core"
	"sensord/internal/models"
	"sync"
	"testing"
	"time"
)

var utc, _ = time.LoadLocation("UTC")
var day1 = time.Date(2023, 1, 1, 0, 0, 0, 0, utc)
var day2 = time.Date(2023, 1, 2, 0, 0, 0, 0, utc)
var day3 = time.Date(2023, 1, 3, 0, 0, 0, 0, utc)
var day7 = time.Date(2023, 1, 7, 0, 0, 0, 0, utc)

// next week
var day8 = time.Date(2023, 1, 8, 0, 0, 0, 0, utc)

var storage SensorsDb

func TestMain(m *testing.M) {
	ctx := context.Background()
	conf := core.LoadConfig()

	storage = NewPostgresDb(conf.DatabaseUrl, conf.DatabaseLog)
	//dbConn = postgresDb
	dbErr := storage.Connect(ctx)
	if dbErr != nil {
		return
	}
	defer storage.Close()
	defer storage.Cleanup(ctx)

	os.Exit(m.Run())
}

func Test_StoreMeasurement(t *testing.T) {
	ctx := context.Background()
	storage.Cleanup(ctx)
	// check that for the day we don't have any measurements
	measurement, sqlErr := storage.GetMeasurementStatsForDay(ctx, day1, 1)
	assert.NoError(t, sqlErr)
	expected := &models.MeasurementRec{}
	assert.Equal(t, expected, measurement)

	// Insert the first record for a day
	storage.StoreMeasurement(ctx, day1, 1, 1.0)
	measurement, sqlErr = storage.GetMeasurementStatsForDay(ctx, day1, 1)
	assert.NoError(t, sqlErr)
	expected = &models.MeasurementRec{
		TotalCount: 1,
		TotalSum:   1,
		AvgValue:   1,
		MinValue:   1,
		MaxValue:   1,
	}
	assert.Equal(t, expected, measurement)
	storage.StoreMeasurement(ctx, day1, 1, 2.0)
	storage.StoreMeasurement(ctx, day1, 1, 3.0)
	// add a record for tomorrow
	storage.StoreMeasurement(ctx, day2, 1, 4.0)
	measurement, sqlErr = storage.GetMeasurementStatsForDay(ctx, day1, 1)
	assert.NoError(t, sqlErr)
	expected = &models.MeasurementRec{
		TotalCount: 3,
		TotalSum:   6,
		AvgValue:   2,
		MinValue:   1,
		MaxValue:   3,
	}
	assert.Equal(t, expected, measurement)
}

func Test_StoreMeasurement_Parallel(t *testing.T) {
	ctx := context.Background()
	storage.Cleanup(ctx)
	// Imitate crazy day with heavy load
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			storage.StoreMeasurement(ctx, day3, 1, 1.0)
		}()
	}

	wg.Wait()
	measurement, sqlErr := storage.GetMeasurementStatsForDay(ctx, day3, 1)
	assert.NoError(t, sqlErr)
	expected := &models.MeasurementRec{
		TotalCount: 100,
		TotalSum:   100,
		AvgValue:   1,
		MinValue:   1,
		MaxValue:   1,
	}
	assert.Equal(t, expected, measurement)
}

func Test_GetMeasurementPeriodStatsTotal(t *testing.T) {
	ctx := context.Background()
	storage.Cleanup(ctx)
	storage.StoreMeasurement(ctx, day1, 1, 1.0)
	storage.StoreMeasurement(ctx, day1, 2, 1.0)

	storage.StoreMeasurement(ctx, day2, 1, 1.0)
	storage.StoreMeasurement(ctx, day2, 2, 1.0)
	// next week
	storage.StoreMeasurement(ctx, day8, 1, 1.0)
	storage.StoreMeasurement(ctx, day8, 2, 1.0)

	stats, sqlErr := storage.GetMeasurementPeriodStatsTotal(ctx, day1, day7)
	assert.NoError(t, sqlErr)
	expected := &models.MeasurementRec{
		PeriodStart: day1,
		PeriodEnd:   day7,
		SensorId:    0,
		TotalCount:  4,
		TotalSum:    4,
		AvgValue:    1,
		MinValue:    1,
		MaxValue:    1,
	}
	assert.Equal(t, expected, stats)
}

func Test_GetMeasurementPeriodStatsForEachSensor(t *testing.T) {
	ctx := context.Background()
	storage.Cleanup(ctx)
	storage.StoreMeasurement(ctx, day1, 1, 1.0)
	storage.StoreMeasurement(ctx, day1, 2, 1.0)

	storage.StoreMeasurement(ctx, day2, 1, 1.0)
	storage.StoreMeasurement(ctx, day2, 2, 1.0)
	// next week
	storage.StoreMeasurement(ctx, day8, 1, 1.0)
	storage.StoreMeasurement(ctx, day8, 2, 1.0)

	stats, sqlErr := storage.GetMeasurementPeriodStatsForEachSensor(ctx, day1, day7)
	assert.NoError(t, sqlErr)
	expected := []*models.MeasurementRec{
		{
			PeriodStart: day1,
			PeriodEnd:   day7,
			SensorId:    1,
			TotalCount:  2,
			TotalSum:    2,
			AvgValue:    1,
			MinValue:    1,
			MaxValue:    1,
		},
		{
			PeriodStart: day1,
			PeriodEnd:   day7,
			SensorId:    2,
			TotalCount:  2,
			TotalSum:    2,
			AvgValue:    1,
			MinValue:    1,
			MaxValue:    1,
		},
	}

	assert.Equal(t, expected, stats)
}

func Test_GetMeasurementPeriodStatsForEachSensorAndDay(t *testing.T) {
	ctx := context.Background()
	storage.Cleanup(ctx)
	storage.StoreMeasurement(ctx, day1, 1, 1.0)
	storage.StoreMeasurement(ctx, day1, 2, 1.0)

	storage.StoreMeasurement(ctx, day2, 1, 1.0)
	storage.StoreMeasurement(ctx, day2, 2, 1.0)
	// next week
	storage.StoreMeasurement(ctx, day8, 1, 1.0)
	storage.StoreMeasurement(ctx, day8, 2, 1.0)

	stats, sqlErr := storage.GetMeasurementPeriodStatsForEachSensorAndDay(ctx, day1, day7)
	assert.NoError(t, sqlErr)
	expected := []*models.MeasurementRec{
		{
			PeriodStart: day1,
			PeriodEnd:   day2,
			SensorId:    1,
			TotalCount:  1,
			TotalSum:    1,
			AvgValue:    1,
			MinValue:    1,
			MaxValue:    1,
		},
		{
			PeriodStart: day2,
			PeriodEnd:   day3,
			SensorId:    1,
			TotalCount:  1,
			TotalSum:    1,
			AvgValue:    1,
			MinValue:    1,
			MaxValue:    1,
		},
		{
			PeriodStart: day1,
			PeriodEnd:   day2,
			SensorId:    2,
			TotalCount:  1,
			TotalSum:    1,
			AvgValue:    1,
			MinValue:    1,
			MaxValue:    1,
		},
		{
			PeriodStart: day2,
			PeriodEnd:   day3,
			SensorId:    2,
			TotalCount:  1,
			TotalSum:    1,
			AvgValue:    1,
			MinValue:    1,
			MaxValue:    1,
		},
	}

	assert.Equal(t, expected, stats)
}
