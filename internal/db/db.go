package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"sensord/internal/config"

	"log"
	"time"
)

var db *pgxpool.Pool

// language=PostgreSQL
var sqlMeasureInsert = `INSERT INTO measurements (
	measure_time, sensor_id, value) 
	VALUES ($1, $2, $3)
	ON CONFLICT (measure_time, sensor_id) DO NOTHING;
`

// language=PostgreSQL
var sqlCreateSensor = `INSERT INTO sensors (
	id, name, room) 
	VALUES ($1, $2, $3);
`

// DbLog logger for SQL queries
type DbLog struct {
}

func (l *DbLog) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	queryArgs := data["args"]
	querySql := data["sql"]
	log.Printf("INFO: SQL %s args: %v %s\n", msg, queryArgs, querySql)
}

func DbConnect(ctx context.Context) error {
	poolConfig, err := pgxpool.ParseConfig(config.Conf.DatabaseUrl)
	if err != nil {
		return errors.Wrap(err, "Unable to parse database URL")
	}
	if config.Conf.DatabaseLog {
		poolConfig.ConnConfig.Logger = &DbLog{}
		poolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace
	}
	var dbErr error
	db, dbErr = pgxpool.ConnectConfig(ctx, poolConfig)
	return dbErr
}

func DbClose() {
	if db == nil {
		return
	}
	db.Close()
	db = nil
	log.Printf("INFO: DB disconnected\n")
}

func StoreMeasureToDb(ctx context.Context, measureTime time.Time, sensorId int, value float64) {
	_, sqlErr := db.Exec(ctx, sqlMeasureInsert,
		measureTime, sensorId, value)
	if sqlErr != nil {
		log.Printf("WARN: Fail to insert measure %v\n", sqlErr)
	}
}

func CreateSensor(ctx context.Context, sensorId int, name, room string) {
	_, sqlErr := db.Exec(ctx, sqlCreateSensor,
		sensorId, name, room)
	if sqlErr != nil {
		log.Printf("WARN: Fail to insert new sensor %v\n", sqlErr)
	}
}
