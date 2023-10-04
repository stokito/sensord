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

	dbConn := NewPostgresDb(config.Conf.DatabaseUrl, config.Conf.DatabaseLog)
	dbErr := dbConn.Connect(ctx)
	if !assert.NoError(t, dbErr) {
		return
	}
	defer dbConn.Close()

	measureTime := time.Now()
	sensorId := 1
	measureValue := 42.0
	dbConn.StoreMeasurement(ctx, measureTime, sensorId, measureValue)
	dbConn.Cleanup(ctx)
}
