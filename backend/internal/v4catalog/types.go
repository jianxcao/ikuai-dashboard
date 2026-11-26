package v4catalog

// OptionItem describes a select choice in a resource form.
type OptionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// FieldSpec describes an editable field in create/update forms.
type FieldSpec struct {
	Key            string       `json:"key"`
	Label          string       `json:"label"`
	Type           string       `json:"type"` // text | number | select | switch | textarea | csv | password
	InterfaceScope string       `json:"interface_scope,omitempty"`
	Required       bool         `json:"required,omitempty"`
	Default        any          `json:"default,omitempty"`
	Placeholder    string       `json:"placeholder,omitempty"`
	Options        []OptionItem `json:"options,omitempty"`
	Help           string       `json:"help,omitempty"`
}

// ColumnSpec describes a display column in resource tables.
type ColumnSpec struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Kind  string `json:"kind,omitempty"` // text | mono | badge | toggle | time | bytes | speed | compound
}

// ActionSpec describes a custom one-click action on a resource.
type ActionSpec struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Dangerous   bool   `json:"dangerous,omitempty"`
	ConfirmText string `json:"confirm_text,omitempty"`
}

// ResourceSpec defines one iKuai V4 REST resource aligned with ikuai-cli.
type ResourceSpec struct {
	Key            string         `json:"key"`
	Group          string         `json:"group"`
	Label          string         `json:"label"`
	Kind           string         `json:"kind"` // collection | singleton
	V4Path         string         `json:"v4_path"`
	CLICommand     string         `json:"cli_command"`
	Description    string         `json:"description"`
	Operations     []string       `json:"operations"` // list, get, create, update, toggle, delete
	Columns        []ColumnSpec   `json:"columns"`
	Fields         []FieldSpec    `json:"fields"`
	Actions        []ActionSpec   `json:"actions,omitempty"`
	SearchKeys     string         `json:"-"`
	CompoundFields []string       `json:"-"`
	ObjectValueKey string         `json:"object_value_key,omitempty"`
	DeleteMode     string         `json:"delete_mode,omitempty"` // id (default) | clear | srcfile
	Defaults       map[string]any `json:"-"`
	UpdateFields   []string       `json:"-"`
	SupportsPaging bool           `json:"supports_paging"`
}

// ListResult represents normalized collection or singleton response data.
type ListResult struct {
	Resource ResourceSpec     `json:"resource"`
	Rows     []map[string]any `json:"rows"`
	Total    int              `json:"total"`
	Config   map[string]any   `json:"config,omitempty"`
}
