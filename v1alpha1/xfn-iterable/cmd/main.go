package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	fnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/crossplane-contrib/xfn-iterable/internal/function"
)

//goland:noinspection ALL
func main() {
	// Read the function IO from stdin
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		failFatal(&fnv1alpha1.FunctionIO{}, errors.Wrap(err, "cannot read stdin"))
		return
	}

	// Unmarshal the function IO
	f := &fnv1alpha1.FunctionIO{}
	if err = yaml.Unmarshal(stdin, f); err != nil {
		failFatal(&fnv1alpha1.FunctionIO{}, errors.Wrap(err, "cannot unmarshal as FunctionIO"))
		return
	}

	// TODO: Validate function config for required fields

	// Run the function
	if err := function.Run(f); err != nil {
		failFatal(&fnv1alpha1.FunctionIO{}, errors.Wrap(err, "failed while running function"))
		return
	}

	// Marshal and write the output
	result, err := yaml.Marshal(f)
	if err != nil {
		failFatal(&fnv1alpha1.FunctionIO{}, errors.Wrap(err, "cannot marshal output"))
		return
	}
	fmt.Print(string(result))
}

func failFatal(io *fnv1alpha1.FunctionIO, err error) {
	io.Results = append(io.Results, fnv1alpha1.Result{
		Severity: fnv1alpha1.SeverityFatal,
		Message:  err.Error(),
	})
	b, err := yaml.Marshal(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to marshal resulting FunctionIO: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}
