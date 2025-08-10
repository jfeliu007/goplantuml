package request

// GenerateRequest represents a request to generate PlantUML diagrams
type GenerateRequest struct {
	// Input directories
	Directories        []string `json:"directories"`
	IgnoredDirectories []string `json:"ignored_directories"`
	Recursive          bool     `json:"recursive"`

	// Output options
	OutputPath   string `json:"output_path"`
	OutputFormat string `json:"output_format"`

	// Rendering options
	ShowAggregations        bool `json:"show_aggregations"`
	HideFields              bool `json:"hide_fields"`
	HideMethods             bool `json:"hide_methods"`
	HideConnections         bool `json:"hide_connections"`
	ShowCompositions        bool `json:"show_compositions"`
	ShowImplementations     bool `json:"show_implementations"`
	ShowAliases             bool `json:"show_aliases"`
	ShowConnectionLabels    bool `json:"show_connection_labels"`
	AggregatePrivateMembers bool `json:"aggregate_private_members"`
	HidePrivateMembers      bool `json:"hide_private_members"`
	ShowOptionsAsNote       bool `json:"show_options_as_note"`

	// Diagram metadata
	Title string `json:"title"`
	Notes string `json:"notes"`

	// Custom resource patterns
	CustomResources []string `json:"custom_resources"`

	// Custom keyword patterns for function categorization
	CustomKeywords map[string][]string `json:"custom_keywords"`
}

// ValidateRequest represents a request to validate directories
type ValidateRequest struct {
	Directories []string `json:"directories"`
}

// AnalyzeRequest represents a request to analyze code structure
type AnalyzeRequest struct {
	Directories []string `json:"directories"`
	Recursive   bool     `json:"recursive"`
}

// RenderingOptionsRequest represents rendering options for PlantUML generation
type RenderingOptionsRequest struct {
	ShowAggregations        bool   `json:"show_aggregations"`
	HideFields              bool   `json:"hide_fields"`
	HideMethods             bool   `json:"hide_methods"`
	HideConnections         bool   `json:"hide_connections"`
	ShowCompositions        bool   `json:"show_compositions"`
	ShowImplementations     bool   `json:"show_implementations"`
	ShowAliases             bool   `json:"show_aliases"`
	ShowConnectionLabels    bool   `json:"show_connection_labels"`
	AggregatePrivateMembers bool   `json:"aggregate_private_members"`
	HidePrivateMembers      bool   `json:"hide_private_members"`
	ShowOptionsAsNote       bool   `json:"show_options_as_note"`
	Title                   string `json:"title"`
	Notes                   string `json:"notes"`
}

// DirectoryRequest represents a request for directory operations
type DirectoryRequest struct {
	Path      string   `json:"path"`
	Paths     []string `json:"paths"`
	Recursive bool     `json:"recursive"`
	Ignore    []string `json:"ignore"`
}
