package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sensord/internal/admin_api"
	"sensord/internal/core"
	"sensord/internal/db"
	"sensord/internal/sensor_api"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Listen to interrupt signal Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// Load config from envs
	conf := core.LoadConfig()

	log.Printf("INFO: Running Sensor Daemon on %s\n", conf.SensorApiListenHttp)

	storage := db.NewPostgresDb(conf.DatabaseUrl, conf.DatabaseLog)
	dbErr := storage.Connect(ctx)
	if dbErr != nil {
		log.Fatal("CRIT: Unable to connect to database: " + dbErr.Error())
	}
	defer storage.Close()

	// start Sensor API server endpoints
	sensorApiServ := sensor_api.NewSensorApiServer(conf.SensorApiListenHttp, storage)
	go sensorApiServ.Start()
	// start Admin API server endpoints
	adminApiServ := admin_api.NewAdminApiServer(conf.AdminApiListenHttp, storage)
	go adminApiServ.Start()

	// Wait until the main context is canceled by Ctrl+C
	<-ctx.Done()
	log.Println("INFO: Gracefully shutting down")
	stop()
	storage.Close()
	return nil
}
