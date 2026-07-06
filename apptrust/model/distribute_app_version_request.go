package model

type DistributionRule struct {
	SiteName     string   `json:"site_name,omitempty"`
	CityName     string   `json:"city_name,omitempty"`
	CountryCodes []string `json:"country_codes,omitempty"`
}

type DistributionPathMapping struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type DistributionModifications struct {
	PathMappings []DistributionPathMapping `json:"mappings"`
}

type DistributeAppVersionRequest struct {
	DryRun            bool                      `json:"dry_run"`
	DistributionRules []DistributionRule        `json:"distribution_rules"`
	AutoCreateRepo    bool                      `json:"auto_create_missing_repositories,omitempty"`
	Modifications     DistributionModifications `json:"modifications"`
}
