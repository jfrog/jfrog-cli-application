package version

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-application/apptrust/service/versions"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type createAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	requestPayload *model.CreateAppVersionRequest
}

type createVersionSpec struct {
	Packages       []model.CreateVersionPackage       `json:"packages,omitempty"`
	Builds         []model.CreateVersionBuild         `json:"builds,omitempty"`
	ReleaseBundles []model.CreateVersionReleaseBundle `json:"release_bundles,omitempty"`
	Versions       []model.CreateVersionReference     `json:"versions,omitempty"`
}

func (cv *createAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*cv.serverDetails)
	if err != nil {
		return err
	}

	return cv.versionService.CreateAppVersion(ctx, cv.requestPayload)
}

func (cv *createAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return cv.serverDetails, nil
}

func (cv *createAppVersionCommand) CommandName() string {
	return commands.VersionCreate
}

func (cv *createAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if err := validateCreateAppVersionContext(ctx); err != nil {
		return err
	}
	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	cv.serverDetails = serverDetails
	cv.requestPayload, err = cv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}
	return commonCLiCommands.Exec(cv)
}

func (cv *createAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.CreateAppVersionRequest, error) {
	var (
		sources *model.CreateVersionSources
		err     error
	)

	if ctx.IsFlagSet(commands.SpecFlag) {
		sources, err = cv.loadFromSpec(ctx)
	} else {
		sources, err = cv.buildSourcesFromFlags(ctx)
	}

	if err != nil {
		return nil, err
	}

	return &model.CreateAppVersionRequest{
		ApplicationKey: ctx.Arguments[0],
		Version:        ctx.Arguments[1],
		Sources:        sources,
		Tag:            ctx.GetStringFlagValue(commands.TagFlag),
	}, nil
}

func (cv *createAppVersionCommand) buildSourcesFromFlags(ctx *components.Context) (*model.CreateVersionSources, error) {
	sources := &model.CreateVersionSources{}
	if buildsStr := ctx.GetStringFlagValue(commands.SourceTypeBuildsFlag); buildsStr != "" {
		builds, err := cv.parseBuilds(buildsStr)
		if err != nil {
			return nil, err
		}
		sources.Builds = builds
	}
	if rbStr := ctx.GetStringFlagValue(commands.SourceTypeReleaseBundlesFlag); rbStr != "" {
		releaseBundles, err := cv.parseReleaseBundles(rbStr)
		if err != nil {
			return nil, err
		}
		sources.ReleaseBundles = releaseBundles
	}
	if srcVersionsStr := ctx.GetStringFlagValue(commands.SourceTypeApplicationVersionsFlag); srcVersionsStr != "" {
		sourceVersions, err := cv.parseSourceVersions(srcVersionsStr)
		if err != nil {
			return nil, err
		}
		sources.Versions = sourceVersions
	}
	return sources, nil
}

