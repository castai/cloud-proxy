package localtest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/castai/cloud-proxy/internal/castai/proto"
	"github.com/castai/cloud-proxy/internal/proxy"
)

// RoundTripper does proxying via Executor instance directly in-process.
// Useful to test without a grpc connection or Cast at all; just to isolate if a http request can be modified successfully.
type RoundTripper struct {
	executor *proxy.Executor
}

func NewProxyRoundTripper(executor *proxy.Executor) *RoundTripper {
	return &RoundTripper{executor: executor}
}

func (p *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	fmt.Println("Sending request to dispatcher")
	requestID := uuid.New().String()

	headers := make(map[string]string)
	for h, v := range request.Header {
		headers[h] = strings.Join(v, ",")
	}
	protoReq := &proto.HttpRequest{
		RequestID: requestID,
		Method:    request.Method,
		Url:       request.URL.String(),
		Headers:   headers,
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
	response, err := p.executor.DoRequest(protoReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	// Convert to http response
	resp := &http.Response{
		Status:     http.StatusText(int(response.Status)),
		StatusCode: int(response.Status),
		Header: func() http.Header {
			headers := make(http.Header)
			for key, value := range response.Headers {
				headers[key] = strings.Split(value, ",")
			}
			return headers
		}(),
		Body:          io.NopCloser(bytes.NewReader(response.Body)),
		ContentLength: int64(len(response.Body)),
		Request:       request,
	}

	return resp, nil
}
