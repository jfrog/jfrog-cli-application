package commands

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
)

const (
	Ping                 = "ping"
	VersionCreate        = "version-create"
	VersionPromote       = "version-promote"
	VersionRollback      = "version-rollback"
	VersionDelete        = "version-delete"
	VersionRelease       = "version-release"
	VersionUpdate        = "version-update"
	VersionUpdateSources = "version-update-sources"
	VersionDistribute    = "version-distribute"
	VersionRemoteDelete  = "version-delete-remote"
	PackageBind          = "package-bind"
	PackageUnbind        = "package-unbind"
	AppCreate            = "app-create"
	AppUpdate            = "app-update"
	AppDelete            = "app-delete"
	AppExport            = "app-export"
	AppImport            = "app-import"
)

const (
	serverId    = "server-id"
	url         = "url"
	user        = "user"
	accessToken = "access-token"
	ProjectFlag = "project"

	SpecFlag                          = "spec"
	SpecVarsFlag                      = "spec-vars"
	StageVarsFlag                     = "stage"
	ApplicationNameFlag               = "application-name"
	DescriptionFlag                   = "desc"
	BusinessCriticalityFlag           = "business-criticality"
	MaturityLevelFlag                 = "maturity-level"
	LabelsFlag                        = "labels"
	AddLabelsFlag                     = "add-labels"
	RemoveLabelsFlag                  = "remove-labels"
	UserOwnersFlag                    = "user-owners"
	GroupOwnersFlag                   = "group-owners"
	MonitorPolicyFlag                 = "monitor-policy"
	SyncFlag                          = "sync"
	PromotionTypeFlag                 = "promotion-type"
	DryRunFlag                        = "dry-run"
	FailFastFlag                      = "fail-fast"
	ExcludeReposFlag                  = "exclude-repos"
	IncludeReposFlag                  = "include-repos"
	PropsFlag                         = "props"
	OverwriteStrategyFlag             = "overwrite-strategy"
	TagFlag                           = "tag"
	DraftFlag                         = "draft"
	SkipUnassignedFlag                = "skip-unassigned"
	SourceTypeBuildsFlag              = "source-type-builds"
	SourceTypeReleaseBundlesFlag      = "source-type-release-bundles"
	SourceTypeApplicationVersionsFlag = "source-type-application-versions"
	SourceTypePackagesFlag            = "source-type-packages"
	SourceTypeArtifactsFlag           = "source-type-artifacts"
	PropertiesFlag                    = "properties"
	DeletePropertiesFlag              = "delete-properties"
	IncludeFilterFlag                 = "include-filter"
	ExcludeFilterFlag                 = "exclude-filter"
	ConflictResolutionFlag            = "conflict-resolution"
	PathMappingFlag                   = "path-mapping"
	DistRulesFlag                     = "dist-rules"
	SiteFlag                          = "site"
	CityFlag                          = "city"
	CountryCodesFlag                  = "country-codes"
	CreateRepoFlag                    = "create-repo"
	MappingPatternFlag                = "mapping-pattern"
	MappingTargetFlag                 = "mapping-target"
	QuietFlag                         = "quiet"
)

