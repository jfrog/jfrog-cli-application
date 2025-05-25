package application

import (
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/applications"
)

type createAppCommand struct {
	serverDetails      *coreConfig.ServerDetails
	applicationService applications.ApplicationService
	requestBody        *model.AppDescriptor
}

func (cac *createAppCommand) Run() error {
	ctx, err := service.NewContext(*cac.serverDetails)
	if err != nil {
		return err
	}

	return cac.applicationService.CreateApplication(ctx, cac.requestBody)
}

func (cac *createAppCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return cac.serverDetails, nil
}

func (cac *createAppCommand) CommandName() string {
	return commands.CreateApp
}

func (cac *createAppCommand) buildRequestPayload(ctx *components.Context) (*model.AppDescriptor, error) {
	applicationKey := ctx.Arguments[0]
	applicationName := ctx.GetStringFlagValue(commands.ApplicationNameFlag)
	if applicationName == "" {
		// Default to the application key if application name is not provided
		applicationName = applicationKey
	}

	project := ctx.GetStringFlagValue(commands.ProjectFlag)
	if project == "" {
		return nil, errorutils.CheckErrorf("--%s is mandatory", commands.ProjectFlag)
	}

	businessCriticalityStr := ctx.GetStringFlagValue(commands.BusinessCriticalityFlag)
	businessCriticality, err := utils.ValidateEnumFlag(
		commands.BusinessCriticalityFlag,
		businessCriticalityStr,
		model.BusinessCriticalityUnspecified,
		model.BusinessCriticalityValues)
	if err != nil {
		return nil, err
	}

	maturityLevelStr := ctx.GetStringFlagValue(commands.MaturityLevelFlag)
	maturityLevel, err := utils.ValidateEnumFlag(
		commands.MaturityLevelFlag,
		maturityLevelStr,
		model.MaturityLevelUnspecified,
		model.MaturityLevelValues)
	if err != nil {
		return nil, err
	}

	description := ctx.GetStringFlagValue(commands.DescriptionFlag)
	userOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.UserOwnersFlag))
	groupOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.GroupOwnersFlag))
	labelsMap, err := utils.ParseMapFlag(ctx.GetStringFlagValue(commands.LabelsFlag))
	if err != nil {
		return nil, err
	}

	return &model.AppDescriptor{
		ApplicationName:     applicationName,
		ApplicationKey:      applicationKey,
		Description:         description,
		ProjectKey:          project,
		MaturityLevel:       maturityLevel,
		BusinessCriticality: businessCriticality,
		Labels:              labelsMap,
		UserOwners:          userOwners,
		GroupOwners:         groupOwners,
	}, nil
}

func (cac *createAppCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	var err error
	cac.requestBody, err = cac.buildRequestPayload(ctx)
	if err != nil {
		return err
	}

	cac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(cac)
}

func GetCreateAppCommand(appContext app.Context) components.Command {
	cmd := &createAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        "create",
		Description: "Create a new application",
		Category:    common.CategoryApplication,
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to create",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.CreateApp),
		Action: cmd.prepareAndRunCommand,
	}
}
