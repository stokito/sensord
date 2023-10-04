package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sensord/internal/models"
	"strings"
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

const (
	DbName = "test_db"
	DbUser = "test_user"
	DbPass = "test_password"
	DbUrl  = "postgres://test_user:test_password@localhost:%s/test_db?sslmode=disable"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	databaseUrl, container := startPostgreSqlContainer(ctx)
	if databaseUrl == "" {
		return
	}
	// remove test container
	defer container.Terminate(context.Background())

	// get location of test
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	testDir := filepath.Dir(path)
	// if the test executing directly then we need to roll to upper folder
	if strings.HasSuffix(testDir, "internal/db") {
		testDir, _ = strings.CutSuffix(testDir, "internal/db")
	}
	pathToMigrationFiles := "file://" + testDir + "migration"

	migration, err := migrate.New(pathToMigrationFiles, databaseUrl)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer migration.Close()

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Printf("Error: %s\n", err)
		return
	}

	log.Println("migration done")

	databaseUrlAndSchema := databaseUrl + "&search_path=sensors"
	storage = NewPostgresDb(databaseUrlAndSchema, true)
	dbErr := storage.Connect(ctx)
	if dbErr != nil {
		fmt.Printf("Connection error: %s\n", dbErr)
		return
	}
	defer storage.Close()
	defer storage.Cleanup(ctx)

	os.Exit(m.Run())
}

func startPostgreSqlContainer(ctx context.Context) (string, testcontainers.Container) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": DbPass,
		"POSTGRES_USER":     DbUser,
		"POSTGRES_DB":       DbName,
	}
	var port = "5432/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:alpine",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		fmt.Printf("failed to start container: %v", err)
		return "", nil
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Printf("failed to get container external port: %v", err)
		return "", nil
	}

	log.Println("PostgreSQL container is ready and running at port: ", p.Port())
	time.Sleep(time.Second)

	databaseUrl := fmt.Sprintf(DbUrl, p.Port())
	return databaseUrl, container
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
