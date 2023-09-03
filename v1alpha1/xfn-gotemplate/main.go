package main

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	fnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/pkg/errors"
	"github.com/upbound/xfn-gotemplate/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const (
	defaultTemplatesDir = "/templates"
)

type Datasource struct {
	Composite fnv1alpha1.ObservedComposite `json:"composite"`

	Resources map[string]fnv1alpha1.ObservedResource `json:"resources,omitempty"`
}

func main() {
	// Read the function IO from stdin.
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		failFatal(fnv1alpha1.FunctionIO{}, errors.Wrap(err, "cannot read stdin"))
		return
	}

	// Unmarshal the function IO.
	in := fnv1alpha1.FunctionIO{}
	if err = yaml.Unmarshal([]byte(strings.TrimSpace(string(stdin))), &in); err != nil {
		failFatal(fnv1alpha1.FunctionIO{}, errors.Wrap(err, "cannot unmarshal as FunctionIO"))
		return
	}

	out := *(in.DeepCopy())

	// Parse the function configuration.
	cfg := v1alpha1.Config{}
	if in.Config != nil {
		if err = yaml.Unmarshal(in.Config.Raw, &cfg); err != nil {
			failFatal(out, errors.Wrap(err, "cannot unmarshal as Config"))
			return
		}
	}

	// Get the templates source, we default to the filesystem if not specified which is the only supported one right now.
	switch cfg.Spec.Template.Source {
	case v1alpha1.TemplateSourceDefault, v1alpha1.TemplateSourceFilesystem:
		break
	default:
		failFatal(out, errors.Errorf("unsupported template source: %s", cfg.Spec.Template.Source))
	}

	// Get the templates directory, we default to /templates if not specified.
	templatesDir := defaultTemplatesDir
	if cfg.Spec.Template.Path != "" {
		templatesDir = cfg.Spec.Template.Path
	}

	// Read the templates.
	templates, err := readTemplates(templatesDir)
	if err != nil {
		failFatal(out, errors.Wrap(err, "cannot read templates"))
		return
	}

	// Parse the templates.
	tpl, err := template.New("manifests").Funcs(sprig.FuncMap()).Parse(templates)
	if err != nil {
		failFatal(out, errors.Wrap(err, "cannot parse template"))
		return
	}

	// Build the datasource.
	// We simply pass everything under "observed" of the input as the datasource (i.e. similar to values.yaml in helm).
	// For convenience, we convert the observed resources to a map keyed by their name so that they can be referenced
	// in the templates by their name.
	ds := Datasource{}
	ds.Composite = in.Observed.Composite
	ds.Resources = make(map[string]fnv1alpha1.ObservedResource, len(in.Observed.Resources))
	for _, r := range in.Observed.Resources {
		ds.Resources[r.Name] = r
	}

	// Convert the datasource to a map to pass it to rendering.
	b, err := yaml.Marshal(ds)
	if err != nil {
		failFatal(out, errors.Wrap(err, "cannot marshal observed resources"))
		return
	}

	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(b, &vals); err != nil {
		failFatal(out, errors.Wrap(err, "cannot unmarshal observed resources"))
		return
	}

	// Render the templates.
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, vals); err != nil {
		failFatal(out, errors.Wrap(err, "cannot execute template"))
		return
	}

	// Parse the rendered manifests.
	var objs []*unstructured.Unstructured
	decoder := kyaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(buf.String()), 1024)
	for {
		u := &unstructured.Unstructured{}
		if err := decoder.Decode(&u); err != nil {
			if err == io.EOF {
				break
			}
			failFatal(out, errors.Wrap(err, "cannot decode manifest"))
			return
		}
		if u != nil {
			objs = append(objs, u)
		}
	}

	// Convert the rendered manifests to a list of desired resources.
	desiredResources := make(map[string]fnv1alpha1.DesiredResource, len(objs))
	for _, obj := range objs {
		res := desiredResources[obj.GetName()]
		paved := fieldpath.Pave(obj.Object)
		// We have two meta types to make connection details and readiness checks configurable in the templates.
		if obj.GetAPIVersion() == "meta.gotemplate.xfn.upbound.io/v1alpha1" {
			switch obj.GetKind() {
			case "ConnectionDetails":
				if err = paved.GetValueInto("connectionDetails", &res.ConnectionDetails); err != nil {
					failFatal(out, errors.Wrap(err, "cannot get connection details"))
					return
				}
			case "ReadinessChecks":
				if err = paved.GetValueInto("readinessChecks", &res.ReadinessChecks); err != nil {
					failFatal(out, errors.Wrap(err, "cannot get readiness checks"))
					return
				}
			default:
				failFatal(out, fmt.Errorf("unknown kind %s", obj.GetKind()))
				return
			}
			desiredResources[obj.GetName()] = res
			continue
		}
		b, err = obj.MarshalJSON()
		if err != nil {
			failFatal(out, errors.Wrap(err, "cannot marshal manifest"))
			return
		}
		res.Resource = runtime.RawExtension{
			Raw:    b,
			Object: obj,
		}
		desiredResources[obj.GetAnnotations()["gotemplate.xfn.upbound.io/resource-name"]] = res
	}

	// Merge the desired resources with the existing ones.
	desired := in.Desired
	for name, res := range desiredResources {
		found := false
		for i, r := range desired.Resources {
			if r.Name == name {
				desired.Resources[i].Resource = res.Resource
				desired.Resources[i].ConnectionDetails = res.ConnectionDetails
				desired.Resources[i].ReadinessChecks = res.ReadinessChecks

				// This is an already existing resource, we just update it without appending again.
				found = true
				break
			}
		}
		if found {
			continue
		}
		desired.Resources = append(desired.Resources, fnv1alpha1.DesiredResource{
			Name:              name,
			Resource:          res.Resource,
			ConnectionDetails: res.ConnectionDetails,
			ReadinessChecks:   res.ReadinessChecks,
		})
	}

	// Marshal and write the output.
	out.Desired = desired
	b, err = yaml.Marshal(out)
	if err != nil {
		failFatal(out, errors.Wrap(err, "cannot marshal output"))
		return
	}

	fmt.Println(string(b))
}

func failFatal(io fnv1alpha1.FunctionIO, err error) {
	io.Results = append(io.Results, fnv1alpha1.Result{
		Severity: fnv1alpha1.SeverityFatal,
		Message:  err.Error(),
	})
	b, _ := yaml.Marshal(io)
	fmt.Println(string(b))
}

func readTemplates(templatesDir string) (string, error) {
	templates := ""
	if err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		templates += string(data)
		templates += "\n---\n"
		return nil
	}); err != nil {
		return "", err
	}
	return templates, nil
}
