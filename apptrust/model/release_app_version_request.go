package model

const (
	ReleaseStage = "PROD"
)

type ReleaseAppVersionRequest struct {
	PromoteAppVersionRequest
}

func NewReleaseAppVersionRequest(
	promotionType string,
	includedRepositoryKeys []string,
	excludedRepositoryKeys []string,
	artifactProperties map[string]string,
) *ReleaseAppVersionRequest {
	return &ReleaseAppVersionRequest{
		PromoteAppVersionRequest: PromoteAppVersionRequest{
			Stage:                        ReleaseStage,
			PromotionType:                promotionType,
			IncludedRepositoryKeys:       includedRepositoryKeys,
			ExcludedRepositoryKeys:       excludedRepositoryKeys,
			ArtifactAdditionalProperties: artifactProperties,
		},
	}
}
