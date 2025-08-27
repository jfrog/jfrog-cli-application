package model

// ReleaseAppVersionRequest represents a request to release an application version to production.
// This struct reuses CommonPromoteAppVersion for consistency with PromoteAppVersionRequest.
// A release is functionally promoted with a hardcoded stage ("prod") set by the backend,
// so the stage is not included here.
// This separation improves readability and intent in the codebase.
type ReleaseAppVersionRequest struct {
	CommonPromoteAppVersion
}

func NewReleaseAppVersionRequest(
	promotionType string,
	includedRepositoryKeys []string,
	excludedRepositoryKeys []string,
	artifactProperties []ArtifactProperty,
) *ReleaseAppVersionRequest {
	return &ReleaseAppVersionRequest{
		CommonPromoteAppVersion: CommonPromoteAppVersion{
			PromotionType:                promotionType,
			IncludedRepositoryKeys:       includedRepositoryKeys,
			ExcludedRepositoryKeys:       excludedRepositoryKeys,
			ArtifactAdditionalProperties: artifactProperties,
		},
	}
}
