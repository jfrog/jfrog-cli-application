package model

type RemoteDeleteAppVersionRequest struct {
	DryRun            bool               `json:"dry_run"`
	DistributionRules []DistributionRule `json:"distribution_rules"`
	AutoCreateRepo    bool               `json:"auto_create_missing_repositories,omitempty"`
}
