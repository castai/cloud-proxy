package dummy

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

// MockCast simulates what cast would do but runs it in the same process:
//   - server side of proxy
//   - "dispatcher" that uses proxy to send requests
//   - "client" that does GCP cloud calls
type MockCast struct {
	proxyServer *MockCastServer
}

func (mc *MockCast) Run() error {
	logger := log.New(os.Stderr, "[CAST-MOCK] ", log.LstdFlags)

	requestChan, respChan := make(chan *proto.HttpRequest), make(chan *proto.HttpResponse)

	// Start the mock server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Panicf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGCPProxyServerServer(grpcServer, NewMockCastServer(requestChan, respChan, logger))

	dispatcher := NewDispatcher(requestChan, respChan, logger)

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
	proto.UnimplementedGCPProxyServerServer

	requestChan  <-chan *proto.HttpRequest
	responseChan chan<- *proto.HttpResponse

	logger *log.Logger
}

func NewMockCastServer(requestChan <-chan *proto.HttpRequest, responseChan chan<- *proto.HttpResponse, logger *log.Logger) *MockCastServer {
	return &MockCastServer{
		requestChan:  requestChan,
		responseChan: responseChan,
		logger:       logger,
	}
}

func (msrv *MockCastServer) Proxy(stream proto.GCPProxyServer_ProxyServer) error {
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

			msrv.logger.Printf("Got a response from client: %v, %v\n", in.RequestID, in.Status)
			msrv.responseChan <- in
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	msrv.logger.Println("Closing proxy connection")
	return nil
}