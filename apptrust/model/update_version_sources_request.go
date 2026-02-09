package model

type UpdateVersionSourcesRequest struct {
	AddSources *CreateVersionSources `json:"add_sources,omitempty"`
	Filters    *CreateVersionFilters `json:"filters,omitempty"`
}
