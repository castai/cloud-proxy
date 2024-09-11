package gcp

import (
	"bytes"
	"fmt"
	"github.com/castai/cloud-proxy/internal/gcpauth"
	"io"
	"net/http"

	proto "github.com/castai/cloud-proxy/proto/v1alpha"
)

type Client struct {
	credentialsSrc gcpauth.GCPCredentialsSource
	httpClient     *http.Client
}

func NewExecutor(credentialsSrc gcpauth.GCPCredentialsSource, client *http.Client) *Client {
	return &Client{credentialsSrc: credentialsSrc, httpClient: client}
}

func (e *Client) DoRequest(request *proto.StreamCloudProxyResponse) (*proto.StreamCloudProxyRequest, error) {
	credentials, err := e.credentialsSrc.GetDefaultCredentials()
	if err != nil {
		return nil, fmt.Errorf("cannot load GCP credentials: %w", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("cannot get access token from src (%T): %w", credentials.TokenSource, err)
	}

	httpReq := request.GetHttpRequest()

	req, err := http.NewRequest(request.HttpRequest.Method, httpReq.Path, bytes.NewReader(httpReq.Body))
	if err != nil {
		errMessage := fmt.Sprintf("http.NewRequest: msgID=%v error: %v", request.GetMessageId(), err)
		return &proto.StreamCloudProxyRequest{
			Request: &proto.StreamCloudProxyRequest_Response{
				Response: &proto.ClusterResponse{
					MessageId: request.MessageId,
					HttpResponse: &proto.HTTPResponse{
						Error: &errMessage,
					},
				}}}, nil
	}

	for header, values := range httpReq.Headers {
		for _, value := range values.Value {
			req.Header.Add(header, value)
		}
	}
	// Set the authorize header manually since we can't rely on mothership auth
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	resp, err := e.httpClient.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("httpClient.Do: request %+v error: %v", request, err)
		return &proto.StreamCloudProxyRequest{
			Request: &proto.StreamCloudProxyRequest_Response{
				Response: &proto.ClusterResponse{
					MessageId: request.MessageId,
					HttpResponse: &proto.HTTPResponse{
						Error: &errMessage,
					},
				}}}, nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	headers := make(map[string]*proto.HeaderValue)
	for h, v := range resp.Header {
		headers[h] = &proto.HeaderValue{Value: v}
	}

	var bodyResp []byte
	var errMessage string
	if resp.Body != nil {
		var err error
		bodyResp, err = io.ReadAll(resp.Body)
		if err != nil {
			errMessage = fmt.Sprintf("io.ReadAll: body for msgID=%v error: %v", request, err)
		}
	}

	return &proto.StreamCloudProxyRequest{
		Request: &proto.StreamCloudProxyRequest_Response{
			Response: &proto.ClusterResponse{
				MessageId: request.MessageId,
				HttpResponse: &proto.HTTPResponse{
					Body:    bodyResp,
					Error:   &errMessage,
					Status:  int32(resp.StatusCode),
					Headers: headers,
				},
			}}}, nil
}
