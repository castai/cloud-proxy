package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/castai/cloud-proxy/internal/castai/dummy"
	"github.com/castai/cloud-proxy/internal/gcpauth"
	"github.com/castai/cloud-proxy/internal/localtest"
	"github.com/castai/cloud-proxy/internal/proxy"
)

const (
	// TODO: Change accordingly for local testing

	projectID   = "engineering-test-353509"
	location    = "europe-north1-a"
	testCluster = "lachezar-2708"
)

var (
	runSanityTests  = flag.Bool("sanity-checks", false, "run sanity checks that validate auth loading and basic executor function")
	runMockCastTest = flag.Bool("mockcast", true, "run a test using a mock Cast.AI server")
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
			log.Println("Starting proxy client")
			conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Panicf("Failed to connect to server: %v", err)
			}
			defer func(conn *grpc.ClientConn) {
				err := conn.Close()
				if err != nil {
					log.Panicf("Failed to close gRPC connection: %v", err)
				}
			}(conn)

			src := gcpauth.GCPCredentialsSource{}
			executor := proxy.NewExecutor(src, http.DefaultClient)
			client := proxy.NewClient(executor)
			client.Run(conn)
		}()
	}

	log.Println("Sleeping for 1h, feel free to kill me")
	time.Sleep(1 * time.Hour)
}
