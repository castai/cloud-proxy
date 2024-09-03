resource "google_project_iam_binding" "cloud_proxy_workload_identity" {
  project = var.project_id
  role    = "roles/container.clusterViewer"
  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[castai-agent/castai-cloud-proxy]"
  ]
}

resource "helm_release" "castai_cloud_proxy" {
  name = "castai-cloud-proxy"
  //repository       = "https://castai.github.io/helm-charts"
  //chart            = "castai-cloud-proxy"
  chart            = "../../../github-helm-charts/charts/castai-cloud-proxy"
  namespace        = "castai-agent"
  create_namespace = true
  cleanup_on_fail  = true
  wait             = true

  set {
    name  = "image.tag"
    value = "latest"
  }

  set {
    name  = "image.pullPolicy"
    value = "Always"
  }

  set {
    name  = "config.grpc.endpoint"
    value = var.castai_grpc_url
  }

  set_sensitive {
    name  = "config.grpc.key"
    value = var.castai_api_token
  }

  depends_on = [module.castai-gke-cluster]
}