// Flag keys mapped to their corresponding components.Flag definition.
var flagsMap = map[string]components.Flag{
	// Common commands flags
	serverId:    components.NewStringFlag(serverId, "Server ID configured using the config command.", func(f *components.StringFlag) { f.Mandatory = false }),
	url:         components.NewStringFlag(url, "JFrog Platform URL.", func(f *components.StringFlag) { f.Mandatory = false }),
	user:        components.NewStringFlag(user, "JFrog username.", func(f *components.StringFlag) { f.Mandatory = false }),
	accessToken: components.NewStringFlag(accessToken, "JFrog access token.", func(f *components.StringFlag) { f.Mandatory = false }),
	ProjectFlag: components.NewStringFlag(ProjectFlag, "Project key associated with the application. This flag is mandatory when the --spec flag is not provided.", func(f *components.StringFlag) { f.Mandatory = false }),

	SpecFlag:                          components.NewStringFlag(SpecFlag, "A path to the specification file.", func(f *components.StringFlag) { f.Mandatory = false }),
	SpecVarsFlag:                      components.NewStringFlag(SpecVarsFlag, "List of semicolon-separated (;) variables in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes) to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.", func(f *components.StringFlag) { f.Mandatory = false }),
	StageVarsFlag:                     components.NewStringFlag(StageVarsFlag, "Promotion stage.", func(f *components.StringFlag) { f.Mandatory = true }),
	ApplicationNameFlag:               components.NewStringFlag(ApplicationNameFlag, "The display name of the application.", func(f *components.StringFlag) { f.Mandatory = false }),
	DescriptionFlag:                   components.NewStringFlag(DescriptionFlag, "The description of the application.", func(f *components.StringFlag) { f.Mandatory = false }),
	BusinessCriticalityFlag:           components.NewStringFlag(BusinessCriticalityFlag, "The business criticality level. The following values are supported: "+coreutils.ListToText(model.BusinessCriticalityValues), func(f *components.StringFlag) { f.Mandatory = false }),
	MaturityLevelFlag:                 components.NewStringFlag(MaturityLevelFlag, "The maturity level. The following values are supported: "+coreutils.ListToText(model.MaturityLevelValues), func(f *components.StringFlag) { f.Mandatory = false }),
	LabelsFlag:                        components.NewStringFlag(LabelsFlag, "List of semicolon-separated (;) labels in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes).", func(f *components.StringFlag) { f.Mandatory = false }),
	AddLabelsFlag:                     components.NewStringFlag(AddLabelsFlag, "List of semicolon-separated (;) labels to add in the form of \"key1=value1;key1=value2;key2=value3;...\" (wrapped by quotes)..", func(f *components.StringFlag) { f.Mandatory = false }),
	RemoveLabelsFlag:                  components.NewStringFlag(RemoveLabelsFlag, "List of semicolon-separated (;) labels to remove in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes).", func(f *components.StringFlag) { f.Mandatory = false }),
	UserOwnersFlag:                    components.NewStringFlag(UserOwnersFlag, "semicolon-separated (;) list of user owners in the form of \"user1;user2;...\" (wrapped by quotes).", func(f *components.StringFlag) { f.Mandatory = false }),
	GroupOwnersFlag:                   components.NewStringFlag(GroupOwnersFlag, "semicolon-separated (;) list of group owners in the form of \"group1;group2;...\" (wrapped by quotes).", func(f *components.StringFlag) { f.Mandatory = false }),
	MonitorPolicyFlag:                 components.NewStringFlag(MonitorPolicyFlag, "Defines how long application versions remain monitored, in the form of 'type=<type>[, value=<n>]'. Supported types: "+coreutils.ListToText(model.MonitorPolicyTypeValues)+". For '"+model.MonitorPolicyTypeTimeframe+"' or '"+model.MonitorPolicyTypeVersionCount+"', 'value' is the number of months or versions to keep monitoring (a positive integer). For '"+model.MonitorPolicyTypeNone+"', monitoring is disabled and 'value' must be omitted.", func(f *components.StringFlag) { f.Mandatory = false }),
	SyncFlag:                          components.NewBoolFlag(SyncFlag, "Whether to synchronize the operation.", components.WithBoolDefaultValueTrue()),
	PromotionTypeFlag:                 components.NewStringFlag(PromotionTypeFlag, "The promotion type. The following values are supported: "+coreutils.ListToText(model.PromotionTypeValues), func(f *components.StringFlag) { f.Mandatory = false; f.DefaultValue = model.PromotionTypeCopy }),
	DryRunFlag:                        components.NewBoolFlag(DryRunFlag, "Perform a simulation of the operation.", components.WithBoolDefaultValueFalse()),
	FailFastFlag:                      components.NewBoolFlag(FailFastFlag, "Stop the operation on the first error. Only relevant when sources are provided.", components.WithBoolDefaultValueTrue()),
	ExcludeReposFlag:                  components.NewStringFlag(ExcludeReposFlag, "Semicolon-separated list of repositories to exclude.", func(f *components.StringFlag) { f.Mandatory = false }),
	IncludeReposFlag:                  components.NewStringFlag(IncludeReposFlag, "Semicolon-separated list of repositories to include.", func(f *components.StringFlag) { f.Mandatory = false }),
	PropsFlag:                         components.NewStringFlag(PropsFlag, "Semicolon-separated list of properties in the form of 'key1=value1;key2=value2;...' to be added to each artifact.", func(f *components.StringFlag) { f.Mandatory = false }),
	OverwriteStrategyFlag:             components.NewStringFlag(OverwriteStrategyFlag, "Strategy for handling target artifacts with the same path but different checksum. Supported values: "+coreutils.ListToText(model.OverwriteStrategyValues)+".", func(f *components.StringFlag) { f.Mandatory = false }),
	TagFlag:                           components.NewStringFlag(TagFlag, "A tag to associate with the version. Must contain only alphanumeric characters, hyphens (-), underscores (_), and dots (.).", func(f *components.StringFlag) { f.Mandatory = false }),
	DraftFlag:                         components.NewBoolFlag(DraftFlag, "Create the application version as a draft.", components.WithBoolDefaultValueFalse()),
	SkipUnassignedFlag:                components.NewBoolFlag(SkipUnassignedFlag, "Automatically promote the new version to the first lifecycle stage when all of its source artifacts reside in repositories mapped to that stage. Otherwise the version is left unassigned and a message explaining why is returned.", components.WithBoolDefaultValueFalse()),
	SourceTypeBuildsFlag:              components.NewStringFlag(SourceTypeBuildsFlag, "List of semicolon-separated (;) builds in the form of 'name=buildName1, id=runID1[, include-deps=true][, repo-key=repo1][, started=2023-01-01T12:34:56.789+0100]; name=buildName2, id=runID2[, include-deps=true][, repo-key=repo2][, started=2023-01-01T12:34:56.789+0100]' to be included in the new version.", func(f *components.StringFlag) { f.Mandatory = false }),
	SourceTypeReleaseBundlesFlag:      components.NewStringFlag(SourceTypeReleaseBundlesFlag, "List of semicolon-separated (;) release bundles in the form of 'name=releaseBundleName1, version=version1[, project-key=project1][, repo-key=repo1]; name=releaseBundleName2, version=version2[, project-key=project2][, repo-key=repo2]' to be included in the new version.", func(f *components.StringFlag) { f.Mandatory = false }),
	SourceTypeApplicationVersionsFlag: components.NewStringFlag(SourceTypeApplicationVersionsFlag, "List of semicolon-separated (;) application versions in the form of 'application-key=app1, version=version1; application-key=app2, version=version2' to be included in the new version.", func(f *components.StringFlag) { f.Mandatory = false }),
	SourceTypePackagesFlag:            components.NewStringFlag(SourceTypePackagesFlag, "List of semicolon-separated (;) packages in the form of 'type=packageType1, name=packageName1, version=version1, repo-key=repo1; type=packageType2, name=packageName2, version=version2, repo-key=repo2' to be included in the new version.", func(f *components.StringFlag) { f.Mandatory = false }),
	IncludeFilterFlag:                 components.NewStringFlag(IncludeFilterFlag, "List of semicolon-separated (;) filters of packages and artifacts in the form of 'filter1; filter2...' to be included in the new version. Each filter must be comma-separated: 'filter_type=package/artifact, field1=value1[, field2=value2...]'. Package filters require at least one of: 'type', 'name', or 'version'. Artifact filters require at least one of: 'path' or 'sha256'.", func(f *components.StringFlag) { f.Mandatory = false }),
	ExcludeFilterFlag:                 components.NewStringFlag(ExcludeFilterFlag, "List of semicolon-separated (;) filters of packages and artifacts in the form of 'filter1; filter2...' to be included in the new version. Each filter must be comma-separated: 'filter_type=package/artifact, field1=value1[, field2=value2...]'. Package filters require at least one of: 'type', 'name', or 'version'. Artifact filters require at least one of: 'path' or 'sha256'.", func(f *components.StringFlag) { f.Mandatory = false }),
	SourceTypeArtifactsFlag:           components.NewStringFlag(SourceTypeArtifactsFlag, "List of semicolon-separated (;) artifacts in the form of 'path=repo/path/to/artifact1[, sha256=hash1]; path=repo/path/to/artifact2[, sha256=hash2]' to be included in the new version.", func(f *components.StringFlag) { f.Mandatory = false }),
	PropertiesFlag:                    components.NewStringFlag(PropertiesFlag, "Sets or updates custom properties for the application version in format 'key1=value1[,value2,...];key2=value3[,value4,...]'", func(f *components.StringFlag) { f.Mandatory = false }),
	DeletePropertiesFlag:              components.NewStringFlag(DeletePropertiesFlag, "Remove a property key and all its values", func(f *components.StringFlag) { f.Mandatory = false }),
	ConflictResolutionFlag:            components.NewStringFlag(ConflictResolutionFlag, "How to resolve source conflicts when the same artifact path appears in multiple sources. Supported values: "+coreutils.ListToText(model.ConflictResolutionValues)+".", func(f *components.StringFlag) { f.Mandatory = false }),
	PathMappingFlag:                   components.NewStringFlag(PathMappingFlag, "List of semicolon-separated (;) path mapping rules in the form of 'input=(.*), output=stable-release/$1[, package-type=.*]; input=(.*\\.jar), output=jars/$1, package-type=maven'. Note: quote the value to prevent shell expansion of $1.", func(f *components.StringFlag) { f.Mandatory = false }),
	DistRulesFlag:                     components.NewStringFlag(DistRulesFlag, "Path to distribution rules.", func(f *components.StringFlag) { f.Mandatory = false }),
	SiteFlag:                          components.NewStringFlag(SiteFlag, "Wildcard filter for site name.", func(f *components.StringFlag) { f.Mandatory = false }),
	CityFlag:                          components.NewStringFlag(CityFlag, "Wildcard filter for site city name.", func(f *components.StringFlag) { f.Mandatory = false }),
	CountryCodesFlag:                  components.NewStringFlag(CountryCodesFlag, "List of semicolon-separated (;) wildcard filters for site country codes.", func(f *components.StringFlag) { f.Mandatory = false }),
	CreateRepoFlag:                    components.NewBoolFlag(CreateRepoFlag, "Set to true to create the repository on the edge if it does not exist.", components.WithBoolDefaultValueFalse()),
	MappingPatternFlag:                components.NewStringFlag(MappingPatternFlag, "Specify along with "+MappingTargetFlag+" to distribute artifacts to a different path on the edge node. You can use wildcards to specify multiple artifacts.", func(f *components.StringFlag) { f.Mandatory = false }),
	MappingTargetFlag:                 components.NewStringFlag(MappingTargetFlag, "The target path for distributed artifacts on the edge node. If not specified, the artifacts will have the same path and name on the edge node, as on the source Artifactory server. For flexibility in specifying the distribution path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the pattern path that are enclosed in parenthesis.", func(f *components.StringFlag) { f.Mandatory = false }),
	QuietFlag:                         components.NewBoolFlag(QuietFlag, "Set to true to skip the confirmation message. When $CI is true, the default value is true.", components.WithBoolDefaultValueFalse()),
}

