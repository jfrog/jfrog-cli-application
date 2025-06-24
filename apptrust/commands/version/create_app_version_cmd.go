package version

import (
	"encoding/json"

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
	if ctx.IsFlagSet(commands.PackagesFlag) {
		parsedPkgs, err := utils.ParsePackagesFlag(ctx.GetStringFlagValue(commands.PackagesFlag))
		if err != nil {
			return nil, err
		}
		for _, pkg := range parsedPkgs {
			sources.Packages = append(sources.Packages, model.CreateVersionPackage{
				Type:       ctx.GetStringFlagValue(commands.PackageTypeFlag),
				Name:       pkg["name"],
				Version:    pkg["version"],
				Repository: ctx.GetStringFlagValue(commands.PackageRepositoryFlag),
			})
		}
	}
	if buildsStr := ctx.GetStringFlagValue(commands.BuildsFlag); buildsStr != "" {
		builds, err := cv.parseBuilds(buildsStr)
		if err != nil {
			return nil, err
		}
		sources.Builds = builds
	}
	if rbStr := ctx.GetStringFlagValue(commands.ReleaseBundlesFlag); rbStr != "" {
		releaseBundles, err := cv.parseReleaseBundles(rbStr)
		if err != nil {
			return nil, err
		}
		sources.ReleaseBundles = releaseBundles
	}
	if srcVersionsStr := ctx.GetStringFlagValue(commands.SourceVersionFlag); srcVersionsStr != "" {
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
	var builds []model.CreateVersionBuild
	for _, parts := range utils.ParseDelimitedSlice(buildsStr) {
		if len(parts) < 2 || len(parts) > 3 {
			return nil, errorutils.CheckErrorf("invalid build format: %v", parts)
		}
		build := model.CreateVersionBuild{
			Name:   parts[0],
			Number: parts[1],
		}
		if len(parts) == 3 {
			build.Started = parts[2]
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (cv *createAppVersionCommand) parseReleaseBundles(rbStr string) ([]model.CreateVersionReleaseBundle, error) {
	pairs, err := utils.ParseNameVersionPairs(rbStr)
	if err != nil {
		return nil, errorutils.CheckErrorf("invalid release bundle format: %v", err)
	}
	var bundles []model.CreateVersionReleaseBundle
	for _, pair := range pairs {
		bundles = append(bundles, model.CreateVersionReleaseBundle{
			Name:    pair[0],
			Version: pair[1],
		})
	}
	return bundles, nil
}

func (cv *createAppVersionCommand) parseSourceVersions(sourceVersionsStr string) ([]model.CreateVersionReference, error) {
	pairs, err := utils.ParseNameVersionPairs(sourceVersionsStr)
	if err != nil {
		return nil, errorutils.CheckErrorf("invalid source version format: %v", err)
	}
	var refs []model.CreateVersionReference
	for _, pair := range pairs {
		refs = append(refs, model.CreateVersionReference{
			ApplicationKey: pair[0],
			Version:        pair[1],
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
		ctx.IsFlagSet(commands.PackageNameFlag) ||
		ctx.IsFlagSet(commands.BuildsFlag) ||
		ctx.IsFlagSet(commands.ReleaseBundlesFlag) ||
		ctx.IsFlagSet(commands.SourceVersionFlag)

	if !hasSource {
		return errorutils.CheckErrorf(
			"At least one source flag is required to create an application version. Please provide one of the following: --%s, --%s, --%s, --%s, or --%s.",
			commands.SpecFlag, commands.PackageNameFlag, commands.BuildsFlag, commands.ReleaseBundlesFlag, commands.SourceVersionFlag)
	}

	// Validate package details if used
	if ctx.IsFlagSet(commands.PackageNameFlag) {
		if !ctx.IsFlagSet(commands.PackageVersionFlag) || !ctx.IsFlagSet(commands.PackageRepositoryFlag) {
			return handleMissingPackageDetailsError()
		}
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

func handleMissingPackageDetailsError() error {
	return errorutils.CheckErrorf("Missing packages information. Please provide the following flags --%s or the set of: --%s, --%s, --%s, --%s",
		commands.SpecFlag, commands.PackageTypeFlag, commands.PackageNameFlag, commands.PackageVersionFlag, commands.PackageRepositoryFlag)
}

// Returns error if both --spec and any other source flag are set
func validateNoSpecAndFlagsTogether(ctx *components.Context) error {
	if ctx.IsFlagSet(commands.SpecFlag) {
		otherSourceFlags := []string{
			commands.PackageNameFlag,
			commands.BuildsFlag,
			commands.ReleaseBundlesFlag,
			commands.SourceVersionFlag,
		}
		for _, flag := range otherSourceFlags {
			if ctx.IsFlagSet(flag) {
				return errorutils.CheckErrorf("--spec provided: all other source flags (e.g., --%s) are not allowed.", flag)
			}
		}
	}
	return nil
}
