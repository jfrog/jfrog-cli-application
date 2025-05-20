package application

import (
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"slices"

	"github.com/jfrog/jfrog-cli-application/application/app"
	"github.com/jfrog/jfrog-cli-application/application/commands"
	"github.com/jfrog/jfrog-cli-application/application/commands/utils"
	"github.com/jfrog/jfrog-cli-application/application/common"
	"github.com/jfrog/jfrog-cli-application/application/model"
	"github.com/jfrog/jfrog-cli-application/application/service"
	"github.com/jfrog/jfrog-cli-application/application/service/applications"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const UpdateApp = "update-app"

type updateAppCommand struct {
	serverDetails      *coreConfig.ServerDetails
	applicationService applications.ApplicationService
	requestBody        *model.AppDescriptor
}

func (uac *updateAppCommand) Run() error {
	ctx, err := service.NewContext(*uac.serverDetails)
	if err != nil {
		return err
	}

	return uac.applicationService.UpdateApplication(ctx, uac.requestBody)
}

func (uac *updateAppCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return uac.serverDetails, nil
}

func (uac *updateAppCommand) CommandName() string {
	return UpdateApp
}

func (uac *updateAppCommand) buildRequestPayload(ctx *components.Context) (*model.AppDescriptor, error) {
	applicationKey := ctx.Arguments[0]
	displayName := ctx.GetStringFlagValue(commands.DisplayNameFlag)

	businessCriticality := ctx.GetStringFlagValue(commands.BusinessCriticalityFlag)
	if businessCriticality != "" && !slices.Contains(model.BusinessCriticalityValues, businessCriticality) {
		return nil, errorutils.CheckErrorf("invalid value for --%s: '%s'. Allowed values: %s", commands.BusinessCriticalityFlag, businessCriticality, coreutils.ListToText(model.BusinessCriticalityValues))
	}

	maturityLevel := ctx.GetStringFlagValue(commands.MaturityLevelFlag)
	if maturityLevel != "" && !slices.Contains(model.MaturityLevelValues, maturityLevel) {
		return nil, errorutils.CheckErrorf("invalid value for --%s: '%s'. Allowed values: %s", commands.MaturityLevelFlag, maturityLevel, coreutils.ListToText(model.MaturityLevelValues))
	}

	description := ctx.GetStringFlagValue(commands.DescriptionFlag)
	userOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.UserOwnersFlag))
	groupOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.GroupOwnersFlag))
	labelsMap, err := utils.ParseMapFlag(ctx.GetStringFlagValue(commands.LabelsFlag))
	if err != nil {
		return nil, err
	}

	return &model.AppDescriptor{
		ApplicationKey:      applicationKey,
		ApplicationName:     displayName,
		Description:         description,
		MaturityLevel:       maturityLevel,
		BusinessCriticality: businessCriticality,
		Labels:              labelsMap,
		UserOwners:          userOwners,
		GroupOwners:         groupOwners,
	}, nil
}

func (uac *updateAppCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	var err error
	uac.requestBody, err = uac.buildRequestPayload(ctx)
	if err != nil {
		return err
	}

	uac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(uac)
}

func GetUpdateAppCommand(appContext app.Context) components.Command {
	cmd := &updateAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        "update",
		Description: "Update an existing application",
		Category:    common.CategoryApplication,
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to update",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.UpdateApp),
		Action: cmd.prepareAndRunCommand,
	}
}
