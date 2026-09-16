package common

// OrderedAppVersionKeys defines the display order for application-version table output
// (shared by version-create, version-update, version-update-sources).
var OrderedAppVersionKeys = []string{
	"application_key",
	"version",
	"status",
	"current_stage",
	"tag",
	"message",
}

// OrderedAppKeys defines the display order for application table output
// (shared by app-create, app-update).
var OrderedAppKeys = []string{
	"application_key",
	"application_name",
	"project_key",
	"description",
	"criticality",
	"maturity_level",
	"auto_promote_stages",
}
