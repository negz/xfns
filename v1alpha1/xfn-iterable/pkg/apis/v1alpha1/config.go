package v1alpha1

import (
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigSpec struct {
	// FromFieldPath is the path of the field on the resource whose value is
	// to be used as input to the iteration.
	FromFieldPath string `json:"fromFieldPath"`

	// Policy configures the specifics of function behaviour.
	// +optional
	Policy *v1.PatchPolicy `json:"policy,omitempty"`

	// PatchSets define a named set of patches that may be included by
	// any resource in this Composition.
	// PatchSets cannot themselves refer to other PatchSets.
	// +optional
	PatchSets []v1.PatchSet `json:"patchSets,omitempty"`

	// Resources is a list of resource templates that will be used when a
	// composite resource referring to this composition is created. At least one
	// of resources and functions must be specififed. If both are specified the
	// resources will be rendered first, then passed to the functions for
	// further processing.
	// +optional
	Resources []v1.ComposedTemplate `json:"resources,omitempty"`
}

type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigSpec `json:"spec"`
}
