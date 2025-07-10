package model

type UpdateAppVersionRequest struct {
	Tag              string              `json:"tag,omitempty"`
	Properties       map[string][]string `json:"properties,omitempty"`
	DeleteProperties []string            `json:"delete_properties,omitempty"`
}
