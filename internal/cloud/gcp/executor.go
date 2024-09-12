//go:generate mockgen -destination ./mock/gcp.go -package=mock_gcp cloud-proxy/internal/cloud/gcp Credentials
package gcp

import (
	"bytes"
	"cloud-proxy/internal/cloud/gcp/gcpauth"
	proto "cloud-proxy/proto/v1alpha"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
)

type Credentials interface {
	GetToken() (string, error)
}
type Client struct {
	credentialsSrc Credentials
	httpClient     *http.Client
}

func New(credentials *gcpauth.CredentialsSource, client *http.Client) *Client {
	return &Client{credentialsSrc: credentials, httpClient: client}
}

func (c *Client) DoRequest(request *proto.StreamCloudProxyResponse) (*proto.StreamCloudProxyRequest, error) {
	reqHTTP, err := c.toGCPRequest(request)
	if err != nil {
		return &proto.StreamCloudProxyRequest{
			Request: &proto.StreamCloudProxyRequest_Response{
				Response: &proto.ClusterResponse{
					MessageId: request.MessageId,
					HttpResponse: &proto.HTTPResponse{
						Error: lo.ToPtr(fmt.Sprintf("bad request: msgID=%v error: %v", request.GetMessageId(), err)),
					},
				}}}, nil
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
				}}}, nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return c.toResponse(request.GetMessageId(), resp), nil
}

var errBadRequest = fmt.Errorf("bad request")

func (c *Client) toGCPRequest(req *proto.StreamCloudProxyResponse) (*http.Request, error) {
	if req == nil || req.GetHttpRequest() == nil {
		return nil, fmt.Errorf("nil request or request body %w", errBadRequest)
	}

	if c.credentialsSrc == nil {
		return nil, fmt.Errorf("nil credentials source %w", errBadRequest)
	}

	reqHTTP, err := http.NewRequest(req.HttpRequest.Method, req.GetHttpRequest().Path, bytes.NewReader(req.GetHttpRequest().Body))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: msgID=%v error: %v", req.GetMessageId(), err)
	}

	for header, values := range req.GetHttpRequest().Headers {
		for _, value := range values.Value {
			reqHTTP.Header.Add(header, value)
		}
	}
	token, err := c.credentialsSrc.GetToken()
	if err != nil {
		return nil, fmt.Errorf("credentialsSrc.GetToken: msgID=%v error: %v", req.GetMessageId(), err)
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
