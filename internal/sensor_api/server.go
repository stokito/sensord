package sensor_api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"sensord/internal/db"
	"sensord/internal/models"
)

// SensorApiServer Collects measurements from sensors
type SensorApiServer struct {
	storage    db.SensorsDb
	listenAddr string
}

func NewSensorApiServer(ListenAddr string, storage db.SensorsDb) *SensorApiServer {
	return &SensorApiServer{
		listenAddr: ListenAddr,
		storage:    storage,
	}
}

func (s *SensorApiServer) Start() {
	log.Printf("NOTICE: Start sensord API server on %s\n", s.listenAddr)
	apiServerHttp := &fasthttp.Server{
		Handler:                       s.handleApiRequest,
		DisableHeaderNamesNormalizing: true,
		NoDefaultServerHeader:         true,
		NoDefaultContentType:          true,
		NoDefaultDate:                 true,
		DisablePreParseMultipartForm:  true, // we don't use multipart forms but exploits may use it
	}
	err := apiServerHttp.ListenAndServe(s.listenAddr)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("CRIT: API server http shutdown error: %s\n", err)
	}
}

var apiEndpoint = []byte("/api/v1/measurement")

func (s *SensorApiServer) handleApiRequest(reqCtx *fasthttp.RequestCtx) {
	// catch panic
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			log.Printf("ERR: Unexpected error %s\n", panicErr)
			reqCtx.Response.SetStatusCode(http.StatusInternalServerError)
		}
	}()

	uri := reqCtx.Request.URI()
	path := uri.Path()

	// if path is /api/v1/measurement
	if bytes.Equal(path, apiEndpoint) {
		// only POST is allowed
		if !reqCtx.IsPost() {
			reqCtx.Response.SetStatusCode(http.StatusMethodNotAllowed)
			return
		}
		// Get the request body
		body, err := reqCtx.Request.BodyUncompressed()
		if err != nil {
			reqCtx.Response.SetStatusCode(http.StatusUnprocessableEntity)
			return
		}
		// parse the JSON and save
		err = s.parseAndStore(body)
		if err != nil {
			reqCtx.Response.SetStatusCode(http.StatusBadRequest)
			return
		}

		reqCtx.Response.SetStatusCode(http.StatusNoContent)
		return
	}
	// 404 for the unknown URL path
	reqCtx.Response.SetStatusCode(http.StatusNotFound)
}

func (s *SensorApiServer) parseAndStore(body []byte) error {
	measurement := &models.MeasurementDto{}
	err := json.Unmarshal(body, measurement)
	if err != nil {
		return err
	}
	ctx := context.Background()
	s.storage.StoreMeasurement(ctx, measurement.Time, measurement.SensorId, measurement.Value)

	return nil
}
