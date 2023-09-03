package v1alpha1

type TemplateSource string

const (
	TemplateSourceDefault TemplateSource = ""
	// TemplateSourceFilesystem configures the template source to be a local directory.
	TemplateSourceFilesystem TemplateSource = "filesystem"

	// Note(turkenh): We may consider adding more sources like configmap, git, http, etc.
	// For now, we only support local directory.
)

type TemplateSourceFilesystemConfig struct {
	Path string `json:"path"`
}
type Template struct {
	Source                         TemplateSource `json:"source"`
	TemplateSourceFilesystemConfig `json:",inline"`
}

type ConfigSpec struct {
	Template Template `json:"template"`
}

type Config struct {
	APIVersion string     `json:"apiVersion"`
	Kind       string     `json:"kind"`
	Spec       ConfigSpec `json:"spec"`
}
