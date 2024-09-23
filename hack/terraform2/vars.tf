# GKE module variables.
variable "cluster_name" {
  type        = string
  description = "GKE cluster name in GCP project."
}

variable "cluster_region" {
  type        = string
  default     = "europe-central2"
  description = "The region to create the cluster."
}

variable "project_id" {
  type        = string
  default     = "engineering-test-353509"
  description = "GCP project ID in which GKE cluster would be created."
}

variable "castai_api_url" {
  type        = string
  description = "URL of alternative CAST AI API to be used during development or testing"
}

# Variables required for connecting EKS cluster to CAST AI
variable "castai_api_token" {
  type        = string
  description = "CAST AI API token created in console.cast.ai API Access keys section."
}

variable "castai_grpc_url" {
  type        = string
  description = "CAST AI gRPC URL"
}

variable "tags" {
  type        = map(any)
  description = "Optional tags for new cluster nodes. This parameter applies only to new nodes - tags for old nodes are not reconciled."
  default     = {}
}
