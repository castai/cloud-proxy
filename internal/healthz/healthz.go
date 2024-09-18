package healthz

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	log *logrus.Logger
}

func NewServer(log *logrus.Logger) *Server {
	return &Server{
		log: log,
	}
}

func (hc *Server) Run(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/readyz", hc.readyCheck)
	mux.HandleFunc("/livez", hc.liveCheck)

	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}

	return server.ListenAndServe()
}

func (hc *Server) readyCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")

	// TODO: Implement proper readiness checks.

	response := map[string]string{
		"status": "ok",
	}

	body, err := json.Marshal(response)
	if err != nil {
		hc.log.WithError(err).Errorf("Failed to marshal readiness check response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		hc.log.WithError(err).Errorf("Failed to write response body")
	}
}

func (hc *Server) liveCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")

	// TODO: Implement proper liveness checks.

	response := map[string]string{
		"status": "ok",
	}

	body, err := json.Marshal(response)
	if err != nil {
		hc.log.WithError(err).Errorf("Failed to marshal readiness check response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		hc.log.WithError(err).Errorf("Failed to write response body")
	}
}
