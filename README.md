# cloud-proxy


To grant IAM permissions to the service account

```bash
PROJECT_ID=XXX PROJECT_NUMBER=YYYY gcloud projects add-iam-policy-binding projects/$PROJECT_ID \
    --role=roles/container.clusterViewer \
    --member=principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/default/sa/castai-cloud-proxy \
    --condition=None
```

Use the following to get the current project number:
```bash
gcloud projects list \
    --filter="$(gcloud config get-value project)" \
    --format="value(PROJECT_NUMBER)"
```



PROJECT_ID=engineering-test-353509 PROJECT_NUMBER=1092480157775 gcloud projects add-iam-policy-binding projects/$PROJECT_ID \
    --role=roles/container.clusterViewer \
    --member=principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/default/sa/castai-cloud-proxy \
    --condition=None



gcloud projects add-iam-policy-binding projects/engineering-test-353509 \
    --role=roles/container.clusterViewer \
    --member=principal://iam.googleapis.com/projects/1092480157775/locations/global/workloadIdentityPools/engineering-test-353509.svc.id.goog/subject/ns/default/sa/castai-cloud-proxy \
    --condition=None

principal://iam.googleapis.com/projects/1092480157775/locations/global/workloadIdentityPools/engineering-test-353509.svc.id.goog/subject/ns/default/sa/castai-cloud-proxy