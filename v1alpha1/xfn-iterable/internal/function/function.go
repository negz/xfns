package function

import (
	"crypto"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	fnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"

	"github.com/crossplane-contrib/xfn-iterable/pkg/apis/v1alpha1"
)

func Run(f *fnv1alpha1.FunctionIO) error {
	if rawFunction, err := json.Marshal(f); err != nil {
		f.Results = append(f.Results, fnv1alpha1.Result{
			Severity: fnv1alpha1.SeverityWarning,
			Message:  errors.Wrap(err, "cannot marshal input function").Error(),
		})
	} else {
		f.Results = append(f.Results, fnv1alpha1.Result{
			Severity: fnv1alpha1.SeverityNormal,
			Message:  string(rawFunction),
		})
	}

	// Parse the function configuration.
	cfg := &v1alpha1.Config{}
	if f.Config != nil {
		if err := yaml.UnmarshalStrict(f.Config.Raw, cfg); err != nil {
			return errors.Wrap(err, "cannot unmarshal as Config")
		}
	}

	xr := &unstructured.Unstructured{}
	if err := json.Unmarshal(f.Observed.Composite.Resource.Raw, xr); err != nil {
		return errors.Wrap(err, "cannot unmarshal observed composite resource")
	}
	pxr := fieldpath.Pave(xr.Object)

	iterable, err := GetArray(pxr, cfg.Spec.FromFieldPath)
	if err != nil {
		if !fieldpath.IsNotFound(err) || cfg.Spec.Policy.GetFromFieldPathPolicy() != v1.FromFieldPathPolicyOptional {
			return errors.Wrap(err, "cannot get value to iterate from observed composite resource")
		}
		// Treat not found as zero-length iterable if FromFieldPathPolicy is Optional
		iterable = []interface{}{}
	}

	for i, v := range iterable {
		for _, r := range cfg.Spec.Resources {
			// Unmarshal base manifest
			obj := &unstructured.Unstructured{}
			if err := json.Unmarshal(r.Base.Raw, obj); err != nil {
				return errors.Wrap(err, "cannot unmarshal manifest")
			}

			// Run manifest patches
			pr := fieldpath.Pave(obj.Object)
			if err := runPatches(pr, r.Patches, pxr, v, cfg.Spec.PatchSets); err != nil {
				return errors.Wrap(err, "while running patches")
			}

			// TODO: Support ConnectionDetails
			// TODO: Support ReadinessChecks
			res := fnv1alpha1.DesiredResource{
				Name: fmt.Sprintf("%s-%s", r.GetName(), hashName(i, obj)[:5]),
			}

			// Set name (deterministic)
			// Including obj in hash will recreate resources on all changes
			obj.SetName(fmt.Sprintf("%s-%s", xr.GetName(), hashName(r.GetName(), i, obj)[:5]))

			// Marshal manifest
			b, err := json.Marshal(obj)
			if err != nil {
				return errors.Wrap(err, "cannot marshal manifest")
			}

			res.Resource = runtime.RawExtension{
				Raw: b,
			}
			f.Desired.Resources = append(f.Desired.Resources, res)
		}
	}

	return nil
}

func runPatches(pr *fieldpath.Paved, patches []v1.Patch, pxr *fieldpath.Paved, iv interface{}, patchSets []v1.PatchSet) error {
	for _, p := range patches {
		var v interface{}

		switch p.GetType() {
		case PatchTypeFromIterableFieldPath:
			v = iv
			if p.FromFieldPath != nil {
				m, ok := iv.(map[string]any)
				if !ok {
					return errors.Errorf("%s: not an object", p.GetFromFieldPath())
				}
				pv, err := fieldpath.Pave(m).GetValue(p.GetFromFieldPath())
				if err != nil {
					return err
				}
				v = pv
			}
		case v1.PatchTypeFromCompositeFieldPath:
			pv, err := pxr.GetValue(p.GetFromFieldPath())
			if err != nil {
				return err
			}
			v = pv
		case v1.PatchTypePatchSet:
			for _, ps := range patchSets {
				if *p.PatchSetName == ps.Name {
					if err := runPatches(pr, ps.Patches, pxr, iv, patchSets); err != nil {
						return err
					}
				}
			}
			// Continue loop; no values set by recursive call
			continue
		default:
			return fmt.Errorf("%s: unsupported patch type", p.Type)
		}

		v, err := transformValue(v, p.Transforms)
		if err != nil {
			return err
		}

		if err := pr.SetValue(p.GetToFieldPath(), v); err != nil {
			return err
		}
	}
	return nil
}

func transformValue(v interface{}, transforms []v1.Transform) (interface{}, error) {
	for _, t := range transforms {
		switch t.Type {
		case v1.TransformTypeString:
			tv, err := transformString(v, *t.String)
			if err != nil {
				return nil, err
			}
			v = tv
		default:
			return nil, fmt.Errorf("%s: unsupported transform type", t.Type)
		}
	}

	return v, nil
}

func transformString(v interface{}, transform v1.StringTransform) (interface{}, error) {
	switch transform.Type {
	case "", v1.StringTransformTypeFormat:
		return fmt.Sprintf(*transform.Format, v), nil
	default:
		return nil, fmt.Errorf("%s: unsupported string transform type", transform.Type)
	}
}

func hashName(objs ...interface{}) string {
	digester := crypto.MD5.New()
	for _, ob := range objs {
		_, _ = fmt.Fprint(digester, ob)
	}
	return fmt.Sprintf("%x", digester.Sum(nil))
}
