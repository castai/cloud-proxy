VERSION ?= latest
REPO ?= us-docker.pkg.dev/castai-hub/library/cloud-proxy

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/castai-cloud-proxy-amd64 ./cmd/proxy
	#docker build -t us-docker.pkg.dev/castai-hub/library/svc:$(VERSION) .
	docker build -t $(REPO):$(VERSION) --platform linux/amd64 .
.PHONY: build

push:
	docker push $(REPO):$(VERSION)
.PHONY: push

release: build push
.PHONY: release

deploy: build push
	# Get the latest digest because it doesn't work for some f. reason and put it in the yaml
	@DIGEST=$$(docker inspect --format='{{index .RepoDigests 0}}' $(REPO):$(VERSION) | awk -F@ '{print $$2}'); \
	sed "s/{{IMAGE_DIGEST}}/$${DIGEST}/g" dummy_deploy.yaml > tmp.yaml
	kubectl apply -f tmp.yaml
	rm tmp.yaml
.PHONY: deploy

generate-grpc:
	mkdir -p proto/gen
	protoc proto/v1alpha/proxy.proto \
			--go_out=proto/gen --go_opt paths=source_relative  \
			--go-grpc_out=proto/gen --go-grpc_opt paths=source_relative
.PHONY: generate-grpc

lint:
	golangci-lint run ./...
.PHONY: lint

test:
	go test ./... -race
.PHONY: test