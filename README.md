# gcs-storage-operator
Create Storage Bucket

## Generate service account key file

Generate a Google Cloud service account key file (which is a JSON file containing the private key for a service account

1. Go to the Google Cloud Console and select the project you want to create the service account for.
1. In the left navigation menu, click on "IAM & Admin" and then select "Service accounts".
1. Click on "Create service account".
1. Enter a name and description for your service account, and click on "Create".
1. On the "Grant this service account access to project" page, select the roles you want to grant to the service account (e.g., "Editor", "Viewer", etc.) and click on "Continue".
1. Skip the "Grant users access to this service account" page by clicking on "Done".
1. On the "Service accounts" page, find the service account you just created and click on the three-dot menu on the right-hand side. Select "Create key".
1. In the "Create private key" dialog, select "JSON" as the key type and click on "Create".
Save the resulting JSON file to your local machine.

## Make service account key file available via Secret

Rename the file generate above to `service-account-key.json`

and create a Secret in your management cluster

```
kubectl create secret generic my-gcs-operator-secret --from-file=<path>/service-account-key.json -n sveltos-demo
```

## Create a Secret containing the Google ProjectID

```
kubectl create secret generic projectid --from-literal=projectid=<YOUR PROJECT ID> -n sveltos-demo
```

## Deploy operator

Assuming kubectl is pointing to your management cluster, 

```bash
kubectl apply -f https://raw.githubusercontent.com/gianlucam76/gcs-storage-operator/main/manifest/manifest.yaml
```

## Create a Bucket

```yaml
apiVersion: demo.projectsveltos.io/v1alpha1
kind: Bucket
metadata:
  name: bucket-sample
spec:
  bucketName: <BUCKET NAME>
  location: us-central1
  serviceAccounts:
    - serviceAccount:<SERVICE ACCOUNT EMAIL>
```

This will create a bucket. Deleting above instance, will cause controller to delete bucket from google cloud storage.
All service accounts listed will be granted roles/storage.objectViewer for the bucket.
