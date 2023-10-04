package admin_api

import (
	"context"
	json "encoding/json"
	"log"
	"net/http"
	"sensord/internal/db"
	"time"
)

// AdminApiServer Admin HTTP API: reporting endpoints
type AdminApiServer struct {
	storage    db.SensorsDb
	listenAddr string
}

func NewAdminApiServer(ListenAddr string, storage db.SensorsDb) *AdminApiServer {
	return &AdminApiServer{
		listenAddr: ListenAddr,
		storage:    storage,
	}
}

func (s *AdminApiServer) Start() {
	log.Printf("NOTICE: Start sensord Admin API server on %s\n", s.listenAddr)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/stats/Total", s.handleGetStatsTotal)
	mux.HandleFunc("/api/v1/stats/EachSensor", s.handleGetStatsForEachSensor)
	mux.HandleFunc("/api/v1/stats/EachSensorAndDay", s.handleGetStatsForEachSensorAndDay)
	apiServerHttp := &http.Server{
		Addr:    s.listenAddr,
		Handler: mux,
	}
	err := apiServerHttp.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("CRIT: Admin API server http shutdown error: %s\n", err)
	}
}

func (s *AdminApiServer) handleGetStatsTotal(w http.ResponseWriter, r *http.Request) {
	// catch panic
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			log.Printf("ERR: Unexpected error %s\n", panicErr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := context.Background()
	endTime, startTime := weekAgo()
	stats, err := s.storage.GetMeasurementPeriodStatsTotal(ctx, startTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonBody, _ := json.Marshal(stats)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBody)
	return
}

func (s *AdminApiServer) handleGetStatsForEachSensor(w http.ResponseWriter, r *http.Request) {
	// catch panic
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			log.Printf("ERR: Unexpected error %s\n", panicErr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := context.Background()
	endTime, startTime := weekAgo()
	stats, err := s.storage.GetMeasurementPeriodStatsForEachSensor(ctx, startTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonBody, _ := json.Marshal(stats)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBody)
	return
}

func (s *AdminApiServer) handleGetStatsForEachSensorAndDay(w http.ResponseWriter, r *http.Request) {
	// catch panic
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			log.Printf("ERR: Unexpected error %s\n", panicErr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := context.Background()
	endTime, startTime := weekAgo()
	stats, err := s.storage.GetMeasurementPeriodStatsForEachSensorAndDay(ctx, startTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonBody, _ := json.Marshal(stats)

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBody)
	return
}

func weekAgo() (time.Time, time.Time) {
	endTime := time.Now().Truncate(24 * time.Hour)
	startTime := endTime.AddDate(0, 0, -7)
	return endTime, startTime
}
