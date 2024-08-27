VERSION ?= poc2

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/castai-cloud-proxy-amd64 ./cmd/proxy
	#docker build -t us-docker.pkg.dev/castai-hub/library/svc:$(VERSION) .
	docker build -t lachezarcast/cloud-proxy:$(VERSION) --platform linux/amd64 .

push:
	docker push lachezarcast/cloud-proxy:$(VERSION)

release: build push

deploy: build push
	# Get the latest digest because it doesn't work for some f. reason and put it in the yaml
	@DIGEST=$$(docker inspect --format='{{index .RepoDigests 0}}' lachezarcast/cloud-proxy:$(VERSION) | awk -F@ '{print $$2}'); \
	sed "s/{{IMAGE_DIGEST}}/$${DIGEST}/g" dummy_deploy.yaml > tmp.yaml
	kubectl apply -f tmp.yaml
	rm tmp.yaml


generate-grpc:
	protoc --go_out=./internal/castai/proto --go-grpc_out=./internal/castai/proto ./internal/castai/proto/proxy.proto
