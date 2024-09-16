package e2etest

import (
	"bytes"
	cloudproxyv1alpha "cloud-proxy/proto/gen/proto/v1alpha"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
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

	headers := make(map[string]*cloudproxyv1alpha.HeaderValue)
	for h, v := range request.Header {
		headers[h] = &cloudproxyv1alpha.HeaderValue{Value: v}
	}

	protoReq := &cloudproxyv1alpha.StreamCloudProxyResponse{
		MessageId: requestID,
		HttpRequest: &cloudproxyv1alpha.HTTPRequest{
			Method:  request.Method,
			Path:    request.URL.String(),
			Headers: headers,
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
		},
	}
	waiter, err := p.dispatcher.SendRequest(protoReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	response := <-waiter
	p.logger.Println("Received a response back from dispatcher", requestID)

	// Convert to response
	resp := &http.Response{
		StatusCode: int(response.GetResponse().GetHttpResponse().GetStatus()),
		Header: func() http.Header {
			headers := make(http.Header)
			for key, value := range response.GetResponse().GetHttpResponse().GetHeaders() {
				for _, v := range value.Value {
					headers.Add(key, v)
				}
			}
			return headers
		}(),
		Body:          io.NopCloser(bytes.NewReader(response.GetResponse().GetHttpResponse().GetBody())),
		ContentLength: int64(len(response.GetResponse().GetHttpResponse().GetBody())),
		Request:       request,
	}

	return resp, nil
}
