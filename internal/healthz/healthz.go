package healthz

import (
	"encoding/json"
	"net/http"

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

	mux.HandleFunc("/healthz", hc.healthCheck)

	return http.ListenAndServe(addr, mux)
}

func (hc *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

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
