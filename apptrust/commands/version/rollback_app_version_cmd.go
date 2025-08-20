package version

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

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
)

type rollbackAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.RollbackAppVersionRequest
	fromStage      string
}

func (rv *rollbackAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*rv.serverDetails)
	if err != nil {
		return err
	}

	return rv.versionService.RollbackAppVersion(ctx, rv.applicationKey, rv.version, rv.requestPayload)
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

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	rv.serverDetails = serverDetails
	rv.requestPayload = model.NewRollbackAppVersionRequest(rv.fromStage)

	return commonCLiCommands.Exec(rv)
}

func GetRollbackAppVersionCommand(appContext app.Context) components.Command {
	cmd := &rollbackAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionRollback,
		Description: "Roll back application version promotion.",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vrb"},
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
		},
		Flags:  commands.GetCommandFlags(commands.VersionRollback),
		Action: cmd.prepareAndRunCommand,
	}
}
