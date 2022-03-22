package gameday

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var logger *log.Logger
var router *mux.Router

func init() {
	logger = log.New()
	logger.Out = os.Stdout
	logger.Formatter = &log.JSONFormatter{}
}

func TestHandleConfigureForm(t *testing.T) {

	req, err := http.NewRequest(http.MethodPost, "/api/v1/configure/form", nil)
	if err != nil {
		t.Fatal("Creating 'POST /api/v1/configure/form' request failed!", err)
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleConfigureForm(logger))
	handler.ServeHTTP(w, req)

	logger.Info("w.Body")

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHandleConfigure(t *testing.T) {

	payload := []byte(`{"scheme":"sqlite3","url": "sqlite3://engine.db" }`)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/configure/submit", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal("Creating 'POST /api/v1/configure/submit' request failed!", err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleConfigure(router, logger))
	handler.ServeHTTP(w, req)

	logger.Info(w.Body)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
