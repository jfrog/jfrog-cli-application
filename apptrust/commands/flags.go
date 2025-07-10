package commands

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
)

const (
	Ping           = "ping"
	VersionCreate  = "version-create"
	VersionPromote = "version-promote"
	VersionDelete  = "version-delete"
	VersionRelease = "version-release"
	VersionUpdate  = "version-update"
	PackageBind    = "package-bind"
	PackageUnbind  = "package-unbind"
	AppCreate      = "app-create"
	AppUpdate      = "app-update"
	AppDelete      = "app-delete"
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
	SyncFlag                = "sync"
	PromotionTypeFlag       = "promotion-type"
	DryRunFlag              = "dry-run"
	ExcludeReposFlag        = "exclude-repos"
	IncludeReposFlag        = "include-repos"
	PropsFlag               = "props"
	TagFlag                 = "tag"
	BuildsFlag              = "builds"
	ReleaseBundlesFlag      = "release-bundles"
	SourceVersionFlag       = "source-version"
	PackagesFlag            = "packages"
	PropertiesFlag          = "properties"
	DeletePropertyFlag      = "delete-property"
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
	SyncFlag:                components.NewBoolFlag(SyncFlag, "Whether to synchronize the operation.", components.WithBoolDefaultValueTrue()),
	PromotionTypeFlag:       components.NewStringFlag(PromotionTypeFlag, "The promotion type. The following values are supported: "+coreutils.ListToText(model.PromotionTypeValues), func(f *components.StringFlag) { f.Mandatory = false; f.DefaultValue = model.PromotionTypeCopy }),
	DryRunFlag:              components.NewBoolFlag(DryRunFlag, "Perform a simulation of the operation.", components.WithBoolDefaultValueFalse()),
	ExcludeReposFlag:        components.NewStringFlag(ExcludeReposFlag, "Semicolon-separated list of repositories to exclude.", func(f *components.StringFlag) { f.Mandatory = false }),
	IncludeReposFlag:        components.NewStringFlag(IncludeReposFlag, "Semicolon-separated list of repositories to include.", func(f *components.StringFlag) { f.Mandatory = false }),
	PropsFlag:               components.NewStringFlag(PropsFlag, "Semicolon-separated list of properties in the form of 'key1=value1;key2=value2;...' to be added to each artifact.", func(f *components.StringFlag) { f.Mandatory = false }),
	TagFlag:                 components.NewStringFlag(TagFlag, "A tag to associate with the version. Must contain only alphanumeric characters, hyphens (-), underscores (_), and dots (.).", func(f *components.StringFlag) { f.Mandatory = false }),
	BuildsFlag:              components.NewStringFlag(BuildsFlag, "List of builds in format 'name1:number1[:timestamp1];name2:number2[:timestamp2]'", func(f *components.StringFlag) { f.Mandatory = false }),
	ReleaseBundlesFlag:      components.NewStringFlag(ReleaseBundlesFlag, "List of release bundles in format 'name1:version1;name2:version2'", func(f *components.StringFlag) { f.Mandatory = false }),
	SourceVersionFlag:       components.NewStringFlag(SourceVersionFlag, "Source versions in format 'app1:version1;app2:version2'", func(f *components.StringFlag) { f.Mandatory = false }),
	PackagesFlag:            components.NewStringFlag(PackagesFlag, "List of packages in format 'name1;name2'", func(f *components.StringFlag) { f.Mandatory = false }),
	PropertiesFlag:          components.NewStringFlag(PropertiesFlag, "Sets or updates custom properties for the application version in format 'key1=value1[,value2,...];key2=value3[,value4,...]'", func(f *components.StringFlag) { f.Mandatory = false }),
	DeletePropertyFlag:      components.NewStringFlag(DeletePropertyFlag, "Remove a property key and all its values", func(f *components.StringFlag) { f.Mandatory = false }),
}

var commandFlags = map[string][]string{
	VersionCreate: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationKeyFlag,
		TagFlag,
		PackagesFlag,
		PackageTypeFlag,
		PackageRepositoryFlag,
		BuildsFlag,
		ReleaseBundlesFlag,
		SourceVersionFlag,
		SpecFlag,
		SpecVarsFlag,
	},
	VersionPromote: {
		url,
		user,
		accessToken,
		serverId,
		SyncFlag,
		PromotionTypeFlag,
		DryRunFlag,
		ExcludeReposFlag,
		IncludeReposFlag,
		PropsFlag,
	},
	VersionRelease: {
		url,
		user,
		accessToken,
		serverId,
		SyncFlag,
		PromotionTypeFlag,
		ExcludeReposFlag,
		IncludeReposFlag,
		PropsFlag,
	},
	VersionDelete: {
		url,
		user,
		accessToken,
		serverId,
	},
	VersionUpdate: {
		url,
		user,
		accessToken,
		serverId,
		TagFlag,
		PropertiesFlag,
		DeletePropertyFlag,
	},

	PackageBind: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationKeyFlag,
		PackagesFlag,
		PackageTypeFlag,
	},
	PackageUnbind: {
		url,
		user,
		accessToken,
		serverId,
		ApplicationKeyFlag,
		PackagesFlag,
		PackageTypeFlag,
	},

	Ping: {
		url,
		user,
		accessToken,
		serverId,
	},

	AppCreate: {
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

	AppUpdate: {
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

	AppDelete: {
		url,
		user,
		accessToken,
		serverId,
	},
}

func GetCommandFlags(cmdKey string) []components.Flag {
	return pluginsCommon.GetCommandFlags(cmdKey, commandFlags, flagsMap)
}
