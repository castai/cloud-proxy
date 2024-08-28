package proxy

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type Client struct {
	executor *Executor

	logger *log.Logger
}

func NewClient(executor *Executor, logger *log.Logger) *Client {
	return &Client{executor: executor, logger: logger}
}

func (client *Client) Run(grpcConn *grpc.ClientConn) {
	grpcClient := proto.NewCloudProxyAPIClient(grpcConn)

	// Outer loop is a dumb re-connect version
	for {
		time.Sleep(1 * time.Second)
		client.logger.Println("Connecting to castai")

		stream, err := grpcClient.StreamCloudProxy(context.Background())
		if err != nil {
			client.logger.Printf("error connecting to castai: %v\n", err)
			continue
		}

		// Inner loop handles "per-message" execution
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				client.logger.Println("Reconnecting")
				break
			}
			if err != nil {
				client.logger.Printf("error receiving from castai: %v; closing stream\n", err)
				err = stream.CloseSend()
				if err != nil {
					client.logger.Println("error closing stream", err)
				}
				// Reconnect by stopping inner loop
				break
			}

			go func() {
				client.logger.Println("Received message from server for proxying:", in.GetMessageId())
				resp, err := client.executor.DoRequest(in.GetHttpRequest())
				if err != nil {
					client.logger.Println("error executing request", err)
					// TODO: Sent error as metadata to cast
					return
				}
				client.logger.Println("got response for request:", in.GetMessageId())

				response := &proto.StreamCloudProxyRequest{
					MessageId:    in.GetMessageId(),
					HttpResponse: resp,
				}
				err = stream.Send(response)
				if err != nil {
					client.logger.Println("error sending response to CAST", err)
					return
				}
			}()
		}
	}
}
