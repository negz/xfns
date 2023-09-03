package function

import (
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/pkg/errors"
)

// GetArray value of the supplied field path.
func GetArray(p *fieldpath.Paved, path string) ([]interface{}, error) {
	v, err := p.GetValue(path)
	if err != nil {
		return nil, err
	}
	a, ok := v.([]any)
	if !ok {
		return nil, errors.Errorf("%s: not an array", path)
	}
	return a, nil
}
