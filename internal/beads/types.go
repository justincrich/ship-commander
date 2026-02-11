package beads

// Dependency describes a relationship returned by Beads dependency queries.
type Dependency struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	Priority       int    `json:"priority"`
	IssueType      string `json:"issue_type"`
	Owner          string `json:"owner"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      string `json:"created_by"`
	UpdatedAt      string `json:"updated_at"`
	DependencyType string `json:"dependency_type,omitempty"`
}

// Bead matches the common fields from `bd --json` issue responses.
type Bead struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       string       `json:"status"`
	Priority     int          `json:"priority"`
	IssueType    string       `json:"issue_type"`
	Owner        string       `json:"owner"`
	CreatedAt    string       `json:"created_at"`
	CreatedBy    string       `json:"created_by"`
	UpdatedAt    string       `json:"updated_at"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
	Parent       string       `json:"parent,omitempty"`
}

// CreateOpts controls issue creation via `bd create`.
type CreateOpts struct {
	Title       string
	Type        string
	Description string
	Parent      *string
	Labels      []string
	Priority    string
}

// ListOpts controls issue listing filters via `bd list`.
type ListOpts struct {
	Type   string
	Status string
	Limit  int
	Parent string
	Labels []string
}
