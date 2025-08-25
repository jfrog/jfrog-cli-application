package model

const (
	PromotionTypeCopy = "copy"
	PromotionTypeMove = "move"

	// This value cannot be set via the --promotion-type flag in the CLI.
	// It is sent to the promotion_type field in the REST API only when the --dry-run flag is used.
	PromotionTypeDryRun = "dry_run"
)

var PromotionTypeValues = []string{
	PromotionTypeCopy,
	PromotionTypeMove,
}

type CommonPromoteAppVersion struct {
	PromotionType                string            `json:"promotion_type,omitempty"`
	IncludedRepositoryKeys       []string          `json:"included_repository_keys,omitempty"`
	ExcludedRepositoryKeys       []string          `json:"excluded_repository_keys,omitempty"`
	ArtifactAdditionalProperties map[string]string `json:"artifact_additional_properties,omitempty"`
}

type PromoteAppVersionRequest struct {
	CommonPromoteAppVersion
	Stage string `json:"target_stage"`
}
