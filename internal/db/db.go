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
	sensor_id, measure_time, value) 
	VALUES ($1, $2, $3)
	ON CONFLICT (sensor_id, measure_time) DO NOTHING;
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

func StoreMeasureToDb(ctx context.Context, sensorId, measureTime time.Time, value float64) {
	_, sqlErr := db.Exec(ctx, sqlMeasureInsert,
		sensorId, measureTime, value)
	if sqlErr != nil {
		log.Printf("WARN: Fail to insert measure %v\n", sqlErr)
	}
}
