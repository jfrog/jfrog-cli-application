package application

import (
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/applications"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

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
	return commands.AppUpdate
}

func (uac *updateAppCommand) buildRequestPayload(ctx *components.Context) (*model.AppDescriptor, error) {
	applicationKey := ctx.Arguments[0]

	descriptor := &model.AppDescriptor{
		ApplicationKey: applicationKey,
	}

	err := populateApplicationFromFlags(ctx, descriptor)
	if err != nil {
		return nil, err
	}

	return descriptor, nil
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
		Name:        commands.AppUpdate,
		Description: "Update an existing application",
		Category:    common.CategoryApplication,
		Aliases:     []string{"au"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to update",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.AppUpdate),
		Action: cmd.prepareAndRunCommand,
	}
}
