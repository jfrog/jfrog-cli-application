package model

// UpdateAppVersionRequest represents a request to update application version annotations
// (tag and custom properties) for a specified application version.
type UpdateAppVersionRequest struct {
	Tag              string              `json:"tag,omitempty"`
	Properties       map[string][]string `json:"properties,omitempty"`
	DeleteProperties []string            `json:"delete_properties,omitempty"`
}
