package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/castai/cloud-proxy/internal/castai/dummy"
	"github.com/castai/cloud-proxy/internal/gcpauth"
	"github.com/castai/cloud-proxy/internal/localtest"
	"github.com/castai/cloud-proxy/internal/proxy"
)

const (
	// TODO: Change accordingly for local testing

	projectID   = "engineering-test-353509"
	location    = "europe-north1-a"
	testCluster = "lachezar-2908"

	grpcmothership   = "grpc-lachezarts.localenv.cast.ai:443"
	apikeyMothership = "764e828ac1267f16a09f20590902ec6986d479b39e6db080da44af14ee7fbe4e"
)

var (
	runSanityTests  = flag.Bool("sanity-checks", false, "run sanity checks that validate auth loading and basic executor function")
	runMockCastTest = flag.Bool("mockcast", true, "run a test using a mock Cast.AI server")
	runRealCastTest = flag.Bool("realcast", false, "run a test using a real Cast.AI mothership")
)

func main() {
	flag.Parse()

	if runSanityTests != nil && *runSanityTests {
		log.Println("run sanity tests is true, starting")
		go func() {
			localtest.RunBasicTests(projectID, location, testCluster)
			localtest.RunProxyTest(projectID, location, testCluster)
		}()
	}

	if runMockCastTest != nil && *runMockCastTest {
		log.Println("run mockcast tests is true, starting")
		go func() {
			log.Println("Starting mock cast instance")
			mockCast := &dummy.MockCast{}
			if err := mockCast.Run(); err != nil {
				log.Panicln("Error running mock Cast:", err)
			}
		}()

		go func() {
			loggerClientProxy := log.New(os.Stderr, "[CLUSTER PROXY] ", log.LstdFlags)
			loggerClientProxy.Println("Starting proxy client")
			conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				loggerClientProxy.Panicf("Failed to connect to server: %v", err)
			}
			defer func(conn *grpc.ClientConn) {
				err := conn.Close()
				if err != nil {
					loggerClientProxy.Panicf("Failed to close gRPC connection: %v", err)
				}
			}(conn)

			src := gcpauth.GCPCredentialsSource{}
			executor := proxy.NewExecutor(src, http.DefaultClient)
			client := proxy.NewClient(executor, loggerClientProxy)
			client.Run(context.Background(), conn)
		}()
	}

	if runRealCastTest != nil && *runRealCastTest {
		log.Println("run realcast tests is true, starting with connection to mothership ")

		go func() {
			loggerClientProxy := log.New(os.Stderr, "[CLUSTER PROXY] ", log.LstdFlags)
			loggerClientProxy.Println("Starting proxy client")
			tls := credentials.NewTLS(nil)
			conn, err := grpc.Dial(grpcmothership, grpc.WithTransportCredentials(tls),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:    10 * time.Second,
					Timeout: 20 * time.Second,
				}))
			if err != nil {
				loggerClientProxy.Panicf("Failed to connect to server: %v", err)
			}
			defer func(conn *grpc.ClientConn) {
				err := conn.Close()
				if err != nil {
					loggerClientProxy.Panicf("Failed to close gRPC connection: %v", err)
				}
			}(conn)

			ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
				"authorization", fmt.Sprintf("Token %s", apikeyMothership),
			))

			src := gcpauth.GCPCredentialsSource{}
			executor := proxy.NewExecutor(src, http.DefaultClient)
			client := proxy.NewClient(executor, loggerClientProxy)
			client.Run(ctx, conn)
		}()
	}

	log.Println("Sleeping for 1h, feel free to kill me")
	time.Sleep(1 * time.Hour)
}