func (cv *createAppVersionCommand) loadFromSpec(ctx *components.Context) (*model.CreateVersionSources, error) {
	specFilePath := ctx.GetStringFlagValue(commands.SpecFlag)
	spec := new(createVersionSpec)
	specVars := coreutils.SpecVarsStringToMap(ctx.GetStringFlagValue(commands.SpecVarsFlag))
	content, err := fileutils.ReadFile(specFilePath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	if len(specVars) > 0 {
		content = coreutils.ReplaceVars(content, specVars)
	}

	err = json.Unmarshal(content, spec)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	// Validation: if all sources are empty, return error
	if (len(spec.Packages) == 0) && (len(spec.Builds) == 0) && (len(spec.ReleaseBundles) == 0) && (len(spec.Versions) == 0) {
		return nil, errorutils.CheckErrorf("Spec file is empty: must provide at least one source (packages, builds, release_bundles, or versions)")
	}

	sources := &model.CreateVersionSources{
		Packages:       spec.Packages,
		Builds:         spec.Builds,
		ReleaseBundles: spec.ReleaseBundles,
		Versions:       spec.Versions,
	}

	return sources, nil
}

func (cv *createAppVersionCommand) parseBuilds(buildsStr string) ([]model.CreateVersionBuild, error) {
	const (
		nameField       = "name"
		idField         = "id"
		includeDepField = "include_deps"
	)

	var builds []model.CreateVersionBuild
	buildEntries := utils.ParseSliceFlag(buildsStr)
	for _, entry := range buildEntries {
		buildEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid build format: %v", err)
		}
		err = validateRequiredFieldsInMap(buildEntryMap, nameField, idField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid build format: %v", err)
		}
		build := model.CreateVersionBuild{
			Name:   buildEntryMap[nameField],
			Number: buildEntryMap[idField],
		}
		if _, ok := buildEntryMap[includeDepField]; ok {
			includeDep, err := strconv.ParseBool(buildEntryMap[includeDepField])
			if err != nil {
				return nil, errorutils.CheckErrorf("invalid build format: %v", err)
			}
			build.IncludeDependencies = includeDep
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (cv *createAppVersionCommand) parseReleaseBundles(rbStr string) ([]model.CreateVersionReleaseBundle, error) {
	const (
		nameField    = "name"
		versionField = "version"
	)

	var bundles []model.CreateVersionReleaseBundle
	releaseBundleEntries := utils.ParseSliceFlag(rbStr)
	for _, entry := range releaseBundleEntries {
		releaseBundleEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid release bundle format: %v", err)
		}
		err = validateRequiredFieldsInMap(releaseBundleEntryMap, nameField, versionField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid release bundle format: %v", err)
		}
		bundles = append(bundles, model.CreateVersionReleaseBundle{
			Name:    releaseBundleEntryMap[nameField],
			Version: releaseBundleEntryMap[versionField],
		})
	}
	return bundles, nil
}

func (cv *createAppVersionCommand) parseSourceVersions(applicationVersionsStr string) ([]model.CreateVersionReference, error) {
	const (
		applicationKeyField = "application-key"
		versionField        = "version"
	)

	var refs []model.CreateVersionReference
	applicationVersionEntries := utils.ParseSliceFlag(applicationVersionsStr)
	for _, entry := range applicationVersionEntries {
		applicationVersionEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid application version format: %v", err)
		}
		err = validateRequiredFieldsInMap(applicationVersionEntryMap, applicationKeyField, versionField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid application version format: %v", err)
		}
		refs = append(refs, model.CreateVersionReference{
			ApplicationKey: applicationVersionEntryMap[applicationKeyField],
			Version:        applicationVersionEntryMap[versionField],
		})
	}
	return refs, nil
}

func validateCreateAppVersionContext(ctx *components.Context) error {
	if err := validateNoSpecAndFlagsTogether(ctx); err != nil {
		return err
	}
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	hasSource := ctx.IsFlagSet(commands.SpecFlag) ||
		ctx.IsFlagSet(commands.SourceTypeBuildsFlag) ||
		ctx.IsFlagSet(commands.SourceTypeReleaseBundlesFlag) ||
		ctx.IsFlagSet(commands.SourceTypeApplicationVersionsFlag)

	if !hasSource {
		return errorutils.CheckErrorf(
			"At least one source flag is required to create an application version. Please provide one of the following: --%s, --%s, --%s, or --%s.",
			commands.SpecFlag, commands.SourceTypeBuildsFlag, commands.SourceTypeReleaseBundlesFlag, commands.SourceTypeApplicationVersionsFlag)
	}

	return nil
}

func GetCreateAppVersionCommand(appContext app.Context) components.Command {
	cmd := &createAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.VersionCreate,
		Description: "Create application version.",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vc"},
		Arguments: []components.Argument{
			{
				Name:        "app-key",
				Description: "The application key of the application for which the version is being created.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version number (in SemVer format) for the new application version.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionCreate),
		Action: cmd.prepareAndRunCommand,
	}
}

// Returns error if both --spec and any other source flag are set
func validateNoSpecAndFlagsTogether(ctx *components.Context) error {
	if ctx.IsFlagSet(commands.SpecFlag) {
		otherSourceFlags := []string{
			commands.SourceTypeBuildsFlag,
			commands.SourceTypeReleaseBundlesFlag,
			commands.SourceTypeApplicationVersionsFlag,
		}
		for _, flag := range otherSourceFlags {
			if ctx.IsFlagSet(flag) {
				return errorutils.CheckErrorf("--spec provided: all other source flags (e.g., --%s) are not allowed.", flag)
			}
		}
	}
	return nil
}

func validateRequiredFieldsInMap(m map[string]string, requiredFields ...string) error {
	if m == nil {
		return errorutils.CheckErrorf("missing required fields: %v", strings.Join(requiredFields, ", "))
	}
	for _, field := range requiredFields {
		if _, exists := m[field]; !exists {
			return errorutils.CheckErrorf("missing required field: %s", field)
		}
	}
	return nil
}
