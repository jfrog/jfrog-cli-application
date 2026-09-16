package application

import (
	"os"

	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/applications"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

type updateAppCommand struct {
	serverDetails      *coreConfig.ServerDetails
	applicationService applications.ApplicationService
	requestBody        *model.AppDescriptor
	responseBody       []byte
}

func (uac *updateAppCommand) Run() error {
	ctx, err := service.NewContext(*uac.serverDetails)
	if err != nil {
		return err
	}

	uac.responseBody, err = uac.applicationService.UpdateApplication(ctx, uac.requestBody)
	return err
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

	outputFormat, err := ctx.GetOutputFormat()
	if err != nil {
		return err
	}

	if err = commonCLiCommands.Exec(uac); err != nil {
		return err
	}

	return common.PrintResponse(uac.responseBody, outputFormat, os.Stdout, common.OrderedAppKeys)
}

func GetUpdateAppCommand(appContext app.Context) components.Command {
	cmd := &updateAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        commands.AppUpdate,
		Description: "Update an existing application",
		AIDescription: `Update metadata (display name, description, criticality, maturity, labels, owners, auto-promotion stages) of an existing application identified by its key.

When to use:
- Change display attributes (name, description, criticality, maturity) of an existing application.
- Add or remove labels, user owners, or group owners.

Prerequisites:
- The application must already exist (see app-create).
- Configured server and update permission on the application's project.

Common patterns:
  $ jf apptrust app-update my-app --desc="Updated description"
  $ jf apptrust app-update my-app --business-criticality=high --maturity-level=production
  $ jf apptrust app-update my-app --add-labels="env=prod;tier=critical"
  $ jf apptrust app-update my-app --remove-labels="env=staging"
  $ jf apptrust app-update my-app --user-owners="alice;bob" --group-owners="platform-team"
  $ jf apptrust app-update my-app --monitor-policy="type=version_count, value=5"
  $ jf apptrust app-update my-app --auto-promote-stages="DEV;PROD"
  $ jf apptrust app-update my-app --auto-promote-stages=""

Gotchas:
- --labels replaces the full label set; --add-labels and --remove-labels modify incrementally.
- --user-owners / --group-owners take a semicolon-separated list and send exactly the owners you specify; there are no incremental add/remove-owner flags (unlike --add-labels / --remove-labels for labels).
- --monitor-policy takes 'type=<type>[, value=<n>]'. 'value' is required (positive integer) when type is "time_frame_in_months" or "version_count", and must be omitted when type is "none". When --monitor-policy is not provided, the current policy is left unchanged.
- --auto-promote-stages uses semicolon separators and preserves the supplied stage order. Pass an empty value to disable auto-promotion; omit the flag to leave the current stages unchanged.
- Application key cannot be changed; use app-delete and app-create if you need a different key.

Related: jf apptrust app-create, jf apptrust app-delete`,
		Category:         common.CategoryApplication,
		Aliases:          []string{"au"},
		SupportedFormats: []coreformat.OutputFormat{coreformat.Table, coreformat.Json},
		DefaultFormat:    coreformat.Json,
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
