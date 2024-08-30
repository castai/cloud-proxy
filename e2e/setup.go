package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	compute "cloud.google.com/go/compute/apiv1"
	container "cloud.google.com/go/container/apiv1"
	"github.com/castai/cloud-proxy/internal/castai/proto"
	"github.com/castai/cloud-proxy/internal/e2etest"
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
	result bool

	proto.UnimplementedGCPProxyServerServer

	grpcServer   *grpc.Server
	dispatcher   *e2etest.Dispatcher
	roundTripper *e2etest.HttpOverGrpcRoundTripper

	requestChan  chan *proto.HttpRequest
	responseChan chan *proto.HttpResponse

	logger *log.Logger
}

func NewTestSetup(grpcSrv *grpc.Server, logger *log.Logger) *TestSetup {
	requestChan, respChan := make(chan *proto.HttpRequest), make(chan *proto.HttpResponse)
	dispatcher := e2etest.NewDispatcher(requestChan, respChan, logger)
	roundTrip := e2etest.NewHttpOverGrpcRoundTripper(dispatcher, logger)

	dispatcher.Run()

	return &TestSetup{
		result:       true,
		grpcServer:   grpcSrv,
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

func (srv *TestSetup) Proxy(stream proto.GCPProxyServer_ProxyServer) error {
	srv.logger.Println("Received a proxy connection from client")

	var eg errgroup.Group

	eg.Go(func() error {
		srv.logger.Println("Starting request sender")

		for req := range srv.requestChan {
			srv.logger.Println("Sending request to cluster proxy client")

			if err := stream.Send(req); err != nil {
				srv.logger.Printf("Error sending request: %v\n", err)
			}
		}
		return nil
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

			srv.logger.Printf("Got a response from client: %v, %v\n", in.RequestID, in.Status)
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
