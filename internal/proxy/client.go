//go:generate mockgen -package=mock_proxy -source $GOFILE -destination mock/$GOFILE .
//go:generate mockgen -package=mock_proxy -destination mock/stream.go cloud-proxy/proto/gen/proto/v1alpha CloudProxyAPI_StreamCloudProxyClient

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	cloudproxyv1alpha "cloud-proxy/proto/gen/proto/v1alpha"
)

const (
	KeepAliveMessageID = "keep-alive"
)

type CloudClient interface {
	DoHTTPRequest(request *http.Request) (*http.Response, error)
}

type Client struct {
	grpcConn    *grpc.ClientConn
	cloudClient CloudClient
	log         *logrus.Logger
	podName     string
	clusterID   string

	errCount       atomic.Int64
	processedCount atomic.Int64

	lastSeen         atomic.Int64
	lastSeenError    atomic.Pointer[error]
	keepAlive        atomic.Int64
	keepAliveTimeout atomic.Int64
	version          string
}

func New(grpcConn *grpc.ClientConn, cloudClient CloudClient, logger *logrus.Logger, podName, clusterID, version string, keepalive, keepaliveTimeout time.Duration) *Client {
	c := &Client{
		grpcConn:    grpcConn,
		cloudClient: cloudClient,
		log:         logger,
		podName:     podName,
		clusterID:   clusterID,
		version:     version,
	}
	c.keepAlive.Store(int64(keepalive))
	c.keepAliveTimeout.Store(int64(keepaliveTimeout))

	return c
}

func (c *Client) Run(ctx context.Context) error {
	t := time.NewTimer(time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			c.log.Info("Starting proxy client")
			stream, closeStream, err := c.getStream(ctx)
			if err != nil {
				c.log.Errorf("Could not get stream, restarting proxy client in %vs: %v", time.Duration(c.keepAlive.Load()).Seconds(), err)
				t.Reset(time.Duration(c.keepAlive.Load()))
				continue
			}

			err = c.run(ctx, stream, closeStream)
			if err != nil {
				c.log.Errorf("Restarting proxy client in %vs: due to error: %v", time.Duration(c.keepAlive.Load()).Seconds(), err)
				t.Reset(time.Duration(c.keepAlive.Load()))
			}
		}
	}
}

func (c *Client) getStream(ctx context.Context) (cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient, func(), error) {
	c.log.Info("Connecting to castai")
	apiClient := cloudproxyv1alpha.NewCloudProxyAPIClient(c.grpcConn)
	stream, err := apiClient.StreamCloudProxy(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("proxyCastAIClient.StreamCloudProxy: %w", err)
	}

	c.log.Info("Connected to castai, sending initial metadata")
	return stream, func() {
		err := stream.CloseSend()
		if err != nil {
			c.log.Errorf("error closing stream %v", err)
		}
	}, nil
}

func (c *Client) sendInitialRequest(stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient) error {
	c.log.Info("Seding initial request to castai")

	err := stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
		Request: &cloudproxyv1alpha.StreamCloudProxyRequest_InitialRequest{
			InitialRequest: &cloudproxyv1alpha.InitialCloudProxyRequest{
				ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
					PodName:   c.podName,
					ClusterId: c.clusterID,
				},
				Version: c.version,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("stream.Send: initial request %w", err)
	}
	c.lastSeen.Store(time.Now().UnixNano())
	c.lastSeenError.Store(nil)

	c.log.Info("Stream to castai started successfully")

	return nil
}

func (c *Client) run(ctx context.Context, stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient, closeStream func()) error {
	defer closeStream()

	err := c.sendInitialRequest(stream)
	if err != nil {
		return fmt.Errorf("c.Connect: %w", err)
	}

	go c.sendKeepAlive(stream)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-stream.Context().Done():
				return
			default:
				if !c.isAlive() {
					return
				}
			}

			c.log.Debugf("Polling stream for messages")

			in, err := stream.Recv()
			if err != nil {
				c.log.Errorf("stream.Recv: got error: %v", err)
				c.lastSeen.Store(0)
				c.lastSeenError.Store(&err)
				return
			}

			c.log.Debugf("Handling message from castai")
			go c.handleMessage(in, stream)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stream.Context().Done():
			return fmt.Errorf("stream closed %w", stream.Context().Err())
		case <-time.After(time.Duration(c.keepAlive.Load())):
			if !c.isAlive() {
				if err := c.lastSeenError.Load(); err != nil {
					return fmt.Errorf("received error: %w", *err)
				}
				return fmt.Errorf("last seen too old, closing stream")
			}
		}
	}
}

