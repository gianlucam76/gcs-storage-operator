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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	// BucketName is the name of the bucket to be created.
	BucketName string `json:"bucketName"`

	// Location is the GCS location where the bucket should be created.
	// The location must be a regional or multi-regional location.
	// For a list of available locations, see:
	// https://cloud.google.com/storage/docs/bucket-locations
	Location string `json:"location,omitempty"`
}

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	// BucketURL is the URL of the created bucket.
	BucketURL string `json:"bucketURL,omitempty"`

	// Status represent bucket status
	Status string `json:"string"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Bucket Name",type=string,JSONPath=`.spec.bucketName`
// +kubebuilder:printcolumn:name="Bucket URL",type=string,JSONPath=`.status.bucketURL`
// +kubebuilder:resource:scope=Namespaced

// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec,omitempty"`
	Status BucketStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
