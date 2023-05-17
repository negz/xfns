package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	iov1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	function "github.com/negz/xfns/xfn-go/sdk"

	"github.com/upbound/provider-dummy/apis/iam/v1alpha1"
)

var colors = []string{"red", "green", "blue", "yellow", "orange", "purple", "black", "white"}

// Function defines the logic of your function. It should mutate the supplied
// FunctionIO. Returning an error will automatically add a result to the
// FunctionIO.
func Function(ctx context.Context, fnio *iov1alpha1.FunctionIO) error {
	// Get the desired XR from the function IO.
	xr := function.NewEmptyCompositeResource()
	if err := function.GetCompositeResource(fnio, xr); err != nil {
		return errors.Wrap(err, "could not get composite resource")
	}

	// Get the desired resource count from the XR.
	count, err := xr.GetInteger("spec.count")
	if err != nil {
		return errors.Wrap(err, "could not get desired resource count")
	}

	// Ensure the desired number of robot resources exist.
	for i := 0; i < int(count); i++ {
		name := fmt.Sprintf("robot-%d", i)

		// Get the desired robot from the FunctionIO. If it doesn't exist we'll
		// create a new one.
		robot := &v1alpha1.Robot{}
		if _, err := function.GetComposedResource(fnio, name, robot); err != nil {
			return errors.Wrap(err, "could not get composed robot resource")
		}

		// The robot's external name should be derived from the XR's. If the
		// XR's external name is "example", the robots' external names will be
		// "example-robot-0", "example-robot-1", etc
		function.SetExternalName(robot, fmt.Sprintf("%s-%s", function.GetExternalName(xr), name))

		// Give this robot a random color!
		if robot.Spec.ForProvider.Color == "" {
			robot.Spec.ForProvider.Color = colors[rand.Intn(len(colors))]
		}

		// Set our new desired robot state. This will be a no-op if our robot
		// already existed.
		if err := function.SetComposedResource(fnio, name, robot); err != nil {
			return errors.Wrapf(err, "could not set composed robot resource %q", name)
		}

		function.Normalf(fnio, "successfully created robot %q", name)
	}

	return nil
}
