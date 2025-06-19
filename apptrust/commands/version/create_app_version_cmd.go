package version

import (
	"encoding/json"
	"strings"
	"time"

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
	signingKey     string
	sync           bool
}

type createVersionSpec struct {
	Packages       []model.CreateVersionPackage       `json:"packages,omitempty"`
	Builds         []model.CreateVersionBuild         `json:"builds,omitempty"`
	ReleaseBundles []model.CreateVersionReleaseBundle `json:"release_bundles,omitempty"`
	Versions       []model.CreateVersionReference     `json:"versions,omitempty"`
	Exclude        []model.ExcludePackage             `json:"exclude,omitempty"`
}

func (cv *createAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*cv.serverDetails)
	if err != nil {
		return err
	}

	return cv.versionService.CreateAppVersion(ctx, cv.requestPayload, cv.signingKey, cv.sync)
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

	cv.signingKey = ctx.GetStringFlagValue(commands.SigningKeyFlag)
	cv.sync = ctx.GetBoolFlagValue(commands.SyncFlag)

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
	sources := &model.CreateVersionSources{}
	var err error

	// Handle spec file if provided
	if ctx.IsFlagSet(commands.SpecFlag) {
		sources, err = cv.loadFromSpec(ctx)
		if errorutils.CheckError(err) != nil {
			return nil, err
		}
	} else {
		if ctx.IsFlagSet(commands.PackageNameFlag) {
			sources.Packages = append(sources.Packages, model.CreateVersionPackage{
				Type:       ctx.GetStringFlagValue(commands.PackageTypeFlag),
				Name:       ctx.GetStringFlagValue(commands.PackageNameFlag),
				Version:    ctx.GetStringFlagValue(commands.PackageVersionFlag),
				Repository: ctx.GetStringFlagValue(commands.PackageRepositoryFlag),
			})
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
		if excludeStr := ctx.GetStringFlagValue(commands.ExcludeFlag); excludeStr != "" {
			excludedPackages, err := cv.parseExcludedPackages(excludeStr)
			if err != nil {
				return nil, err
			}
			sources.Exclude = excludedPackages
		}
	}

	// Only include sources in the request if any sources were provided
	var sourcesPointer *model.CreateVersionSources
	if len(sources.Packages) > 0 || len(sources.Builds) > 0 ||
		len(sources.ReleaseBundles) > 0 || len(sources.Versions) > 0 || len(sources.Exclude) > 0 {
		sourcesPointer = sources
	}

	return &model.CreateAppVersionRequest{
		ApplicationKey: ctx.GetStringFlagValue(commands.ApplicationKeyFlag),
		Version:        ctx.Arguments[1],
		Sources:        sourcesPointer,
		Tag:            ctx.GetStringFlagValue(commands.TagFlag),
	}, nil
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

	sources := &model.CreateVersionSources{
		Packages:       spec.Packages,
		Builds:         spec.Builds,
		ReleaseBundles: spec.ReleaseBundles,
		Versions:       spec.Versions,
		Exclude:        spec.Exclude,
	}

	return sources, nil
}

// Helper method to parse builds string format: "name1:number1[:timestamp1];name2:number2[:timestamp2]"
func (cv *createAppVersionCommand) parseBuilds(buildsStr string) ([]model.CreateVersionBuild, error) {
	var builds []model.CreateVersionBuild
	buildEntries := strings.Split(buildsStr, ";")

	for _, entry := range buildEntries {
		parts := strings.Split(entry, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return nil, errorutils.CheckErrorf("invalid build format: '%s'. Expected format: name:number[:timestamp]", entry)
		}

		build := model.CreateVersionBuild{
			Name:   parts[0],
			Number: parts[1],
		}

		// Add timestamp if provided (optional)
		if len(parts) == 3 {
			build.Started = parts[2]

			// Validate timestamp format
			_, err := time.Parse(time.RFC3339, build.Started)
			if err != nil {
				return nil, errorutils.CheckErrorf("invalid timestamp format for build '%s': %s. Expected RFC3339 format (e.g., 2006-01-02T15:04:05Z)", build.Name, err.Error())
			}
		}

		builds = append(builds, build)
	}

	return builds, nil
}

// Helper method to parse release bundles string format: "name1:version1;name2:version2"
func (cv *createAppVersionCommand) parseReleaseBundles(rbStr string) ([]model.CreateVersionReleaseBundle, error) {
	var releaseBundles []model.CreateVersionReleaseBundle
	rbEntries := strings.Split(rbStr, ";")

	for _, entry := range rbEntries {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return nil, errorutils.CheckErrorf("invalid release bundle format: '%s'. Expected format: name:version", entry)
		}

		rb := model.CreateVersionReleaseBundle{
			Name:    parts[0],
			Version: parts[1],
		}

		releaseBundles = append(releaseBundles, rb)
	}

	return releaseBundles, nil
}

// Helper method to parse source versions string format: "app1:version1;app2:version2"
func (cv *createAppVersionCommand) parseSourceVersions(sourceVersionsStr string) ([]model.CreateVersionReference, error) {
	var sourceVersions []model.CreateVersionReference
	versionEntries := strings.Split(sourceVersionsStr, ";")

	for _, entry := range versionEntries {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return nil, errorutils.CheckErrorf("invalid source version format: '%s'. Expected format: app-key:version", entry)
		}

		sv := model.CreateVersionReference{
			ApplicationKey: parts[0],
			Version:        parts[1],
		}

		sourceVersions = append(sourceVersions, sv)
	}

	return sourceVersions, nil
}

// Helper method to parse excluded packages string format: "name1:version1;name2:version2"
func (cv *createAppVersionCommand) parseExcludedPackages(excludeStr string) ([]model.ExcludePackage, error) {
	var excludedPackages []model.ExcludePackage
	packageEntries := strings.Split(excludeStr, ";")

	for _, entry := range packageEntries {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return nil, errorutils.CheckErrorf("invalid exclude package format: '%s'. Expected format: name:version", entry)
		}

		pkg := model.ExcludePackage{
			Name:    parts[0],
			Version: parts[1],
		}

		excludedPackages = append(excludedPackages, pkg)
	}

	return excludedPackages, nil
}

// Updated validation logic to support multiple source types
func validateCreateAppVersionContext(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	hasSource := ctx.IsFlagSet(commands.SpecFlag) ||
		ctx.IsFlagSet(commands.PackageNameFlag) ||
		ctx.IsFlagSet(commands.BuildsFlag) ||
		ctx.IsFlagSet(commands.ReleaseBundlesFlag) ||
		ctx.IsFlagSet(commands.SourceVersionFlag)

	if !hasSource {
		return errorutils.CheckErrorf("Missing source information. Please provide at least one source: --%s, --%s, --%s, --%s, or --%s",
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
				Name:        "application-key",
				Description: "The application key.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to create (must follow SemVer format).",
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
