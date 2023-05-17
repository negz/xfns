package main

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	iov1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

func TestFunction(t *testing.T) {
	type args struct {
		ctx  context.Context
		fnio *iov1alpha1.FunctionIO
	}
	type want struct {
		fnio *iov1alpha1.FunctionIO
		err  error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		// TODO(negz): Add tests. :)
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := Function(tc.args.ctx, tc.args.fnio)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nFunction(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.fnio, tc.args.fnio); diff != "" {
				t.Errorf("\nFunction(...): -want, +got:\n%s", diff)
			}
		})
	}

}
