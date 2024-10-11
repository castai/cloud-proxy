# CAST AI cloud-proxy

A component that is used to proxy calls from CAST AI to the Google Cloud Platform API. It can be used to ensure that all CAST AI interaction with GCP is done from the GKE cluster network perimeter.
The cloud-proxy also injects GCP credentials to the calls using [Workload Identity Federation for GKE](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity).

## Enabling workload identity

You can follow [this guide](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) configure your GKE cluster and IAM to support workload identity.

## Helm chart

Helm chart for the CAST AI cloud-proxy is available in the [castai/helm-charts repository](https://github.com/castai/helm-charts).

