package model

type BindPackageRequest struct {
	Type    string `json:"package_type"`
	Name    string `json:"package_name"`
	Version string `json:"package_version"`
}
