package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"log"
	"time"
)

// language=PostgreSQL
var sqlMeasurementInsert = `
INSERT INTO measurement (
	measure_time, sensor_id, value) 
VALUES ($1, $2, $3)
ON CONFLICT (measure_time, sensor_id) DO NOTHING;
`

// language=PostgreSQL
var sqlCreateSensor = `
INSERT INTO sensors (
	id, name, room) 
VALUES ($1, $2, $3);
`

// language=PostgreSQL
var sqlCleanup = `TRUNCATE sensors RESTART IDENTITY CASCADE;
`

// DbLog logger for SQL queries
type DbLog struct {
}

func (l *DbLog) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	queryArgs := data["args"]
	querySql := data["sql"]
	log.Printf("DEBUG: SQL %s args: %v %s\n", msg, queryArgs, querySql)
}

type PostgresDb struct {
	pool        *pgxpool.Pool
	DatabaseUrl string
	DatabaseLog bool
}

func NewPostgresDb(databaseUrl string, databaseLog bool) *PostgresDb {
	return &PostgresDb{
		DatabaseUrl: databaseUrl,
		DatabaseLog: databaseLog,
	}
}

func (db *PostgresDb) Connect(ctx context.Context) error {
	poolConfig, err := pgxpool.ParseConfig(db.DatabaseUrl)
	if err != nil {
		return errors.Wrap(err, "Unable to parse database URL.")
	}
	if db.DatabaseLog {
		poolConfig.ConnConfig.Logger = &DbLog{}
		poolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace
	}
	pool, dbErr := pgxpool.ConnectConfig(ctx, poolConfig)
	if dbErr != nil {
		return dbErr
	}
	db.pool = pool
	log.Printf("INFO: Connected to database\n")
	return nil
}

func (db *PostgresDb) Close() {
	if db.pool == nil {
		return
	}
	db.pool.Close()
	db.pool = nil
	log.Printf("INFO: DB disconnected\n")
}

func (db *PostgresDb) CreateSensor(ctx context.Context, sensorId int, name, room string) {
	_, sqlErr := db.pool.Exec(ctx, sqlCreateSensor,
		sensorId, name, room)
	if sqlErr != nil {
		log.Printf("WARN: Fail to insert new sensor %v\n", sqlErr)
	}
}

func (db *PostgresDb) StoreMeasurement(ctx context.Context, measureTime time.Time, sensorId int, value float64) {
	_, sqlErr := db.pool.Exec(ctx, sqlMeasurementInsert,
		measureTime, sensorId, value)
	if sqlErr != nil {
		log.Printf("WARN: Fail to insert measure %v\n", sqlErr)
	}
}

// Cleanup DB e.g. remove sensors and all their measurements.
// Used for testing
func (db *PostgresDb) Cleanup(ctx context.Context) {
	_, sqlErr := db.pool.Exec(ctx, sqlCleanup)
	if sqlErr != nil {
		log.Printf("WARN: Fail to cleanup %v\n", sqlErr)
	}
}
