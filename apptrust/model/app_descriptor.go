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

	MonitorPolicyTypeNone         = "none"
	MonitorPolicyTypeTimeframe    = "time_frame_in_months"
	MonitorPolicyTypeVersionCount = "version_count"
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

	MonitorPolicyTypeValues = []string{
		MonitorPolicyTypeNone,
		MonitorPolicyTypeTimeframe,
		MonitorPolicyTypeVersionCount,
	}
)

type LabelEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MonitorPolicy struct {
	Type  string `json:"type"`
	Value *int   `json:"value,omitempty"`
}

type LabelUpdates struct {
	Remove []LabelEntry `json:"remove,omitempty"`
	Add    []LabelEntry `json:"add,omitempty"`
}

type AppDescriptor struct {
	ApplicationKey      string         `json:"application_key"`
	ApplicationName     string         `json:"application_name,omitempty"`
	ProjectKey          string         `json:"project_key,omitempty"`
	Description         *string        `json:"description,omitempty"`
	MaturityLevel       *string        `json:"maturity_level,omitempty"`
	BusinessCriticality *string        `json:"criticality,omitempty"`
	Labels              *[]LabelEntry  `json:"labels,omitempty"`
	LabelUpdates        *LabelUpdates  `json:"label_updates,omitempty"`
	UserOwners          *[]string      `json:"user_owners,omitempty"`
	GroupOwners         *[]string      `json:"group_owners,omitempty"`
	MonitorPolicy       *MonitorPolicy `json:"monitor_policy,omitempty"`
	AutoPromoteStages   *[]string      `json:"auto_promote_stages,omitempty"`
}
