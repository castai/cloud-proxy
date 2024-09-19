resource "google_project_iam_binding" "cloud_proxy_workload_identity" {
  project = var.project_id
  role    = "roles/container.clusterViewer"
  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[castai-agent/castai-cloud-proxy]"
  ]
}

resource "helm_release" "castai_cloud_proxy" {
  name             = "castai-cloud-proxy"
  repository       = "https://castai.github.io/helm-charts"
  chart            = "castai-cloud-proxy"
  namespace        = "castai-agent"
  create_namespace = true
  cleanup_on_fail  = true
  wait             = true

  set {
    name  = "castai.clusterID"
    value = module.castai-gke-cluster.cluster_id
  }

  set_sensitive {
    name  = "castai.apiKey"
    value = var.castai_api_token
  }

  set {
    name  = "castai.grpcURL"
    value = var.castai_grpc_url
  }

  depends_on = [module.castai-gke-cluster]
}
