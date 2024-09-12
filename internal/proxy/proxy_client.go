package proxy

import (
	"cloud-proxy/internal/cloud/gcp"
	cloudproxyv1alpha "cloud-proxy/proto/v1alpha"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
)

type StreamCloudProxyClient = cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient

type Client struct {
	executor  *gcp.Client
	log       *logrus.Logger
	clusterID string
}

func New(executor *gcp.Client, logger *logrus.Logger, clusterID string) *Client {
	return &Client{executor: executor, log: logger, clusterID: clusterID}
}

func (c *Client) Run(ctx context.Context, grpcClient cloudproxyv1alpha.CloudProxyAPIClient) error {
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

	for {
		if ctx.Err() != nil {
			return nil
		}
		for {
			in, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				c.log.Println("Reconnecting")
				break
			}
			if err != nil {
				c.log.Printf("error receiving from castai: %v; closing stream\n", err)
				err = stream.CloseSend()
				if err != nil {
					c.log.Println("error closing stream", err)
				}
				// Reconnect by stopping inner loop
				break
			}

			go func() {
				c.log.Println("Received message from server for proxying:", in.MessageId)
				resp, err := c.executor.DoRequest(in)
				if err != nil {
					c.log.Println("error executing request", err)
					// TODO: Sent error as metadata to cast
					return
				}
				c.log.Println("got response for request:", resp.GetResponse().GetMessageId())

				err = stream.Send(resp)
				if err != nil {
					c.log.Println("error sending response to CAST", err)
					return
				}
			}()
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
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("stream.Send: initial request %w", err)
	}
	return stream, nil
}

func (c *Client) run() {

}
