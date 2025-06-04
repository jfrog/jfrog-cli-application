package application

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/applications"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

type deleteAppCommand struct {
	serverDetails      *coreConfig.ServerDetails
	applicationService applications.ApplicationService
	applicationKey     string
}

func (dac *deleteAppCommand) Run() error {
	ctx, err := service.NewContext(*dac.serverDetails)
	if err != nil {
		return err
	}

	return dac.applicationService.DeleteApplication(ctx, dac.applicationKey)
}

func (dac *deleteAppCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return dac.serverDetails, nil
}

func (dac *deleteAppCommand) CommandName() string {
	return commands.DeleteApp
}

func (dac *deleteAppCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	dac.applicationKey = ctx.Arguments[0]

	var err error
	dac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(dac)
}

func GetDeleteAppCommand(appContext app.Context) components.Command {
	cmd := &deleteAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        "delete",
		Description: "Delete an application",
		Category:    common.CategoryApplication,
		Aliases:     []string{"d"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to delete",
				Optional:    false,
			},
		},
		Action: cmd.prepareAndRunCommand,
	}
}
