//go:generate mockgen -destination ./mock/cloud.go -package=mock_cloud cloud-proxy/internal/proxy CloudClient
//go:generate mockgen -destination ./mock/stream.go -package=mock_cloud cloud-proxy/internal/proxy StreamCloudProxyClient
package proxy

import (
	"bytes"
	cloudproxyv1alpha "cloud-proxy/proto/v1alpha"
	proto "cloud-proxy/proto/v1alpha"
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	KeepAliveMessageID      = "keep-alive"
	KeepAliveDefault        = 10 * time.Second
	KeepAliveTimeoutDefault = time.Minute
)

type StreamCloudProxyClient = cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient
type CloudClient interface {
	DoHTTPRequest(request *http.Request) (*http.Response, error)
}

type Client struct {
	cloudClient CloudClient
	log         *logrus.Logger
	clusterID   string

	errCount       atomic.Int64
	processedCount atomic.Int64

	lastSeen         atomic.Int64
	keepAlive        atomic.Int64
	keepAliveTimeout atomic.Int64
	version          string
}

func New(executor CloudClient, logger *logrus.Logger, clusterID, version string) *Client {
	c := &Client{
		cloudClient: executor,
		log:         logger,
		clusterID:   clusterID,
		version:     version,
	}
	c.keepAlive.Store(int64(KeepAliveDefault))
	c.keepAliveTimeout.Store(int64(KeepAliveTimeoutDefault))

	return c
}

func (c *Client) Run(ctx context.Context, grpcClient cloudproxyv1alpha.CloudProxyAPIClient) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		err := c.run(ctx, grpcClient)
		if err != nil {
			c.log.Printf("c.run: %v", err)
		}
	}
}

func (c *Client) initializeStream(ctx context.Context, proxyCastAIClient cloudproxyv1alpha.CloudProxyAPIClient) (StreamCloudProxyClient, error) {
	c.log.Println("Connecting to castai")

	stream, err := proxyCastAIClient.StreamCloudProxy(ctx)
	if err != nil {
		return nil, fmt.Errorf("proxyCastAIClient.StreamCloudProxy: %w", err)
	}

	err = stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
		Request: &cloudproxyv1alpha.StreamCloudProxyRequest_InitialRequest{
			InitialRequest: &cloudproxyv1alpha.InitialCloudProxyRequest{
				ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
					ClusterId: c.clusterID,
				},
				Version: c.version,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("stream.Send: initial request %w", err)
	}
	c.lastSeen.Store(time.Now().UnixNano())

	return stream, nil
}

func (c *Client) run(ctx context.Context, grpcClient cloudproxyv1alpha.CloudProxyAPIClient) error {
	stream, err := c.initializeStream(ctx, grpcClient)
	if err != nil {
		return fmt.Errorf("c.Connect: %w", err)
	}

	defer func() {
		err := stream.CloseSend()
		if err != nil {
			c.log.Println("error closing stream", err)
		}
	}()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	go c.sendKeepAlive(ctxWithCancel, stream)

	for {
		if !c.isAlive() {
			return fmt.Errorf("last seen too old, closing stream")
		}
		in, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream.Recv: %w", err)
		}

		go c.handleMessage(in, stream)
	}
}

func (c *Client) handleMessage(in *cloudproxyv1alpha.StreamCloudProxyResponse, stream StreamCloudProxyClient) {
	c.processConfigurationRequest(in)

	// skip processing http request if keep alive message
	if in.GetMessageId() == KeepAliveMessageID {
		c.lastSeen.Store(time.Now().UnixNano())
		return
	}

	resp := c.processHttpRequest(in.GetHttpRequest())
	err := stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
		Request: &cloudproxyv1alpha.StreamCloudProxyRequest_Response{
			Response: &cloudproxyv1alpha.ClusterResponse{
				ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
					ClusterId: c.clusterID,
				},
				MessageId:    in.GetMessageId(),
				HttpResponse: resp,
			},
		},
	})
	if err != nil {
		c.log.Errorf("error sending response for msg_id=%v %v", in.GetMessageId(), err)
	}
	return
}

func (c *Client) processConfigurationRequest(in *cloudproxyv1alpha.StreamCloudProxyResponse) {
	if in.ConfigurationRequest != nil {
		if in.ConfigurationRequest.GetKeepAlive() != 0 {
			c.keepAlive.Store(int64(in.ConfigurationRequest.GetKeepAlive()))
		}
		if in.ConfigurationRequest.GetKeepAliveTimeout() != 0 {
			c.keepAliveTimeout.Store(int64(in.ConfigurationRequest.GetKeepAliveTimeout()))
		}
	}
}

func (c *Client) processHttpRequest(req *cloudproxyv1alpha.HTTPRequest) *cloudproxyv1alpha.HTTPResponse {
	if req == nil {
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr("nil http request"),
		}
	}
	httpReq, err := c.toHTTPRequest(req)
	if err != nil {
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr(fmt.Sprintf("toHTTPRequest: %v", err)),
		}
	}
	resp, err := c.cloudClient.DoHTTPRequest(httpReq)
	if err != nil {
		c.errCount.Add(1)
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr(fmt.Sprintf("c.cloudClient.DoHTTPRequest: %v", err)),
		}
	}
	c.processedCount.Add(1)

	return c.toResponse(resp)
}

func (c *Client) isAlive() bool {
	lastSeen := c.lastSeen.Load()
	if time.Now().UnixNano()-lastSeen > c.keepAliveTimeout.Load() {
		return false
	}
	return true
}

func (c *Client) sendKeepAlive(ctx context.Context, stream StreamCloudProxyClient) {
	ticker := time.NewTimer(time.Duration(c.keepAlive.Load()))
	defer ticker.Stop()

	for {
		if !c.isAlive() {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
				Request: &cloudproxyv1alpha.StreamCloudProxyRequest_ClientStats{
					ClientStats: &cloudproxyv1alpha.ClientStats{
						ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
							ClusterId: c.clusterID,
						},
						Status: cloudproxyv1alpha.ClientStats_OK,
					},
				},
			})
			if err != nil {
				c.lastSeen.Store(0)
				c.log.Errorf("error sending keep alive message: %v", err)
				return
			}
			ticker.Reset(time.Duration(c.keepAlive.Load()))
		}
	}
}

var errBadRequest = fmt.Errorf("bad request")

func (c *Client) toHTTPRequest(req *proto.HTTPRequest) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("nil http request %w", errBadRequest)
	}

	reqHTTP, err := http.NewRequestWithContext(context.Background(), req.GetMethod(), req.GetPath(), bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: error: %v", err)
	}

	for header, values := range req.GetHeaders() {
		for _, value := range values.Value {
			reqHTTP.Header.Add(header, value)
		}
	}

	return reqHTTP, nil
}

func (c *Client) toResponse(resp *http.Response) *proto.HTTPResponse {
	if resp == nil {
		return &proto.HTTPResponse{
			Error: lo.ToPtr("nil response"),
		}
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
			errMessage = lo.ToPtr(fmt.Sprintf("io.ReadAll: body for error: %v", err))
			bodyResp = nil
		}
	}

	return &proto.HTTPResponse{
		Body:    bodyResp,
		Error:   errMessage,
		Status:  int32(resp.StatusCode),
		Headers: headers,
	}
}
