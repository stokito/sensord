package db

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sensord/internal/config"
	"testing"
	"time"
)

// The integration test
func Test_Db(t *testing.T) {
	ctx := context.Background()
	config.LoadConfig()
	dbErr := DbConnect(ctx)
	if !assert.NoError(t, dbErr) {
		return
	}
	defer DbClose()

	CreateSensor(ctx, 1, "Sensor1", "Room1")
	measureTime := time.Now()
	sensorId := 1
	measureValue := 42.0
	StoreMeasureToDb(ctx, measureTime, sensorId, measureValue)
}
