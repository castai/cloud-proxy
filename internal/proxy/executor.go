package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

func (e *Executor) DoRequest(request *proto.HttpRequest) (*proto.HttpResponse, error) {
	credentials, err := e.credentialsSrc.GetDefaultCredentials()
	if err != nil {
		return nil, fmt.Errorf("cannot load GCP credentials: %w", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("cannot get access token from src (%T): %w", credentials.TokenSource, err)
	}

	httpReq, err := http.NewRequest(request.Method, request.Url, bytes.NewReader(request.Body))
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %w", err)
	}
	for header, val := range request.Headers {
		httpReq.Header.Add(header, val)
	}
	// Set the authorize header manually since we can't rely on mothership auth
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	//log.Println("Sending request:", httpReq)
	//log.Println("Auth header:", httpReq.Header.Get("Authorization"))

	httpResponse, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("unexpected err for %+v: %w", request, err)
	}

	response := &proto.HttpResponse{
		RequestID: request.RequestID,
		Status:    int32(httpResponse.StatusCode),
		Headers:   make(map[string]string),
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
		response.Headers[header] = strings.Join(val, ",")
	}

	return response, nil
}
