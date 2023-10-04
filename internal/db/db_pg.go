package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"log"
	"time"
)

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

// Cleanup DB e.g. remove sensors and all their measurements.
// Useful for testing
func (db *PostgresDb) Cleanup(ctx context.Context) {
	_, sqlErr := db.pool.Exec(ctx, `TRUNCATE measurement`)
	if sqlErr != nil {
		log.Printf("WARN: Fail to cleanup %v\n", sqlErr)
	}
}

// StoreMeasurement Saves the measurement for a day.
// The value is stored in aggregated form for the day.
// Total count, sum, min, max, avg values are updated.
func (db *PostgresDb) StoreMeasurement(ctx context.Context, day time.Time, sensorId int, value float64) {
	_, sqlErr := db.pool.Exec(ctx, `
INSERT INTO measurement (
	measurement_day, sensor_id, total_count, total_sum, avg_value, min_value, max_value) 
VALUES ($1, $2, 1, $3, $3, $3, $3)
ON CONFLICT (measurement_day, sensor_id) DO
UPDATE SET total_sum = measurement.total_sum + $3,
total_count = measurement.total_count + 1,
avg_value = (measurement.total_sum + $3) / (measurement.total_count + 1),
min_value = LEAST(measurement.min_value, $3),
max_value = GREATEST(measurement.max_value, $3)
WHERE measurement.measurement_day = $1 AND measurement.sensor_id = $2
`,
		day, sensorId, value)
	if sqlErr != nil {
		log.Printf("WARN: Fail to insert measure %v\n", sqlErr)
	}
}

// GetMeasurementStatsForDay returns a stats for a day.
// If no any measurements exists for the day then all counters will be zero.
func (db *PostgresDb) GetMeasurementStatsForDay(ctx context.Context, day time.Time, sensorId int) (*MeasurementRec, error) {
	row := db.pool.QueryRow(ctx, `
SELECT measurement_day, sensor_id, total_count, total_sum, avg_value, min_value, max_value
FROM measurement
WHERE measurement_day = $1 AND sensor_id = $2`,
		day, sensorId)

	measurement := &MeasurementRec{}
	sqlErr := row.Scan(&measurement.Day, &measurement.SensorId, &measurement.TotalCount, &measurement.TotalSum,
		&measurement.AvgValue, &measurement.MinValue, &measurement.MaxValue)
	if sqlErr == pgx.ErrNoRows {
		measurement.Day = day
		measurement.SensorId = sensorId
		return measurement, nil
	}
	if sqlErr != nil {
		return nil, sqlErr
	}
	return measurement, nil
}
