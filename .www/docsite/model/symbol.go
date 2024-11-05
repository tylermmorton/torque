package model

type Symbol struct {
	Package    string   `json:"package"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	IsExported bool     `json:"is_exported"`
	Source     string   `json:"source"`
	Comments   string   `json:"comments"`
	Parameters []string `json:"parameters,omitempty"`
	Returns    []string `json:"returns,omitempty"`
	Receiver   *string  `json:"receiver,omitempty"`

	FileName     string `json:"file_name"`
	LineNumber   uint   `json:"line_number"`   // Line number of the declaration
	LinePosition uint   `json:"line_position"` // Column position in the line
}
