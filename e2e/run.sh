#!/bin/bash

set -eEu

CLUSTER_NAME=cloud-proxy-e2e
IMAGE_TAG=$RANDOM

kind::ensure() {
	if ! kind export kubeconfig --name "${CLUSTER_NAME}"; then
		kind create cluster --name "${CLUSTER_NAME}"
	fi
}

kind::load_images() {
	kind load docker-image --name "${CLUSTER_NAME}" cloud-proxy:$IMAGE_TAG
	kind load docker-image --name "${CLUSTER_NAME}" cloud-proxy-e2e:$IMAGE_TAG

}

cloud_proxy::build_image() {
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/castai-cloud-proxy-amd64 ./cmd/proxy
	docker build -t cloud-proxy:$IMAGE_TAG .
}

cloud_proxy::helm_install() {
	helm upgrade --install --wait \
		--set image.repository=cloud-proxy \
		--set image.tag=$IMAGE_TAG \
		--set config.grpc.endpoint=cloud-proxy-e2e:50051 \
		--set config.grpc.key=test \
		--set-file config.gcpCredentials="${GCP_CREDENTIALS}" \
		cast-cloud-proxy ./charts/cast-cloud-proxy
}

e2e::build_image() {
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/cloud-proxy-e2e ./e2e
	docker build -t cloud-proxy-e2e:$IMAGE_TAG -f ./e2e/Dockerfile .
}

e2e::helm_install() {
	helm upgrade --install --wait \
		--set image.tag=$IMAGE_TAG \
		cloud-proxy-e2e ./e2e/chart
}

e2e::helm_uninstall() {
	helm delete cloud-proxy-e2e
}

main() {
	[[ -z "${GCP_CREDENTIALS:-}" ]] && echo "Missing GCP_CREDENTIALS" && exit 1

	kind::ensure
	cloud_proxy::build_image
	e2e::build_image
	kind::load_images
	cloud_proxy::helm_install
	e2e::helm_install

	#e2e::helm_uninstall
}

main $@
