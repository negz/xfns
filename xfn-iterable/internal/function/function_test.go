package function

import (
	"os"
	"path"
	"testing"

	"github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"sigs.k8s.io/yaml"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		in      *v1alpha1.FunctionIO
		out     *v1alpha1.FunctionIO
		wantErr bool
	}{
		{
			name: "should create resources",
			in:   mustLoadFunction("create", "in.yaml"),
			out:  mustLoadFunction("create", "out.yaml"),
		},
		{
			name: "should delete managed",
			in:   mustLoadFunction("delete", "in.yaml"),
			out:  mustLoadFunction("delete", "out.yaml"),
		},
		{
			name: "should support empty iterable",
			in:   mustLoadFunction("empty-iterable", "in.yaml"),
			out:  mustLoadFunction("empty-iterable", "in.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.in); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			// First assert out ignoring raw resource
			if diff := cmp.Diff(tt.out, tt.in, cmpopts.IgnoreFields(v1alpha1.FunctionIO{}, "Results")); diff != "" {
				t.Errorf("a.Equal(b): -want, +got:\n%s", diff)
			}
		})
	}
}

func mustLoadFunction(testcase string, filename string) *v1alpha1.FunctionIO {
	file, err := os.ReadFile(path.Join("tests", testcase, filename))
	if err != nil {
		panic(err)
	}
	f := &v1alpha1.FunctionIO{}
	err = yaml.UnmarshalStrict(file, f)
	if err != nil {
		panic(err)
	}
	return f
}
