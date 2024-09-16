package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"cloud-proxy/internal/e2etest"
	cloudproxyv1alpha "cloud-proxy/proto/gen/proto/v1alpha"
	compute "cloud.google.com/go/compute/apiv1"
	container "cloud.google.com/go/container/apiv1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type Server interface {
	StartServer() error
	StopServer()
	GracefulStopServer()
}

type TestSetup struct {
	cloudproxyv1alpha.UnimplementedCloudProxyAPIServer

	result bool

	grpcServer   *grpc.Server
	dispatcher   *e2etest.Dispatcher
	roundTripper *e2etest.HttpOverGrpcRoundTripper

	requestChan  chan *cloudproxyv1alpha.StreamCloudProxyResponse
	responseChan chan *cloudproxyv1alpha.StreamCloudProxyRequest

	logger *log.Logger
}

func NewTestSetup(logger *log.Logger) *TestSetup {
	requestChan, respChan := make(chan *cloudproxyv1alpha.StreamCloudProxyResponse), make(chan *cloudproxyv1alpha.StreamCloudProxyRequest)
	dispatcher := e2etest.NewDispatcher(requestChan, respChan, logger)
	roundTrip := e2etest.NewHttpOverGrpcRoundTripper(dispatcher, logger)

	dispatcher.Run()

	return &TestSetup{
		result:       true,
		dispatcher:   dispatcher,
		roundTripper: roundTrip,
		requestChan:  requestChan,
		responseChan: respChan,
		logger:       logger,
	}
}

func (srv *TestSetup) StartServer() error {
	list, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("listening: %w", err)
	}

	srv.grpcServer = grpc.NewServer()
	cloudproxyv1alpha.RegisterCloudProxyAPIServer(srv.grpcServer, srv)

	go func() {
		if err := srv.grpcServer.Serve(list); err != nil {
			srv.logger.Printf("when serving grpc: %v", err)
		}
	}()

	return nil
}

func (srv *TestSetup) StopServer() {
	srv.grpcServer.Stop()
}

func (srv *TestSetup) GracefulStopServer() {
	srv.grpcServer.Stop()
}

func (srv *TestSetup) StreamCloudProxy(stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyServer) error {
	srv.logger.Println("Received a proxy connection from client")

	//md, ok := metadata.FromIncomingContext(stream.Context())
	//if !ok {
	//	return fmt.Errorf("missing metadata")
	//}
	//if token := md.Get("authorization"); token[0] != "Token dummytoken" {
	//	fmt.Println(token)
	//	return fmt.Errorf("wrong authentication token")
	//}

	var eg errgroup.Group

	eg.Go(func() error {
		srv.logger.Println("Starting request sender")

		for {
			select {
			case req := <-srv.requestChan:
				srv.logger.Println("Sending request to cluster proxy client")

				if err := stream.Send(req); err != nil {
					srv.logger.Printf("Error sending request: %v\n", err)
				}
			case <-stream.Context().Done():
				srv.logger.Printf("stream closed, stopping sending responses")
				return nil
			}
		}
	})

	eg.Go(func() error {
		srv.logger.Println("Starting response receiver")

		for {
			in, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("stream was closed by client")
				return err
			}
			if err != nil {
				srv.logger.Printf("Error in response receiver: %v\n", err)
				return err
			}

			srv.logger.Printf("Got a response from client: %v, %v\n",
				in.GetResponse().GetMessageId(), in.GetResponse().GetHttpResponse().GetStatus())
			srv.responseChan <- in
		}
	})

	return eg.Wait()
}

func (srv *TestSetup) ExecuteTest(ctx context.Context, name string, f func(ctx context.Context, server Server, rt http.RoundTripper) error) {
	if err := f(ctx, srv, srv.roundTripper); err != nil {
		srv.logger.Printf("TEST FAILED: %s", name)
		srv.logger.Printf("error: %v", err)
		srv.result = false
	} else {
		srv.logger.Printf("TEST PASSED: %s", name)
	}
}

func GetContainerClient(ctx context.Context, rt http.RoundTripper) (*container.ClusterManagerClient, error) {
	httpClient := &http.Client{}
	httpClient.Transport = rt
	return container.NewClusterManagerRESTClient(ctx,
		option.WithoutAuthentication(),
		option.WithHTTPClient(httpClient))
}

func GetZonesClient(ctx context.Context, rt http.RoundTripper) (*compute.ZonesClient, error) {
	httpClient := &http.Client{}
	httpClient.Transport = rt
	return compute.NewZonesRESTClient(ctx,
		option.WithoutAuthentication(),
		option.WithHTTPClient(httpClient))
}
