FROM gcr.io/distroless/static-debian11:nonroot

ARG TARGETARCH="amd64"

COPY bin/castai-cloud-proxy-$TARGETARCH /usr/local/bin/castai-cloud-proxy
CMD ["castai-cloud-proxy"]
