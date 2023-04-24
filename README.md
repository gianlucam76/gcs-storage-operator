# gcs-storage-operator
Create Storage Bucket

## Generate service account key file

Generate a Google Cloud service account key file (which is a JSON file containing the private key for a service account). You can skip this if you already have one.

1. authenticate with your Google Cloud account by running the following command in your terminal
```
gcloud auth login
```
2. Once you're authenticated, you can use the following command to create a new service account and generate a key for it
```
gcloud iam service-accounts keys create service-account-key.json --iam-account=[SA-EMAIL]
```
replace [SA-EMAIL] with the email address of the service account you want to create the key for.

## Make service account key file available via Secret

Create a Secret in your management cluster, using the service-account-key.json created above.

```
kubectl create ns sveltos-demo

kubectl create secret generic my-gcs-operator-secret --from-file=service-account-key.json -n sveltos-demo
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
  serviceAccount: serviceAccount:<SERVICE ACCOUNT EMAIL>
```

This will create a bucket. Deleting above instance, will cause controller to delete bucket from google cloud storage.
Service account listed will be granted roles/storage.objectViewer for the bucket.
