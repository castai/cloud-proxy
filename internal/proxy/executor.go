package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/castai/cloud-proxy/internal/castai/proto"
	"github.com/castai/cloud-proxy/internal/gcpauth"
)

type Executor struct {
	credentialsSrc gcpauth.GCPCredentialsSource
	client         *http.Client
}

func NewExecutor(credentialsSrc gcpauth.GCPCredentialsSource, client *http.Client) *Executor {
	return &Executor{credentialsSrc: credentialsSrc, client: client}
}

func (e *Executor) DoRequest(request *proto.HTTPRequest) (*proto.HTTPResponse, error) {
	credentials, err := e.credentialsSrc.GetDefaultCredentials()
	if err != nil {
		return nil, fmt.Errorf("cannot load GCP credentials: %w", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("cannot get access token from src (%T): %w", credentials.TokenSource, err)
	}

	httpReq, err := http.NewRequest(request.Method, request.GetPath(), bytes.NewReader(request.Body))
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %w", err)
	}
	for header, val := range request.Headers {
		for _, hv := range val.GetValue() {
			httpReq.Header.Add(header, hv)
		}
	}
	// Set the authorize header manually since we can't rely on mothership auth
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	//log.Println("Sending request:", httpReq)
	//log.Println("Auth header:", httpReq.Header.Get("Authorization"))

	httpResponse, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("unexpected err for %+v: %w", request, err)
	}

	response := &proto.HTTPResponse{
		//Status:  int32(httpResponse.StatusCode),
		Status:  http.StatusText(httpResponse.StatusCode),
		Headers: make(map[string]*proto.HeaderValue),
		Body: func() []byte {
			if httpResponse.Body == nil {
				return []byte{}
			}
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				panic(fmt.Errorf("failed to serialize body from request: %w", err))
			}
			return body
		}(),
	}
	for header, val := range httpResponse.Header {
		response.Headers[header] = &proto.HeaderValue{Value: val}
	}

	return response, nil
}
