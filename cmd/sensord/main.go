package main

import (
	"context"
	"log"
	"os"
	"os/signal"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Listen to interrupt signal Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	apiListenHttp := os.Getenv("LISTEN_HTTP")
	log.Printf("Running Sensor Daemon on %s\n", apiListenHttp)

	<-ctx.Done()
	stop()
	log.Println("Gracefully shutting down")
	return nil
}
