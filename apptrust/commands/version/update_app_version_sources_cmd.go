package version

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/versions"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type updateAppVersionSourcesCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.UpdateVersionSourcesRequest
	sync           bool
	dryRun         bool
	failFast       bool
}

func (cmd *updateAppVersionSourcesCommand) Run() error {
	ctx, err := service.NewContext(*cmd.serverDetails)
	if err != nil {
		log.Error("Failed to create service context:", err)
		return err
	}

	err = cmd.versionService.UpdateAppVersionSources(ctx, cmd.applicationKey, cmd.version, cmd.requestPayload, cmd.sync, cmd.dryRun, cmd.failFast)
	if err != nil {
		log.Error("Failed to update application version sources:", err)
		return err
	}

	return nil
}

func (cmd *updateAppVersionSourcesCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return cmd.serverDetails, nil
}

func (cmd *updateAppVersionSourcesCommand) CommandName() string {
	return commands.VersionUpdateSources
}

func (cmd *updateAppVersionSourcesCommand) prepareAndRunCommand(ctx *components.Context) error {
	if err := validateUpdateSourcesContext(ctx); err != nil {
		return err
	}

	if err := cmd.parseFlagsAndSetFields(ctx); err != nil {
		return err
	}

	var err error
	cmd.requestPayload, err = cmd.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}

	return commonCLiCommands.Exec(cmd)
}

func validateUpdateSourcesContext(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}
	if err := validateNoSpecAndFlagsTogether(ctx); err != nil {
		return err
	}
	return validateAtLeastOneSourceFlag(ctx)
}

// parseFlagsAndSetFields parses CLI flags and sets struct fields accordingly.
func (cmd *updateAppVersionSourcesCommand) parseFlagsAndSetFields(ctx *components.Context) error {
	cmd.applicationKey = ctx.Arguments[0]
	cmd.version = ctx.Arguments[1]

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	cmd.serverDetails = serverDetails

	cmd.sync = ctx.GetBoolTFlagValue(commands.SyncFlag)
	cmd.dryRun = ctx.GetBoolFlagValue(commands.DryRunFlag)
	cmd.failFast = ctx.GetBoolTFlagValue(commands.FailFastFlag)

	return nil
}

func (cmd *updateAppVersionSourcesCommand) buildRequestPayload(ctx *components.Context) (*model.UpdateVersionSourcesRequest, error) {
	sources, filters, err := buildSourcesAndFiltersFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &model.UpdateVersionSourcesRequest{
		AddSources: sources,
		Filters:    filters,
	}, nil
}

func GetUpdateAppVersionSourcesCommand(appContext app.Context) components.Command {
	cmd := &updateAppVersionSourcesCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.VersionUpdateSources,
		Description: "Updates the sources for a draft application version.",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vus"},
		Arguments: []components.Argument{
			{
				Name:        "app-key",
				Description: "The application key of the application for which the version sources are being updated.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version number (in SemVer format) for the application version to update sources.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionUpdateSources),
		Action: cmd.prepareAndRunCommand,
	}
}
