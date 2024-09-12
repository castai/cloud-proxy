//go:generate mockgen -destination ./mock/gcp.go -package=mock_gcp cloud-proxy/internal/cloud/gcp Credentials
package gcp

import (
	"bytes"
	"cloud-proxy/internal/cloud/gcp/gcpauth"
	proto "cloud-proxy/proto/v1alpha"
	"context"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
)

type Credentials interface {
	GetToken() (string, error)
}
type Client struct {
	credentials Credentials
	httpClient  *http.Client
}

func New(credentials *gcpauth.CredentialsSource, client *http.Client) *Client {
	return &Client{credentials: credentials, httpClient: client}
}

func (c *Client) DoRequest(request *proto.StreamCloudProxyResponse) *proto.StreamCloudProxyRequest {
	reqHTTP, err := c.toGCPRequest(request.GetMessageId(), request.GetHttpRequest())
	if err != nil {
		return &proto.StreamCloudProxyRequest{
			Request: &proto.StreamCloudProxyRequest_Response{
				Response: &proto.ClusterResponse{
					MessageId: request.MessageId,
					HttpResponse: &proto.HTTPResponse{
						Error: lo.ToPtr(fmt.Sprintf("bad request: msgID=%v error: %v", request.GetMessageId(), err)),
					},
				}}}
	}
	resp, err := c.httpClient.Do(reqHTTP)
	if err != nil {
		return &proto.StreamCloudProxyRequest{
			Request: &proto.StreamCloudProxyRequest_Response{
				Response: &proto.ClusterResponse{
					MessageId: request.MessageId,
					HttpResponse: &proto.HTTPResponse{
						Error: lo.ToPtr(fmt.Sprintf("httpClient.Do: request %+v error: %v", request, err)),
					},
				}}}
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return c.toResponse(request.GetMessageId(), resp)
}

var errBadRequest = fmt.Errorf("bad request")

func (c *Client) toGCPRequest(msgID string, req *proto.HTTPRequest) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("nil http request %w", errBadRequest)
	}

	reqHTTP, err := http.NewRequestWithContext(context.Background(), req.GetMethod(), req.GetPath(), bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: msgID=%v error: %v", msgID, err)
	}

	for header, values := range req.GetHeaders() {
		for _, value := range values.Value {
			reqHTTP.Header.Add(header, value)
		}
	}

	token, err := c.credentials.GetToken()
	if err != nil {
		return nil, fmt.Errorf("credentialsSrc.GetToken: msgID=%v error: %v", msgID, err)
	}
	// Set the authorize header manually since we can't rely on mothership auth
	reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return reqHTTP, nil
}

func (c *Client) toResponse(msgID string, resp *http.Response) *proto.StreamCloudProxyRequest {
	if resp == nil {
		return &proto.StreamCloudProxyRequest{
			Request: &proto.StreamCloudProxyRequest_Response{
				Response: &proto.ClusterResponse{
					MessageId: msgID,
					HttpResponse: &proto.HTTPResponse{
						Error: lo.ToPtr("nil response"),
					},
				}}}
	}
	var headers map[string]*proto.HeaderValue
	if resp.Header != nil {
		headers = make(map[string]*proto.HeaderValue)
		for h, v := range resp.Header {
			headers[h] = &proto.HeaderValue{Value: v}
		}
	}
	var bodyResp []byte
	var errMessage *string
	if resp.Body != nil {
		var err error
		bodyResp, err = io.ReadAll(resp.Body)
		if err != nil {
			errMessage = lo.ToPtr(fmt.Sprintf("io.ReadAll: body for msgID=%v error: %v", msgID, err))
			bodyResp = nil
		}
	}

	return &proto.StreamCloudProxyRequest{
		Request: &proto.StreamCloudProxyRequest_Response{
			Response: &proto.ClusterResponse{
				MessageId: msgID,
				HttpResponse: &proto.HTTPResponse{
					Body:    bodyResp,
					Error:   errMessage,
					Status:  int32(resp.StatusCode),
					Headers: headers,
				},
			}}}
}
