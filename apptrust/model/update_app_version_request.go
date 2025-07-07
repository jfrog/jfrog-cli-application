package model

type UpdateAppVersionRequest struct {
	ApplicationKey   string              `json:"application_key"`
	Version          string              `json:"version"`
	Tag              string              `json:"tag,omitempty"`
	Properties       map[string][]string `json:"properties,omitempty"`
	DeleteProperties []string            `json:"delete_properties,omitempty"`
}
