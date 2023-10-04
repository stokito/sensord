package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"log"
	"sensord/internal/models"
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
		log.Printf("ERROR: Fail to cleanup %v\n", sqlErr)
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
		log.Printf("ERROR: Fail to insert measure %v\n", sqlErr)
	}
}

// GetMeasurementStatsForDay returns a stats for a day.
// If no any measurements exists for the day then all counters will be zero.
func (db *PostgresDb) GetMeasurementStatsForDay(ctx context.Context, day time.Time, sensorId int) (*models.MeasurementRec, error) {
	row := db.pool.QueryRow(ctx, `
SELECT total_count, total_sum, avg_value, min_value, max_value
FROM measurement
WHERE measurement_day = $1 AND sensor_id = $2`,
		day, sensorId)

	measurement := &models.MeasurementRec{}
	sqlErr := row.Scan(&measurement.TotalCount, &measurement.TotalSum,
		&measurement.AvgValue, &measurement.MinValue, &measurement.MaxValue)
	if sqlErr == pgx.ErrNoRows {
		return measurement, nil
	}
	if sqlErr != nil {
		return nil, sqlErr
	}
	return measurement, nil
}

// GetMeasurementPeriodStatsTotal returns a stats for a period e.g. day, week.
// If no any measurements exists for the day then all counters will be zero.
func (db *PostgresDb) GetMeasurementPeriodStatsTotal(ctx context.Context, periodStart, periodEnd time.Time) (*models.MeasurementRec, error) {
	row := db.pool.QueryRow(ctx, `
SELECT
	SUM(total_count) AS total_count,
	SUM(total_sum) AS total_sum,
	SUM(total_sum) / SUM(total_count) AS avg_value,
	MIN(min_value) AS min_value,
	MAX(max_value) AS max_value
FROM measurement
WHERE measurement_day >= $1 AND measurement_day < $2
HAVING COUNT(*) > 0 -- remove records with all NULL
`,
		periodStart, periodEnd)

	measurement := &models.MeasurementRec{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		SensorId:    0,
	}
	scanErr := row.Scan(&measurement.TotalCount, &measurement.TotalSum,
		&measurement.AvgValue, &measurement.MinValue, &measurement.MaxValue)

	if scanErr == pgx.ErrNoRows {
		return measurement, nil
	}
	if scanErr != nil {
		log.Printf("ERROR: scan error %v\n", scanErr)
		return nil, scanErr
	}
	return measurement, nil
}

// GetMeasurementPeriodStatsForEachSensor returns a stats for a period e.g. day, week.
// If no any measurements exists for the day then all counters will be zero.
func (db *PostgresDb) GetMeasurementPeriodStatsForEachSensor(ctx context.Context, periodStart, periodEnd time.Time) ([]*models.MeasurementRec, error) {
	stats := []*models.MeasurementRec{}

	rows, sqlErr := db.pool.Query(ctx, `
SELECT
	sensor_id,
	SUM(total_count) AS total_count,
	SUM(total_sum) AS total_sum,
	SUM(total_sum) / SUM(total_count) AS avg_value,
	MIN(min_value) AS min_value,
	MAX(max_value) AS max_value
FROM measurement
WHERE measurement_day >= $1 AND measurement_day < $2
GROUP BY sensor_id
ORDER BY sensor_id
`,
		periodStart, periodEnd)

	if sqlErr != nil {
		return nil, sqlErr
	}
	defer rows.Close()

	for rows.Next() {
		measurement := &models.MeasurementRec{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
		}
		scanErr := rows.Scan(&measurement.SensorId, &measurement.TotalCount, &measurement.TotalSum,
			&measurement.AvgValue, &measurement.MinValue, &measurement.MaxValue)
		if scanErr != nil {
			log.Printf("ERROR: scan error %v\n", scanErr)
			continue
		}
		stats = append(stats, measurement)
	}
	return stats, nil
}

// GetMeasurementPeriodStatsForEachSensorAndDay returns a stats for a period e.g. day, week.
// If no any measurements exists for the day then all counters will be zero.
func (db *PostgresDb) GetMeasurementPeriodStatsForEachSensorAndDay(ctx context.Context, periodStart, periodEnd time.Time) ([]*models.MeasurementRec, error) {
	stats := []*models.MeasurementRec{}

	rows, sqlErr := db.pool.Query(ctx, `
SELECT
	sensor_id,
	measurement_day,
	SUM(total_count) AS total_count,
	SUM(total_sum) AS total_sum,
	SUM(total_sum) / SUM(total_count) AS avg_value,
	MIN(min_value) AS min_value,
	MAX(max_value) AS max_value
FROM measurement
WHERE measurement_day >= $1 AND measurement_day < $2
GROUP BY sensor_id, measurement_day
ORDER BY sensor_id, measurement_day
`,
		periodStart, periodEnd)

	if sqlErr != nil {
		return nil, sqlErr
	}
	defer rows.Close()

	for rows.Next() {
		measurement := &models.MeasurementRec{}
		scanErr := rows.Scan(&measurement.SensorId, &measurement.PeriodStart, &measurement.TotalCount, &measurement.TotalSum,
			&measurement.AvgValue, &measurement.MinValue, &measurement.MaxValue)
		if scanErr != nil {
			log.Printf("ERROR: scan error %v\n", scanErr)
			continue
		}
		measurement.PeriodEnd = measurement.PeriodStart.AddDate(0, 0, 1)
		stats = append(stats, measurement)
	}
	return stats, nil
}
