generate-grpc:
	protoc --go_out=./internal/castai/proto --go-grpc_out=./internal/castai/proto ./internal/castai/proto/proxy.proto