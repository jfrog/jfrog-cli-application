package application

import (
	"slices"

	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/application/commands/utils"
	"github.com/jfrog/jfrog-cli-application/application/model"
	"github.com/jfrog/jfrog-cli-application/application/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"

	"github.com/jfrog/jfrog-cli-application/application/app"
	"github.com/jfrog/jfrog-cli-application/application/commands"
	"github.com/jfrog/jfrog-cli-application/application/common"
	"github.com/jfrog/jfrog-cli-application/application/service/applications"
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
	appKey := ctx.Arguments[0]
	displayName := ctx.GetStringFlagValue(commands.DisplayNameFlag)
	if displayName == "" {
		// Default to the application key if display name is not provided
		displayName = appKey
	}

	project := ctx.GetStringFlagValue(commands.ProjectFlag)
	if project == "" {
		return nil, errorutils.CheckErrorf("--%s is mandatory", commands.ProjectFlag)
	}

	businessCriticality := ctx.GetStringFlagValue(commands.BusinessCriticalityFlag)
	if businessCriticality == "" {
		// Default to "unspecified" if not provided
		businessCriticality = model.BusinessCriticalityValues[0]
	} else if !slices.Contains(model.BusinessCriticalityValues, businessCriticality) {
		return nil, errorutils.CheckErrorf("invalid value for --%s: '%s'. Allowed values: %s", commands.BusinessCriticalityFlag, businessCriticality, coreutils.ListToText(model.BusinessCriticalityValues))
	}

	maturityLevel := ctx.GetStringFlagValue(commands.MaturityLevelFlag)
	if maturityLevel == "" {
		// Default to "unspecified" if not provided
		maturityLevel = model.MaturityLevelValues[0]
	} else if !slices.Contains(model.MaturityLevelValues, maturityLevel) {
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
		ApplicationName:     displayName,
		ApplicationKey:      appKey,
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
