package model

type CreateAppVersionRequest struct {
	ApplicationKey string                 `json:"application_key"`
	Version        string                 `json:"version"`
	Packages       []CreateVersionPackage `json:"packages"`
}

type CreateVersionPackage struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}
