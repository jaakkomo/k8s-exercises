# backup of todo-backend-postgres to Google Cloud Storage

## Deployment

Enable [Workload Identity Federation](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) on the cluster and node pool.

Create a bucket:

``` shell
BUCKET=<bucket name>
gcloud storage buckets create gs://$BUCKET
```

Grant bucket access privileges to the `todo-backup` ServiceAccount:

``` shell
PROJECT_NUMBER=<Google Cloud project number>
PROJECT_ID=<Google Cloud project id>

gcloud storage buckets add-iam-policy-binding gs://$BUCKET \
    --role=roles/storage.objectViewer \
    --member=principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/project/sa/todo-backup \
    --condition=None

gcloud storage buckets add-iam-policy-binding gs://$BUCKET \
    --role=roles/storage.objectCreator \
    --member=principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/project/sa/todo-backup \
    --condition=None
```

Deploy ServiceAccount and backup CronJob:

``` shell
kubectl apply -k .
```
