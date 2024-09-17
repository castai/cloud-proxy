package healthz

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type ProxyClient interface {
	IsAlive() bool
}

type Server struct {
	log         *logrus.Logger
	proxyClient ProxyClient
}

func NewServer(log *logrus.Logger, proxyClient ProxyClient) *Server {
	return &Server{
		log:         log,
		proxyClient: proxyClient,
	}
}

func (hc *Server) Run(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", hc.healthCheck)

	return http.ListenAndServe(addr, mux)
}

func (hc *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	status := true
	response := make(map[string]string)

	if hc.proxyClient.IsAlive() {
		response["proxyClient"] = "alive"
	} else {
		response["proxyClient"] = "not alive"
		status = false
	}

	w.Header().Set("content-type", "application/json")

	body, err := json.Marshal(response)
	if err != nil {
		hc.log.WithError(err).Errorf("Failed to marshal readiness check response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if status {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if _, err := w.Write(body); err != nil {
		hc.log.WithError(err).Errorf("Failed to write response body")
	}
}
