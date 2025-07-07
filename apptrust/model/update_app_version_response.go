package model

// UpdateAppVersionResponse represents the response from updating an application version.
type UpdateAppVersionResponse struct {
	ApplicationKey string              `json:"application_key"`
	Version        string              `json:"version"`
	Tag            *string             `json:"tag,omitempty"`
	Properties     map[string][]string `json:"properties"`
	ModifiedBy     string              `json:"modified_by"`
	ModifiedAt     string              `json:"modified_at"`
	Status         string              `json:"status"`
}
