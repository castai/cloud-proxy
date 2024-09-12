package proxy

import (
	"cloud-proxy/internal/cloud/gcp"
	cloudproxyv1alpha "cloud-proxy/proto/v1alpha"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

const (
	KeepAliveMessageID      = "keep-alive"
	KeepAliveDefault        = 10 * time.Second
	KeepAliveTimeoutDefault = time.Minute
)

type StreamCloudProxyClient = cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient

type Client struct {
	executor  *gcp.Client
	log       *logrus.Logger
	clusterID string

	errCount       atomic.Int64
	processedCount atomic.Int64

	lastSeen         atomic.Int64
	keepAlive        atomic.Int64
	keepAliveTimeout atomic.Int64
	version          string
}

func New(executor *gcp.Client, logger *logrus.Logger, clusterID, version string) *Client {
	c := &Client{
		executor:  executor,
		log:       logger,
		clusterID: clusterID,
		version:   version,
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
	if in.GetMessageId() == KeepAliveMessageID {
		c.lastSeen.Store(time.Now().UnixNano())
	}

	if in.ConfigurationRequest != nil {
		if in.ConfigurationRequest.GetKeepAlive() != 0 {
			c.keepAlive.Store(int64(in.ConfigurationRequest.GetKeepAlive()))
		}
		if in.ConfigurationRequest.GetKeepAliveTimeout() != 0 {
			c.keepAliveTimeout.Store(int64(in.ConfigurationRequest.GetKeepAliveTimeout()))
		}
	}

	if in.GetHttpRequest() != nil {
		resp := c.executor.DoRequest(in)
		err := stream.Send(resp)
		if err != nil {
			c.log.Errorf("error sending response for msg_id=%v %v", in.GetMessageId(), err)
			c.errCount.Add(1)
		} else {
			c.processedCount.Add(1)
		}
	}
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
