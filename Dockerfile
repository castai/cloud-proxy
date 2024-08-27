FROM gcr.io/distroless/static-debian11

#FROM ubuntu:latest
# TODO: Multi-arch build

COPY bin/castai-cloud-proxy-amd64  /usr/local/bin/castai-cloud-proxy
CMD ["castai-cloud-proxy"]
