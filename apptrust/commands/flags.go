package commands

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
)

const (
	Ping              = "ping"
	CreateAppVersion  = "create-app-version"
	PromoteAppVersion = "promote-app-version"
	DeleteAppVersion  = "delete-app-version"
	CreateApp         = "create-app"
	UpdateApp         = "update-app"
	DeleteApp         = "delete-app"
)

const (
	serverId    = "server-id"
	url         = "url"
	user        = "user"
	accessToken = "access-token"
	ProjectFlag = "project"

	ApplicationKeyFlag      = "application-key"
	PackageTypeFlag         = "package-type"
	PackageNameFlag         = "package-name"
	PackageVersionFlag      = "package-version"
	PackageRepositoryFlag   = "package-repository"
	SpecFlag                = "spec"
	SpecVarsFlag            = "spec-vars"
	StageVarsFlag           = "stage"
	ApplicationNameFlag     = "application-name"
	DescriptionFlag         = "desc"
	BusinessCriticalityFlag = "business-criticality"
	MaturityLevelFlag       = "maturity-level"
	LabelsFlag              = "labels"
	UserOwnersFlag          = "user-owners"
	GroupOwnersFlag         = "group-owners"
	SigningKeyFlag          = "signing-key"
)

// Flag keys mapped to their corresponding components.Flag definition.
var flagsMap = map[string]components.Flag{
	// Common commands flags
	serverId:    components.NewStringFlag(serverId, "Server ID configured using the config command.", func(f *components.StringFlag) { f.Mandatory = false }),
	url:         components.NewStringFlag(url, "JFrog Platform URL.", func(f *components.StringFlag) { f.Mandatory = false }),
	user:        components.NewStringFlag(user, "JFrog username.", func(f *components.StringFlag) { f.Mandatory = false }),
	accessToken: components.NewStringFlag(accessToken, "JFrog access token.", func(f *components.StringFlag) { f.Mandatory = false }),
	ProjectFlag: components.NewStringFlag(ProjectFlag, "Project key associated with the application.", func(f *components.StringFlag) { f.Mandatory = true }),

	ApplicationKeyFlag:      components.NewStringFlag(ApplicationKeyFlag, "Application key.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageTypeFlag:         components.NewStringFlag(PackageTypeFlag, "Package type.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageNameFlag:         components.NewStringFlag(PackageNameFlag, "Package name.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageVersionFlag:      components.NewStringFlag(PackageVersionFlag, "Package version.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageRepositoryFlag:   components.NewStringFlag(PackageRepositoryFlag, "Package storing repository.", func(f *components.StringFlag) { f.Mandatory = false }),
	SpecFlag:                components.NewStringFlag(SpecFlag, "A path to the specification file.", func(f *components.StringFlag) { f.Mandatory = false }),
	SpecVarsFlag:            components.NewStringFlag(SpecVarsFlag, "List of semicolon-separated (;) variables in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes) to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.", func(f *components.StringFlag) { f.Mandatory = false }),
	StageVarsFlag:           components.NewStringFlag(StageVarsFlag, "Promotion stage.", func(f *components.StringFlag) { f.Mandatory = true }),
	ApplicationNameFlag:     components.NewStringFlag(ApplicationNameFlag, "The display name of the application.", func(f *components.StringFlag) { f.Mandatory = false }),
	DescriptionFlag:         components.NewStringFlag(DescriptionFlag, "The description of the application.", func(f *components.StringFlag) { f.Mandatory = false }),
	BusinessCriticalityFlag: components.NewStringFlag(BusinessCriticalityFlag, "The business criticality level. The following values are supported: "+coreutils.ListToText(model.BusinessCriticalityValues), func(f *components.StringFlag) { f.DefaultValue = model.BusinessCriticalityValues[0] }),
	MaturityLevelFlag:       components.NewStringFlag(MaturityLevelFlag, "The maturity level.", func(f *components.StringFlag) { f.DefaultValue = model.MaturityLevelValues[0] }),
	LabelsFlag:              components.NewStringFlag(LabelsFlag, "List of semicolon-separated (;) labels in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes).", func(f *components.StringFlag) { f.Mandatory = false }),
	UserOwnersFlag:          components.NewStringFlag(UserOwnersFlag, "Comma-separated list of user owners.", func(f *components.StringFlag) { f.Mandatory = false }),
	GroupOwnersFlag:         components.NewStringFlag(GroupOwnersFlag, "Comma-separated list of group owners.", func(f *components.StringFlag) { f.Mandatory = false }),
	SigningKeyFlag:          components.NewStringFlag(SigningKeyFlag, "The GPG/RSA key-pair name given in Artifactory.", func(f *components.StringFlag) { f.Mandatory = false }),
}

var commandFlags = map[string][]string{
	CreateAppVersion: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationKeyFlag,
		PackageTypeFlag,
		PackageNameFlag,
		PackageVersionFlag,
		PackageRepositoryFlag,
		SpecFlag,
		SpecVarsFlag,
	},
	PromoteAppVersion: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationKeyFlag,
		StageVarsFlag,
	},
	DeleteAppVersion: {
		url,
		user,
		accessToken,
		serverId,
	},

	Ping: {
		url,
		user,
		accessToken,
		serverId,
	},

	CreateApp: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationNameFlag,
		ProjectFlag,
		DescriptionFlag,
		BusinessCriticalityFlag,
		MaturityLevelFlag,
		LabelsFlag,
		UserOwnersFlag,
		GroupOwnersFlag,
		SigningKeyFlag,
		SpecFlag,
		SpecVarsFlag,
	},

	UpdateApp: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationNameFlag,
		DescriptionFlag,
		BusinessCriticalityFlag,
		MaturityLevelFlag,
		LabelsFlag,
		UserOwnersFlag,
		GroupOwnersFlag,
		SigningKeyFlag,
		SpecFlag,
		SpecVarsFlag,
	},
}

func GetCommandFlags(cmdKey string) []components.Flag {
	return pluginsCommon.GetCommandFlags(cmdKey, commandFlags, flagsMap)
}
