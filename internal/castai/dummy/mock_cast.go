package dummy

import (
	"fmt"
	cloudproxyv1alpha "github.com/castai/cloud-proxy/proto/gen/proto/v1alpha"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"cloud-proxy/internal/e2etest"
)

// MockCast simulates what cast would do but runs it in the same process:
//   - server side of proxy
//   - "dispatcher" that uses proxy to send requests
//   - "client" that does GCP cloud calls
type MockCast struct {
}

func (mc *MockCast) Run() error {
	logger := log.New(os.Stderr, "[CAST-MOCK] ", log.LstdFlags)

	requestChan, respChan := make(chan *cloudproxyv1alpha.StreamCloudProxyResponse), make(chan *cloudproxyv1alpha.StreamCloudProxyRequest)

	// Start the mock server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Panicf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	cloudproxyv1alpha.RegisterCloudProxyAPIServer(grpcServer, NewMockCastServer(requestChan, respChan, logger))

	dispatcher := e2etest.NewDispatcher(requestChan, respChan, logger)

	epMock, err := newMockEP(dispatcher, logger)
	if err != nil {
		logger.Panicf("Failed to create ep mock: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Start the "sender" simulation
	go func() {
		epMock.simulateActivity()
		wg.Done()
	}()

	// Start the server simulation
	go func() {
		defer wg.Done()
		dispatcher.Run()

		if err := grpcServer.Serve(listener); err != nil {
			logger.Panicf("Failed to serve mock cast: %v", err)
		}
	}()

	wg.Wait()

	return nil
}

type MockCastServer struct {
	cloudproxyv1alpha.UnimplementedCloudProxyAPIServer

	requestChan  <-chan *cloudproxyv1alpha.StreamCloudProxyResponse
	responseChan chan<- *cloudproxyv1alpha.StreamCloudProxyRequest

	logger *log.Logger
}

func NewMockCastServer(requestChan <-chan *cloudproxyv1alpha.StreamCloudProxyResponse, responseChan chan<- *cloudproxyv1alpha.StreamCloudProxyRequest, logger *log.Logger) *MockCastServer {
	return &MockCastServer{
		requestChan:  requestChan,
		responseChan: responseChan,
		logger:       logger,
	}
}

func (msrv *MockCastServer) Proxy(stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyServer) error {
	msrv.logger.Println("Received a proxy connection from client")

	var eg errgroup.Group

	// TODO: errs
	eg.Go(func() error {
		msrv.logger.Println("Starting request sender")

		for req := range msrv.requestChan {
			msrv.logger.Println("Sending request to cluster proxy client")

			if err := stream.Send(req); err != nil {
				msrv.logger.Printf("Error sending request: %v\n", err)
			}
		}
		return nil
	})

	eg.Go(func() error {
		msrv.logger.Println("Starting response receiver")

		for {
			in, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("stream was closed by client")
				return err
			}
			if err != nil {
				msrv.logger.Printf("Error in response receiver: %v\n", err)
				return err
			}

			msrv.logger.Printf("Got a response from client: %v, %v\n", in.GetResponse().GetMessageId(), in.GetResponse().GetHttpResponse().GetStatus())
			msrv.responseChan <- in
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	msrv.logger.Println("Closing proxy connection")
	return nil
}
