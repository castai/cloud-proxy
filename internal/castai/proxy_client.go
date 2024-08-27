package castai

//
//func RunProxyClient() {
//	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		log.Fatalf("Failed to connect to server: %v", err)
//	}
//	defer func(conn *grpc.ClientConn) {
//		err := conn.Close()
//		if err != nil {
//			log.Fatalf("Failed to close gRPC connection: %v", err)
//		}
//	}(conn)
//
//	client := proto.NewGCPProxyServerClient(conn)
//	executor := NewExecutor()
//
//	// Start the proxy stream
//	stream, err := client.Proxy(context.Background())
//	if err != nil {
//		log.Fatalf("Failed to create stream: %v", err)
//	}
//
//	go func() {
//		// Receive http requests to handle from server
//		for {
//			in, err := stream.Recv()
//			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				log.Fatalf("Failed to receive message: %v", err)
//			}
//			log.Printf("Received message from server: %+v", in)
//
//			go func() {
//				response, err := executor.DoRequest(in)
//				if err != nil {
//					fmt.Println("Failed to execute request", err)
//					return
//				}
//				err = stream.Send(response)
//				if err != nil {
//					fmt.Println("Failed to send response back", err)
//					return
//				}
//				fmt.Println("Sent response back for", in.RequestID, " successfully")
//			}()
//		}
//	}()
//
//	// Just keep it alive for now
//	time.Sleep(time.Hour)
//}
