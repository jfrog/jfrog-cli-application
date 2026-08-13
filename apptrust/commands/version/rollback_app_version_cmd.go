package version

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"os"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/versions"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

// orderedRollbackAppVersionKeys defines the display order for version-rollback table output.
var orderedRollbackAppVersionKeys = []string{
	"application_key",
	"version",
	"project_key",
	"rollback_from_stage",
	"rollback_to_stage",
}

type rollbackAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.RollbackAppVersionRequest
	fromStage      string
	sync           bool
	responseBody   []byte
}

func (rv *rollbackAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*rv.serverDetails)
	if err != nil {
		return err
	}

	rv.responseBody, err = rv.versionService.RollbackAppVersion(ctx, rv.applicationKey, rv.version, rv.requestPayload, rv.sync)
	return err
}

func (rv *rollbackAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return rv.serverDetails, nil
}

func (rv *rollbackAppVersionCommand) CommandName() string {
	return commands.VersionRollback
}

func (rv *rollbackAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 3 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	rv.applicationKey = ctx.Arguments[0]
	rv.version = ctx.Arguments[1]
	rv.fromStage = ctx.Arguments[2]

	rv.sync = ctx.GetBoolTFlagValue(commands.SyncFlag)

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	rv.serverDetails = serverDetails
	rv.requestPayload = model.NewRollbackAppVersionRequest(rv.fromStage)

	outputFormat, err := ctx.GetOutputFormat()
	if err != nil {
		return err
	}

	if err = commonCLiCommands.Exec(rv); err != nil {
		return err
	}

	return common.PrintResponse(rv.responseBody, outputFormat, os.Stdout, orderedRollbackAppVersionKeys)
}

func GetRollbackAppVersionCommand(appContext app.Context) components.Command {
	cmd := &rollbackAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionRollback,
		Description: "Roll back application version promotion.",
		AIDescription: `Roll back a previous promotion of an application version from a given stage, reverting artifact placement performed by that promotion.

When to use:
- Undo a recent promotion that introduced regressions or wrong content.
- Revert a release that needs to be retracted from a stage.

Prerequisites:
- The application version must currently be promoted to the specified from-stage.
- Configured server and rollback permission on the application's project.

Common patterns:
  $ jf apptrust version-rollback my-app 1.0.0 PROD
  $ jf apptrust version-rollback my-app 1.0.0 STAGING --sync=false
  $ jf at vrb my-app 1.0.0 QA --server-id=my-server

Gotchas:
- All three positional arguments (application-key, version, from-stage) are mandatory.
- Rollback acts on the most recent promotion to from-stage; it does not delete the version itself (use version-delete for that).

Related: jf apptrust version-promote, jf apptrust version-delete`,
		Category: common.CategoryVersion,
		Aliases:  []string{"vrb"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The application key.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to roll back.",
				Optional:    false,
			},
			{
				Name:        "from-stage",
				Description: "The name of the stage from which to roll back the application version.",
				Optional:    false,
			},
		},
		Flags:            commands.GetCommandFlags(commands.VersionRollback),
		SupportedFormats: []coreformat.OutputFormat{coreformat.Table, coreformat.Json},
		DefaultFormat:    coreformat.Json,
		Action:           cmd.prepareAndRunCommand,
	}
}
