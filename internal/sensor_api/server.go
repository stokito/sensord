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

type ApiServer struct {
	storage    db.SensorsDb
	listenAddr string
}

func NewSensorApiServer(ListenAddr string, storage db.SensorsDb) *ApiServer {
	return &ApiServer{
		listenAddr: ListenAddr,
		storage:    storage,
	}
}

func (s *ApiServer) Start() {
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

func (s *ApiServer) handleApiRequest(reqCtx *fasthttp.RequestCtx) {
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

	if bytes.Equal(path, apiEndpoint) {
		if !reqCtx.IsPost() {
			reqCtx.Response.SetStatusCode(http.StatusMethodNotAllowed)
			return
		}
		err := s.parseAndStore(reqCtx.Request.Body())
		if err != nil {
			reqCtx.Response.SetStatusCode(http.StatusBadRequest)
			return
		}

		reqCtx.Response.SetStatusCode(http.StatusNoContent)
		return
	}
	// unknown URL path
	reqCtx.Response.SetStatusCode(http.StatusNotFound)
}

func (s *ApiServer) parseAndStore(body []byte) error {
	measurement := &models.MeasurementDto{}
	err := json.Unmarshal(body, measurement)
	if err != nil {
		return err
	}
	ctx := context.Background()
	s.storage.StoreMeasurement(ctx, measurement.Time, measurement.SensorId, measurement.Value)

	return nil
}
