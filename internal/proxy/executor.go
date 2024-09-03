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

func (e *Executor) DoRequest(request *proto.StreamCloudProxyResponse) (*proto.StreamCloudProxyRequest, error) {
	credentials, err := e.credentialsSrc.GetDefaultCredentials()
	if err != nil {
		return nil, fmt.Errorf("cannot load GCP credentials: %w", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("cannot get access token from src (%T): %w", credentials.TokenSource, err)
	}

	httpReq := request.HttpRequest

	req, err := http.NewRequest(request.HttpRequest.Method, httpReq.Path, bytes.NewReader(httpReq.Body))
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %w", err)
	}
	for header, values := range httpReq.Headers {
		for _, value := range values.Value {
			req.Header.Add(header, value)
		}
	}
	// Set the authorize header manually since we can't rely on mothership auth
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	//log.Println("Sending request:", httpReq)
	//log.Println("Auth header:", httpReq.Header.Get("Authorization"))

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unexpected err for %+v: %w", request, err)
	}
	defer resp.Body.Close()

	response := &proto.StreamCloudProxyRequest{
		MessageId: request.MessageId,
		HttpResponse: &proto.HTTPResponse{
			Status: int32(resp.StatusCode),
			Headers: func() map[string]*proto.HeaderValue {
				headers := make(map[string]*proto.HeaderValue, len(resp.Header))
				for h, v := range resp.Header {
					headers[h] = &proto.HeaderValue{
						Value: v,
					}
				}
				return headers
			}(),
			Body: func() []byte {
				if resp.Body == nil {
					return []byte{}
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(fmt.Errorf("failed to serialize body from request: %w", err))
				}
				return body
			}(),
		},
	}

	return response, nil
}
