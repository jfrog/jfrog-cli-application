package model

const (
	PromotionTypeCopy = "copy"
	PromotionTypeMove = "move"

	// This is not a valid value for the --promotion-type flag, but is passed to the API if the --dry-run flag is set.
	PromotionTypeDryRun = "dry_run"
)

var (
	PromotionTypeValues = []string{
		PromotionTypeCopy,
		PromotionTypeMove,
	}
)

type PromoteAppVersionRequest struct {
	Stage                  string   `json:"stage"`
	PromotionType          string   `json:"promotion_type,omitempty"`
	IncludedRepositoryKeys []string `json:"included_repository_keys,omitempty"`
	ExcludedRepositoryKeys []string `json:"excluded_repository_keys,omitempty"`
}
