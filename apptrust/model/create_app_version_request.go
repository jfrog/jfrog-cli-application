package model

type CreateAppVersionRequest struct {
	ApplicationKey string                `json:"application_key"`
	Version        string                `json:"version"`
	Sources        *CreateVersionSources `json:"sources,omitempty"`
	Tag            string                `json:"tag,omitempty"`
}

type CreateVersionSources struct {
	Packages       []CreateVersionPackage       `json:"packages,omitempty"`
	Builds         []CreateVersionBuild         `json:"builds,omitempty"`
	ReleaseBundles []CreateVersionReleaseBundle `json:"release_bundles,omitempty"`
	Versions       []CreateVersionReference     `json:"versions,omitempty"`
	Exclude        []ExcludePackage             `json:"exclude,omitempty"`
}

type CreateVersionPackage struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

type CreateVersionArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type CreateVersionBuild struct {
	RepositoryKey       string `json:"repository_key,omitempty"`
	Name                string `json:"name"`
	Number              string `json:"number"`
	Started             string `json:"started,omitempty"`
	IncludeDependencies bool   `json:"include_dependencies,omitempty"`
}

type CreateVersionReleaseBundle struct {
	ProjectKey    string `json:"project_key"`
	RepositoryKey string `json:"repository_key"`
	Name          string `json:"name"`
	Version       string `json:"version"`
}

type CreateVersionReference struct {
	ApplicationKey string `json:"application_key,omitempty"`
	Version        string `json:"version"`
}

type ExcludePackage struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
