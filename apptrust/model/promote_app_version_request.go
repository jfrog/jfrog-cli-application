package model

type PromoteAppVersionRequest struct {
	ApplicationKey string `json:"application_key"`
	Version        string `json:"version"`
	Environment    string `json:"environment"`
}
