package localtest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

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
	fmt.Println("Sending request to executor")
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
	response, err := p.executor.DoRequest(protoReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	// Convert to http response
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
