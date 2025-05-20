package model

type AppDescriptor struct {
	ApplicationKey      string            `json:"application_key"`
	ApplicationName     string            `json:"application_name,omitempty"`
	ProjectKey          string            `json:"project_key,omitempty"`
	Description         string            `json:"description,omitempty"`
	MaturityLevel       string            `json:"maturity_level,omitempty"`
	BusinessCriticality string            `json:"criticality,omitempty"`
	Labels              map[string]string `json:"labels,omitempty"`
	UserOwners          []string          `json:"user_owners,omitempty"`
	GroupOwners         []string          `json:"group_owners,omitempty"`
}

var BusinessCriticalityValues = []string{
	"unspecified",
	"low",
	"medium",
	"high",
	"critical",
}

var MaturityLevelValues = []string{
	"unspecified",
	"experimental",
	"production",
	"end_of_life",
}
