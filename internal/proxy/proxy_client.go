package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type Client struct {
	executor *Executor
}

func NewClient(executor *Executor) *Client {
	return &Client{executor: executor}
}

func (client *Client) Run(grpcConn *grpc.ClientConn) {
	grpcClient := proto.NewGCPProxyServerClient(grpcConn)

	// Outer loop is a dumb re-connect version
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("Connecting to castai")

		stream, err := grpcClient.Proxy(context.Background())
		if err != nil {
			log.Printf("error connecting to castai: %v\n", err)
			continue
		}

		// Inner loop handles "per-message" execution
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				log.Println("Reconnecting")
				break
			}
			if err != nil {
				log.Printf("error receiving from castai: %v; closing stream\n", err)
				err = stream.CloseSend()
				if err != nil {
					log.Println("error closing stream", err)
				}
				// Reconnect by stopping inner loop
				break
			}

			go func() {
				log.Println("Received message from server for proxying:", in.RequestID)
				resp, err := client.executor.DoRequest(in)
				if err != nil {
					log.Println("error executing request", err)
					// TODO: Sent error as metadata to cast
					return
				}
				log.Println("got response for request:", resp.RequestID)

				err = stream.Send(resp)
				if err != nil {
					log.Println("error sending response to CAST", err)
					return
				}
			}()
		}
	}
}
