# cloud-proxy


## Running the PoC

PoC has two modes - a "local-only" test and a "mocked CAST" test so far. They are controlled via command-line args, see `cmd/proxy/main.go`.

The local-only test is good to debug issues with the "replace-credentials" part or workload identity (e.g. does my identity have access?).

The "mocked CAST" test is useful to simulate the flow running real GRPC connection, albeit local. 

`make deploy` can be used to build, push an image and deploy the test proxy in the cluster kubectx points to currently. 
You will have to provide the REPO and VERSION variables as defaults are specific to one developer right now :) 

Easiest way to change modes is to edit dummy_deploy.yaml and redeploy. 

## Enabling workload identity

The [GCP guide](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) is OK. 

Essential steps are:
- Create a GKE cluster and enable workload identity in settings
- If using existing cluster, update the cluster to use workload identity AND node pools to have GKE metadata server (restarts the nodes)
- Create a SA in k8s (one is provided in dummy_deploy.yaml)
- Grant the SA IAM permissions (or test without them first to see that they are required)

To grant IAM permissions to the service account:

Use the following to get the current project number:
```bash
gcloud projects list \
    --filter="$(gcloud config get-value project)" \
    --format="value(PROJECT_NUMBER)"
```

```bash
PROJECT_ID=XXX PROJECT_NUMBER=YYYY gcloud projects add-iam-policy-binding projects/$PROJECT_ID \
    --role=roles/container.clusterViewer \
    --member=principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/default/sa/castai-cloud-proxy \
    --condition=None
```

**For some reason, above can fail with 404; no time to find out why now.**
Just replace the values in this command manually and it will work:

```bash
gcloud projects add-iam-policy-binding projects/<PROJECT_ID> \
    --role=roles/container.clusterViewer \
    --member=principal://iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/<PROJECT_NUMBER>.svc.id.goog/subject/ns/default/sa/castai-cloud-proxy \
    --condition=None
```

## Dev cloud-proxy deployment

You can use this [Terraform module](./hack/terraform/) to create a GKE cluster, onboard to CAST AI and install the castai-cloud-proxy Helm chart.

You might need to tweak the values in the [helm_release](./hack/terraform/cloud-proxy.tf) resource for the castai-cloud-proxy, to use the proper image and Helm chart. For now the Helm chart is not published, to you have to clone the [helm-charts repo](https://github.com/castai/helm-charts) and provide a local path in the `helm_release`.

To deploy the Terraform module execute:

```bash
export TF_VAR_castai_api_url=https://api-...localenv.cast.ai
export TF_VAR_castai_api_token=<your-token>
export TF_VAR_castai_grpc_url=grpc-...localenv.cast.ai:443
export TF_VAR_cluster_name=<cluster_name>

terraform -chdir=hack/terraform apply
```

The castai-cloud-proxy will try to connect the GRPC server provided in the variables.