var commandFlags = map[string][]string{
	VersionCreate: {
		url,
		user,
		accessToken,
		serverId,
		SyncFlag,
		TagFlag,
		DraftFlag,
		SkipUnassignedFlag,
		ConflictResolutionFlag,
		SourceTypeBuildsFlag,
		SourceTypeReleaseBundlesFlag,
		SourceTypeApplicationVersionsFlag,
		SourceTypePackagesFlag,
		SourceTypeArtifactsFlag,
		SpecFlag,
		IncludeFilterFlag,
		ExcludeFilterFlag,
		SpecVarsFlag,
		DryRunFlag,
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
		OverwriteStrategyFlag,
		PathMappingFlag,
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
		OverwriteStrategyFlag,
		PathMappingFlag,
	},
	VersionDelete: {
		url,
		user,
		accessToken,
		serverId,
	},
	VersionRollback: {
		url,
		user,
		accessToken,
		serverId,
		SyncFlag,
	},
	VersionUpdate: {
		url,
		user,
		accessToken,
		serverId,
		TagFlag,
		PropertiesFlag,
		DeletePropertiesFlag,
	},
	VersionUpdateSources: {
		url,
		user,
		accessToken,
		serverId,
		SyncFlag,
		DryRunFlag,
		FailFastFlag,
		SourceTypeBuildsFlag,
		SourceTypeReleaseBundlesFlag,
		SourceTypeApplicationVersionsFlag,
		SourceTypePackagesFlag,
		SourceTypeArtifactsFlag,
		SpecFlag,
		SpecVarsFlag,
		IncludeFilterFlag,
		ExcludeFilterFlag,
	},

	VersionDistribute: {
		url,
		user,
		accessToken,
		serverId,
		DistRulesFlag,
		SiteFlag,
		CityFlag,
		CountryCodesFlag,
		CreateRepoFlag,
		MappingPatternFlag,
		MappingTargetFlag,
	},
	VersionRemoteDelete: {
		url,
		user,
		accessToken,
		serverId,
		QuietFlag,
		DryRunFlag,
		DistRulesFlag,
		SiteFlag,
		CityFlag,
		CountryCodesFlag,
	},

	PackageBind: {
		url,
		user,
		accessToken,
		serverId,
	},
	PackageUnbind: {
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
		MonitorPolicyFlag,
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
		AddLabelsFlag,
		RemoveLabelsFlag,
		UserOwnersFlag,
		GroupOwnersFlag,
		MonitorPolicyFlag,
	},

	AppDelete: {
		url,
		user,
		accessToken,
		serverId,
	},

	AppExport: {
		url,
		user,
		accessToken,
		serverId,
	},

	AppImport: {
		url,
		user,
		accessToken,
		serverId,
	},
}

func GetCommandFlags(cmdKey string) []components.Flag {
	return pluginsCommon.GetCommandFlags(cmdKey, commandFlags, flagsMap)
}
