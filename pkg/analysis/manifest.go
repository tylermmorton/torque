package analysis

// ProjectManifest is the top-level output of torque static analysis.
// It collects all components discovered during analysis and is designed
// to grow: Routes, Layouts, and other project-level entities will be
// added as top-level fields alongside Components.
type ProjectManifest struct {
	Components map[string]*Component `json:"components"`
}

// Component represents a single torque component discovered during static analysis.
type Component struct {
	TypeName    string         `json:"typeName"`
	PackageName string         `json:"packageName"`
	PackagePath string         `json:"packagePath"`
	SourceFile  string         `json:"sourceFile"`
	Interfaces  []string       `json:"interfaces"`
	Template    *string        `json:"template,omitempty"`
	StyleSheet  *string        `json:"stylesheet,omitempty"`
	Children    []ComponentRef `json:"children"`
}

// ComponentRef links a parent component to a child component via a struct field.
type ComponentRef struct {
	FieldName    string `json:"fieldName"`
	TemplateName string `json:"templateName"`
	TypeRef      string `json:"typeRef"`
}
