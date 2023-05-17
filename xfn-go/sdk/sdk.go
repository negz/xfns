// Package function is an example stub that demonstrates how a Go Composition
// Function SDK might help make authoring functions easier.
package function

import (
	"fmt"
	"io"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/resource/unstructured/composite"
	iov1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
)

// ReadIO reads the function IO from the function container's stdin.
func ReadIO() (*iov1alpha1.FunctionIO, error) {
	y, err := io.ReadAll(os.Stdin)
	if err != nil {
		return &iov1alpha1.FunctionIO{}, errors.Wrap(err, "cannot read function IO from stdin")
	}
	fnio := &iov1alpha1.FunctionIO{}
	err = yaml.Unmarshal(y, fnio)
	return fnio, errors.Wrap(err, "cannot unmarshal function IO")
}

// WriteIO writes the function IO to the function container's stdout.
func WriteIO(fnio *iov1alpha1.FunctionIO) error {
	y, err := yaml.Marshal(fnio)
	if err != nil {
		return errors.Wrap(err, "cannot marshal function IO")
	}
	_, err = os.Stdout.Write(y)
	return errors.Wrap(err, "cannot write function IO to stdout")
}

// Fatal adds a fatal result to the function IO.
func Fatal(fnio *iov1alpha1.FunctionIO, err error) {
	fnio.Results = append(fnio.Results, iov1alpha1.Result{
		Severity: iov1alpha1.SeverityFatal,
		Message:  err.Error(),
	})
}

// Warning adds a warning result to the function IO.
func Warning(fnio *iov1alpha1.FunctionIO, err error) {
	fnio.Results = append(fnio.Results, iov1alpha1.Result{
		Severity: iov1alpha1.SeverityWarning,
		Message:  err.Error(),
	})
}

// Normal adds a normal result to the function IO.
func Normal(fnio *iov1alpha1.FunctionIO, message string) {
	fnio.Results = append(fnio.Results, iov1alpha1.Result{
		Severity: iov1alpha1.SeverityNormal,
		Message:  message,
	})
}

// Normalf adds a normal result to the function IO. It supports formatting.
func Normalf(fnio *iov1alpha1.FunctionIO, format string, a ...any) {
	fnio.Results = append(fnio.Results, iov1alpha1.Result{
		Severity: iov1alpha1.SeverityNormal,
		Message:  fmt.Sprintf(format, a...),
	})
}

// An Object represents an arbitrary Kubernetes object.
type Object interface {
	runtime.Object
	metav1.Object
}

// ResourceType within a FunctionIO.
type ResourceType string

// Resource types.
const (
	ResourceTypeDesired  ResourceType = "Desired"
	ResourceTypeObserved ResourceType = "Observed"
)

// GetOptions for getting resources from a FunctionIO.
type GetOptions struct {
	rtype ResourceType
}

// A GetOption configures how to get a resource from a FunctionIO.
type GetOption func(*GetOptions)

// Observed causes a Get function to return the observed resource, instead of
// the default behavior of returning the desired resource.
func Observed() GetOption {
	return func(o *GetOptions) {
		o.rtype = ResourceTypeObserved
	}
}

// GetCompositeResource returns the composite resource. It always returns the
// desired resource, not the observed one.
func GetCompositeResource(fnio *iov1alpha1.FunctionIO, into resource.Composite, o ...GetOption) error {
	return nil
}

// GetComposedResource returns the composed resource with the given name. It
// always returns the desired resource, not the observed one. Returns false if
// the named composed resource doesn't exist in the FunctionIO.
func GetComposedResource(fnio *iov1alpha1.FunctionIO, name string, into Object, o ...GetOption) (bool, error) {
	return false, nil
}

// SetCompositeResource updates the composite resource. Note that the entire
// Composite Resource is replaced, not just the fields that have changed.
func SetCompositeResource(fnio *iov1alpha1.FunctionIO, from resource.Composite) error {
	return nil
}

// SetComposedResource creates or updates the composed resource with the given
// name. Note that the entire Composed Resource is replaced, not just the fields
// that have changed.
func SetComposedResource(fnio *iov1alpha1.FunctionIO, name string, from Object) error {
	return nil
}

// NewEmptyCompositeResource returns a new, empty composite resource that can be
// populated by GetCompositeResource.
func NewEmptyCompositeResource() *CompositeResource {
	xr := composite.New()
	pv := fieldpath.Pave(xr.UnstructuredContent())
	return &CompositeResource{Unstructured: xr, Paved: pv}
}

// A CompositeResource can be used to make working with composite resources
// (XRs) easier. It has getters and setters for:
//
// - Common Kubernetes metadata (name, namespace, labels, annotations, etc)
// - Common Composite Resource fields (composition ref, claim ref, etc)
// - Arbitrary field paths (GetString, GetBool, etc)
type CompositeResource struct {
	*composite.Unstructured
	*fieldpath.Paved
}

func GetExternalName(o metav1.Object) string {
	return meta.GetExternalName(o)
}

func SetExternalName(o metav1.Object, name string) {
	meta.SetExternalName(o, name)
}
