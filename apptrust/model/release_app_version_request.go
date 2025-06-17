package model

// ReleaseAppVersionRequest represents a request to release an app version to the release stage
type ReleaseAppVersionRequest struct {
	PromotionType                string            `json:"promotion_type,omitempty"`
	IncludedRepositoryKeys       []string          `json:"included_repository_keys,omitempty"`
	ExcludedRepositoryKeys       []string          `json:"excluded_repository_keys,omitempty"`
	ArtifactAdditionalProperties map[string]string `json:"artifact_additional_properties,omitempty"`
}
