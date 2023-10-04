package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sensord/internal/api"
	"sensord/internal/core"
	"sensord/internal/db"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Listen to interrupt signal Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	conf := core.LoadConfig()

	log.Printf("INFO: Running Sensor Daemon on %s\n", conf.ApiListenHttp)

	storage := db.NewPostgresDb(conf.DatabaseUrl, conf.DatabaseLog)
	dbErr := storage.Connect(ctx)
	if dbErr != nil {
		log.Fatal("CRIT: Unable to connect to database: " + dbErr.Error())
	}
	defer storage.Close()

	apiServ := api.NewApiServer(conf.ApiListenHttp, storage)
	go apiServ.Start()

	<-ctx.Done()
	log.Println("INFO: Gracefully shutting down")
	stop()
	storage.Close()
	return nil
}
