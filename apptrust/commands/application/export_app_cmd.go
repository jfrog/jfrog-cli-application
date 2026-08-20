package application

import (
	"fmt"
	"os"
	"path/filepath"

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
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type exportAppCommand struct {
	serverDetails      *coreConfig.ServerDetails
	applicationService applications.ApplicationService
	applicationKey     string
	targetPath         string
}

func (eac *exportAppCommand) Run() error {
	ctx, err := service.NewContext(*eac.serverDetails)
	if err != nil {
		return err
	}

	applicationEnvelope, err := eac.applicationService.ExportApplication(ctx, eac.applicationKey)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(eac.targetPath), 0o755); err != nil {
		return errorutils.CheckError(err)
	}
	if err := os.WriteFile(eac.targetPath, applicationEnvelope, 0o644); err != nil {
		return errorutils.CheckError(err)
	}

	log.Info(fmt.Sprintf("Application export written to %s", eac.targetPath))
	return nil
}

func (eac *exportAppCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return eac.serverDetails, nil
}

func (eac *exportAppCommand) CommandName() string {
	return commands.AppExport
}

func (eac *exportAppCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) < 1 || len(ctx.Arguments) > 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	eac.applicationKey = ctx.Arguments[0]
	target := ""
	if len(ctx.Arguments) == 2 {
		target = ctx.Arguments[1]
	}
	eac.targetPath = resolveExportTargetPath(eac.applicationKey, target)

	var err error
	eac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(eac)
}

func resolveExportTargetPath(applicationKey, target string) string {
	if target == "" {
		target = "./"
	}

	// If the target ends with a slash, treat it as a directory and append the default filename
	dir, fileName := fileutils.GetLocalPathAndFile(applicationKey+".json", "", target, true, false)
	return filepath.Join(dir, fileName)
}

func GetExportAppCommand(appContext app.Context) components.Command {
	cmd := &exportAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        commands.AppExport,
		Description: "Export an application to a local JSON file.",
		AIDescription: `Export an application's AppTrust metadata (descriptor, owners, labels, monitor policy, package bindings) to a local JSON file for air-gap transfer.

When to use:
- Copy application metadata between disconnected AppTrust instances.

Prerequisites:
- The application must exist.
- Configured server and project-admin permission on the application's project.
- Export is not available on Edge nodes.

Common patterns:
  $ jf apptrust app-export my-app
  $ jf apptrust app-export my-app ./exports/
  $ jf apptrust app-export my-app ./my-app.json
  $ jf at aexp my-app ./exports/ --server-id=my-server

Gotchas:
- This exports application metadata only, not versions or artifacts.
- Trailing slash on target = directory ({application-key}.json is written inside); no slash = rename to that file.
- An existing file at the target path is overwritten without a prompt.

Related: jf apptrust app-import, jf apptrust app-create, jf apptrust app-delete`,
		Category: common.CategoryApplication,
		Aliases:  []string{"aexp"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to export.",
				Optional:    false,
			},
			{
				Name:        "target pattern",
				Description: "Local filesystem target path. If it ends with a slash, it is assumed to be a directory and {application-key}.json is written into it. If there is no terminal slash, the target path is assumed to be a file to which the export file should be renamed.",
				Optional:    true,
			},
		},
		Flags:  commands.GetCommandFlags(commands.AppExport),
		Action: cmd.prepareAndRunCommand,
	}
}
