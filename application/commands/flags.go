package commands

import (
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

const (
	Ping              = "ping"
	CreateAppVersion  = "create-app-version"
	PromoteAppVersion = "promote-app-version"
)

const (
	ServerId    = "server-id"
	url         = "url"
	user        = "user"
	accessToken = "access-token"
	ProjectFlag = "project"

	ApplicationKeyFlag    = "app-key"
	PackageTypeFlag       = "package-type"
	PackageNameFlag       = "package-name"
	PackageVersionFlag    = "package-version"
	PackageRepositoryFlag = "package-repository"
	SpecFlag              = "spec"
	SpecVarsFlag          = "spec-vars"
	EnvironmentVarsFlag   = "env"
)

// Flag keys mapped to their corresponding components.Flag definition.
var flagsMap = map[string]components.Flag{
	// Common commands flags
	ServerId:    components.NewStringFlag(ServerId, "Server ID configured using the config command.", func(f *components.StringFlag) { f.Mandatory = false }),
	url:         components.NewStringFlag(url, "JFrog Platform URL.", func(f *components.StringFlag) { f.Mandatory = false }),
	user:        components.NewStringFlag(user, "JFrog username.", func(f *components.StringFlag) { f.Mandatory = false }),
	accessToken: components.NewStringFlag(accessToken, "JFrog access token.", func(f *components.StringFlag) { f.Mandatory = false }),
	ProjectFlag: components.NewStringFlag(ProjectFlag, "Project key associated with the created evidence.", func(f *components.StringFlag) { f.Mandatory = false }),

	ApplicationKeyFlag:    components.NewStringFlag(ApplicationKeyFlag, "Application key.", func(f *components.StringFlag) { f.Mandatory = true }),
	PackageTypeFlag:       components.NewStringFlag(PackageTypeFlag, "Package type.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageNameFlag:       components.NewStringFlag(PackageNameFlag, "Package name.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageVersionFlag:    components.NewStringFlag(PackageVersionFlag, "Package version.", func(f *components.StringFlag) { f.Mandatory = false }),
	PackageRepositoryFlag: components.NewStringFlag(PackageRepositoryFlag, "Package storing repository.", func(f *components.StringFlag) { f.Mandatory = false }),
	SpecFlag:              components.NewStringFlag(SpecFlag, "A path to the specification file.", func(f *components.StringFlag) { f.Mandatory = false }),
	SpecVarsFlag:          components.NewStringFlag(SpecVarsFlag, "List of semicolon-separated(;) variables in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes) to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.` `", func(f *components.StringFlag) { f.Mandatory = false }),
	EnvironmentVarsFlag:   components.NewStringFlag(EnvironmentVarsFlag, "Environment.", func(f *components.StringFlag) { f.Mandatory = true }),
}

var commandFlags = map[string][]string{
	CreateAppVersion: {
		url,
		user,
		accessToken,
		ServerId,
		ProjectFlag,
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
		ServerId,
		ProjectFlag,
		ApplicationKeyFlag,
		EnvironmentVarsFlag,
	},

	Ping: {
		url,
		user,
		accessToken,
		ServerId,
	},
}

func GetCommandFlags(cmdKey string) []components.Flag {
	return pluginsCommon.GetCommandFlags(cmdKey, commandFlags, flagsMap)
}
