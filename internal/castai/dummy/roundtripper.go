package dummy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type HttpOverGrpcRoundTripper struct {
	dispatcher *Dispatcher

	logger *log.Logger
}

func NewHttpOverGrpcRoundTripper(dispatcher *Dispatcher, logger *log.Logger) *HttpOverGrpcRoundTripper {
	return &HttpOverGrpcRoundTripper{dispatcher: dispatcher, logger: logger}
}

func (p *HttpOverGrpcRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	requestID := uuid.New().String()

	headers := make(map[string]string)
	for h, v := range request.Header {
		headers[h] = strings.Join(v, ",")
	}
	protoReq := &proto.HTTPRequest{
		Method: request.Method,
		Path:   request.URL.String(),
		Headers: func() map[string]*proto.HeaderValue {
			result := make(map[string]*proto.HeaderValue)
			for h, v := range request.Header {
				result[h] = &proto.HeaderValue{
					Value: make([]string, 0, len(v)),
				}
				for i := range v {
					result[h].Value = append(result[h].Value, v[i])
				}
			}
			return result
		}(),
		Body: func() []byte {
			if request.Body == nil {
				return []byte{}
			}
			body, err := io.ReadAll(request.Body)
			if err != nil {
				panic(fmt.Sprintf("Failed to read body: %v", err))
			}
			return body
		}(),
	}
	waiter, err := p.dispatcher.SendRequest(protoReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	response := <-waiter
	p.logger.Println("Received a response back from dispatcher", requestID)

	// Convert to response
	resp := &http.Response{
		//Status:     http.StatusText(int(response.Status)),
		//StatusCode: int(response.Status),
		Status: response.Status,
		Header: func() http.Header {
			headers := make(http.Header)
			for key, value := range response.Headers {
				for _, hv := range value.GetValue() {
					headers[key] = append(headers[key], hv)
				}
			}
			return headers
		}(),
		Body:          io.NopCloser(bytes.NewReader(response.Body)),
		ContentLength: int64(len(response.Body)),
		Request:       request,
	}

	return resp, nil
}
