package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sensord/internal/config"
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

	config.LoadConfig()

	log.Printf("INFO: Running Sensor Daemon on %s\n", config.Conf.ApiListenHttp)

	dbConn := db.NewPostgresDb(config.Conf.DatabaseUrl, config.Conf.DatabaseLog)
	dbErr := dbConn.Connect(ctx)
	if dbErr != nil {
		log.Fatal("CRIT: Unable to connect to database: " + dbErr.Error())
	}
	defer dbConn.Close()

	<-ctx.Done()
	log.Println("INFO: Gracefully shutting down")
	stop()
	dbConn.Close()
	return nil
}
