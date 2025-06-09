package model

type BindPackageRequest struct {
	ApplicationKey string   `json:"application_key"`
	Type           string   `json:"type"`
	Name           string   `json:"name"`
	Versions       []string `json:"versions"`
}
