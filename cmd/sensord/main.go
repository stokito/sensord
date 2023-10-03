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

	// create config from envs
	conf := &config.SensordConf{
		ApiListenHttp: os.Getenv("LISTEN_HTTP"),
		DatabaseUrl:   os.Getenv("DB_URL"),
		DatabaseLog:   os.Getenv("DB_LOG") == "true",
	}
	config.Conf = conf

	log.Printf("INFO: Running Sensor Daemon on %s\n", conf.ApiListenHttp)

	dbErr := db.DbConnect(ctx)
	if dbErr != nil {
		log.Fatal("CRIT: Unable to connect to database: " + dbErr.Error())
	}
	log.Printf("INFO: Connected to database\n")

	<-ctx.Done()
	stop()
	db.DbClose()
	log.Println("INFO: Gracefully shutting down")
	return nil
}