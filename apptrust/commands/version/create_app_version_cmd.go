package version

import (
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
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type createAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	requestPayload *model.CreateAppVersionRequest
	sync           bool
}

func (cv *createAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*cv.serverDetails)
	if err != nil {
		return err
	}

	return cv.versionService.CreateAppVersion(ctx, cv.requestPayload, cv.sync)
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
	cv.sync = ctx.GetBoolTFlagValue(commands.SyncFlag)
	cv.requestPayload, err = cv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}
	return commonCLiCommands.Exec(cv)
}

func (cv *createAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.CreateAppVersionRequest, error) {
	sources, filters, err := buildSourcesAndFiltersFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &model.CreateAppVersionRequest{
		ApplicationKey: ctx.Arguments[0],
		Version:        ctx.Arguments[1],
		Sources:        sources,
		Tag:            ctx.GetStringFlagValue(commands.TagFlag),
		Draft:          ctx.GetBoolFlagValue(commands.DraftFlag),
		Filters:        filters,
	}, nil
}

func validateCreateAppVersionContext(ctx *components.Context) error {
	if err := validateNoSpecAndFlagsTogether(ctx); err != nil {
		return err
	}
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	if !hasSourceFlags(ctx) {
		return errorutils.CheckErrorf(
			"At least one source flag is required to create an application version. Please provide --%s or at least one of the following: --%s, --%s, --%s, --%s, --%s.",
			commands.SpecFlag, commands.SourceTypeBuildsFlag, commands.SourceTypeReleaseBundlesFlag, commands.SourceTypeApplicationVersionsFlag, commands.SourceTypePackagesFlag, commands.SourceTypeArtifactsFlag)
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

// Returns error if both --spec and any other source flag or filter flag are set
func validateNoSpecAndFlagsTogether(ctx *components.Context) error {
	if ctx.IsFlagSet(commands.SpecFlag) {
		otherSourceFlags := []string{
			commands.SourceTypeBuildsFlag,
			commands.SourceTypeReleaseBundlesFlag,
			commands.SourceTypeApplicationVersionsFlag,
			commands.SourceTypePackagesFlag,
			commands.SourceTypeArtifactsFlag,
		}
		for _, flag := range otherSourceFlags {
			if ctx.IsFlagSet(flag) {
				return errorutils.CheckErrorf("--spec provided: all other source flags (e.g., --%s) are not allowed.", flag)
			}
		}
		if ctx.IsFlagSet(commands.IncludeFilterFlag) {
			return errorutils.CheckErrorf("--spec provided: filter flags (e.g., --%s) are not allowed.", commands.IncludeFilterFlag)
		}
		if ctx.IsFlagSet(commands.ExcludeFilterFlag) {
			return errorutils.CheckErrorf("--spec provided: filter flags (e.g., --%s) are not allowed.", commands.ExcludeFilterFlag)
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