func (c *Client) handleMessage(in *cloudproxyv1alpha.StreamCloudProxyResponse, stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient) {
	if in == nil {
		c.log.Error("nil message")
		return
	}
	c.processConfigurationRequest(in)

	// skip processing http request if keep alive message.
	if in.GetMessageId() == KeepAliveMessageID {
		c.lastSeen.Store(time.Now().UnixNano())
		c.log.Debugf("Received keep-alive message from castai for %s", in.GetClientMetadata().GetClusterId())
		return
	}

	c.log.Debugf("Received request for proxying msg_id=%v from castai", in.GetMessageId())
	resp := c.processHTTPRequest(in.GetHttpRequest())
	if resp.GetError() != "" {
		c.log.Errorf("Failed to proxy request msg_id=%v with %v", in.GetMessageId(), resp.GetError())
	} else {
		c.log.Debugf("Proxied request msg_id=%v, sending response to castai", in.GetMessageId())
	}
	err := stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
		Request: &cloudproxyv1alpha.StreamCloudProxyRequest_Response{
			Response: &cloudproxyv1alpha.ClusterResponse{
				ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
					PodName:   c.podName,
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
}

func (c *Client) processConfigurationRequest(in *cloudproxyv1alpha.StreamCloudProxyResponse) {
	if in.ConfigurationRequest != nil {
		if in.ConfigurationRequest.GetKeepAlive() != 0 {
			c.keepAlive.Store(in.ConfigurationRequest.GetKeepAlive())
		}
		if in.ConfigurationRequest.GetKeepAliveTimeout() != 0 {
			c.keepAliveTimeout.Store(in.ConfigurationRequest.GetKeepAliveTimeout())
		}
	}
}

func (c *Client) processHTTPRequest(req *cloudproxyv1alpha.HTTPRequest) *cloudproxyv1alpha.HTTPResponse {
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

	return time.Now().UnixNano()-lastSeen <= c.keepAliveTimeout.Load()
}

func (c *Client) sendKeepAlive(stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient) {
	ticker := time.NewTimer(time.Duration(c.keepAlive.Load()))
	defer ticker.Stop()

	c.log.Info("Starting keep-alive loop")
	for {
		select {
		case <-stream.Context().Done():
			c.log.Infof("Stopping keep-alive loop: stream ended with %v", stream.Context().Err())
			return
		case <-ticker.C:
			if !c.isAlive() {
				c.log.Info("Stopping keep-alive loop: client connection is not alive")
				return
			}
			c.log.Debug("Sending keep-alive to castai")
			err := stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
				Request: &cloudproxyv1alpha.StreamCloudProxyRequest_ClientStats{
					ClientStats: &cloudproxyv1alpha.ClientStats{
						ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
							PodName:   c.podName,
							ClusterId: c.clusterID,
						},
						Stats: &cloudproxyv1alpha.ClientStats_Stats{
							Status: cloudproxyv1alpha.ClientStats_Stats_OK,
						},
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

func (c *Client) toHTTPRequest(req *cloudproxyv1alpha.HTTPRequest) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("nil http request %w", errBadRequest)
	}

	reqHTTP, err := http.NewRequestWithContext(context.Background(), req.GetMethod(), req.GetPath(), bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: error: %w", err)
	}

	for header, values := range req.GetHeaders() {
		for _, value := range values.Value {
			reqHTTP.Header.Add(header, value)
		}
	}

	return reqHTTP, nil
}

func (c *Client) toResponse(resp *http.Response) *cloudproxyv1alpha.HTTPResponse {
	if resp == nil {
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr("nil response"),
		}
	}
	var headers map[string]*cloudproxyv1alpha.HeaderValue
	if resp.Header != nil {
		headers = make(map[string]*cloudproxyv1alpha.HeaderValue)
		for h, v := range resp.Header {
			headers[h] = &cloudproxyv1alpha.HeaderValue{Value: v}
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

	return &cloudproxyv1alpha.HTTPResponse{
		Body:    bodyResp,
		Error:   errMessage,
		Status:  int64(resp.StatusCode),
		Headers: headers,
	}
}
