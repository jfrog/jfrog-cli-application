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
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type importAppCommand struct {
	serverDetails       *coreConfig.ServerDetails
	applicationService  applications.ApplicationService
	applicationEnvelope []byte
}

func (iac *importAppCommand) Run() error {
	ctx, err := service.NewContext(*iac.serverDetails)
	if err != nil {
		return err
	}

	return iac.applicationService.ImportApplication(ctx, iac.applicationEnvelope)
}

func (iac *importAppCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return iac.serverDetails, nil
}

func (iac *importAppCommand) CommandName() string {
	return commands.AppImport
}

func (iac *importAppCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	content, err := fileutils.ReadFile(ctx.Arguments[0])
	if errorutils.CheckError(err) != nil {
		return err
	}
	iac.applicationEnvelope = content

	iac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(iac)
}

func GetImportAppCommand(appContext app.Context) components.Command {
	cmd := &importAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        commands.AppImport,
		Description: "Import an application from a local file.",
		AIDescription: `Import application metadata (descriptor, owners, labels, monitor policy, package bindings) from a local export file into AppTrust.

When to use:
- Restore or copy application metadata onto an AppTrust instance after jf apptrust app-export.

Prerequisites:
- A file produced by app-export.
- Configured server and project-admin permission on the project's key in the file.

Common patterns:
  $ jf apptrust app-import ./my-app
  $ jf at aimp ./my-app.json --server-id=my-server

Gotchas:
- This imports application metadata only, not versions or artifacts.
- Import upserts the application in place; a mismatched projectKey in the file is rejected by the server.

Related: jf apptrust app-export, jf apptrust app-create, jf apptrust app-delete`,
		Category: common.CategoryApplication,
		Aliases:  []string{"aimp"},
		Arguments: []components.Argument{
			{
				Name:        "path to file",
				Description: "Local path to the application export file.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.AppImport),
		Action: cmd.prepareAndRunCommand,
	}
}
