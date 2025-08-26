package model

const (
	BusinessCriticalityUnspecified = "unspecified"
	BusinessCriticalityLow         = "low"
	BusinessCriticalityMedium      = "medium"
	BusinessCriticalityHigh        = "high"
	BusinessCriticalityCritical    = "critical"

	MaturityLevelUnspecified  = "unspecified"
	MaturityLevelExperimental = "experimental"
	MaturityLevelProduction   = "production"
	MaturityLevelEndOfLife    = "end_of_life"
)

var (
	BusinessCriticalityValues = []string{
		BusinessCriticalityUnspecified,
		BusinessCriticalityLow,
		BusinessCriticalityMedium,
		BusinessCriticalityHigh,
		BusinessCriticalityCritical,
	}

	MaturityLevelValues = []string{
		MaturityLevelUnspecified,
		MaturityLevelExperimental,
		MaturityLevelProduction,
		MaturityLevelEndOfLife,
	}
)

type AppDescriptor struct {
	ApplicationKey      string             `json:"application_key"`
	ApplicationName     string             `json:"application_name,omitempty"`
	ProjectKey          string             `json:"project_key,omitempty"`
	Description         *string            `json:"description,omitempty"`
	MaturityLevel       *string            `json:"maturity_level,omitempty"`
	BusinessCriticality *string            `json:"criticality,omitempty"`
	Labels              *map[string]string `json:"labels,omitempty"`
	UserOwners          *[]string          `json:"user_owners,omitempty"`
	GroupOwners         *[]string          `json:"group_owners,omitempty"`
}
