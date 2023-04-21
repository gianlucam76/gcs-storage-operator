/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"os"

	"cloud.google.com/go/storage"
	"github.com/go-logr/logr"
	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	storagev1alpha1 "github.com/gianlucam76/gcs-storage-operator/api/v1alpha1"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	storageClient *storage.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	Recorder      record.EventRecorder
}

const (
	finalizerName = "demo.projectsveltos.io/storage-finalizer"
)

//+kubebuilder:rbac:groups=demo.projectsveltos.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.projectsveltos.io,resources=buckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.projectsveltos.io,resources=buckets/finalizers,verbs=update;patch

func (r *BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Bucket", req.NamespacedName)

	// Set the GoogleAPIHTTPDebug environment variable to true
	os.Setenv("GoogleAPIHTTPDebug", "true")

	log.Info("reconciling")

	// Fetch the Bucket instance
	gcsBucket := &storagev1alpha1.Bucket{}
	err := r.Get(ctx, req.NamespacedName, gcsBucket)
	if err != nil {
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Create a new Google Cloud Storage client
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile("/var/secrets/google/service-account-key.json"))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Set the storageClient field of the ReconcileGCSBucket object
	r.storageClient = storageClient

	// Check if the Bucket instance is being deleted
	if !gcsBucket.ObjectMeta.DeletionTimestamp.IsZero() {
		// If the bucket is being deleted, delete it from Google Cloud Storage
		if err := r.deleteBucket(gcsBucket); err != nil {
			log.Error(err, "failed to delete")
			return ctrl.Result{}, err
		}
		// Bucket deleted successfully - remove the finalizer
		controllerutil.RemoveFinalizer(gcsBucket, finalizerName)
		if err := r.Update(ctx, gcsBucket); err != nil {
			log.Error(err, "failed to update")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Check if the Bucket instance has the finalizer
	if !controllerutil.ContainsFinalizer(gcsBucket, finalizerName) {
		// If not, add it
		controllerutil.AddFinalizer(gcsBucket, finalizerName)
		if err := r.Update(ctx, gcsBucket); err != nil {
			log.Error(err, "failed to update")
			return ctrl.Result{}, err
		}
	}

	log.Info("creating bucket")
	// Create the Google Cloud Storage bucket
	if err := r.createBucket(gcsBucket, log); err != nil {
		log.Error(err, "failed to create")
		return ctrl.Result{}, err
	}

	log.Info("granting bucket")
	// Grant access to the Google Cloud Storage bucket
	if err := r.grantBucketAccess(gcsBucket, log); err != nil {
		log.Error(err, "failed to grant access")
		return ctrl.Result{}, err
	}

	// Set the status of the Bucket instance to "Created"
	gcsBucket.Status = storagev1alpha1.BucketStatus{BucketURL: "https://storage.googleapis.com/" + gcsBucket.Spec.BucketName}
	if err := r.Status().Update(ctx, gcsBucket); err != nil {
		log.Error(err, "failed to update")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BucketReconciler) deleteBucket(gcsBucket *storagev1alpha1.Bucket) error {
	ctx := context.Background()
	bucket := r.storageClient.Bucket(gcsBucket.Spec.BucketName)

	// Check if the bucket exists before deleting
	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			// If the bucket doesn't exist, ignore the error
			return nil
		}
		return err
	}

	// Delete the bucket
	if err := bucket.Delete(ctx); err != nil {
		return err
	}
	return nil
}

func (r *BucketReconciler) bucketExists(bucketName string, log logr.Logger) (bool, error) {
	// Create a new Storage client
	ctx := context.Background()

	// Get the bucket handle
	bucket := r.storageClient.Bucket(bucketName)

	// Check if the bucket exists
	_, err := bucket.Attrs(ctx)
	if err == storage.ErrBucketNotExist {
		log.Error(err, "failed to verify if bucket exists")
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (r *BucketReconciler) createBucket(gcsBucket *storagev1alpha1.Bucket, log logr.Logger) error {
	ctx := context.Background()
	bucket := r.storageClient.Bucket(gcsBucket.Spec.BucketName)

	exist, err := r.bucketExists(gcsBucket.Spec.BucketName, log)
	if err != nil {
		log.Error(err, "failed to verify if bucket exists")
		return err
	}

	if exist {
		log.Info("bucket already exists")
		return nil
	}

	// Create the bucket
	if err := bucket.Create(ctx, os.Getenv("PROJECT_ID"), nil); err != nil {
		return err
	}
	return nil
}

func (r *BucketReconciler) grantBucketAccess(gcsBucket *storagev1alpha1.Bucket, log logr.Logger) error {
	ctx := context.Background()

	// Get the bucket handle
	bucket := r.storageClient.Bucket(gcsBucket.Spec.BucketName)
	iamHandle := bucket.IAM()

	policy, err := iamHandle.Policy(ctx)
	if err != nil {
		log.Error(err, "failed to grant access")
		return err
	}

	for i := range gcsBucket.Spec.ServiceAccounts {
		policy.Add(gcsBucket.Spec.ServiceAccounts[i], "roles/storage.objectViewer")
	}
	err = iamHandle.SetPolicy(ctx, policy)
	if err != nil {
		log.Error(err, "failed to grant access")
		return err
	}

	log.Info("granted access")
	return nil
}

func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.Bucket{}).
		Complete(r)
}
