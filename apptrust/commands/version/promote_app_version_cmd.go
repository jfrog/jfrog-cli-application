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
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type promoteAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	requestPayload *model.PromoteAppVersionRequest
}

func (pv *promoteAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*pv.serverDetails)
	if err != nil {
		return err
	}

	return pv.versionService.PromoteAppVersion(ctx, pv.requestPayload)
}

func (pv *promoteAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return pv.serverDetails, nil
}

func (pv *promoteAppVersionCommand) CommandName() string {
	return commands.PromoteAppVersion
}

func (pv *promoteAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}
	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	pv.serverDetails = serverDetails
	pv.requestPayload, err = pv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}
	return commonCLiCommands.Exec(pv)
}

func (pv *promoteAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.PromoteAppVersionRequest, error) {
	return &model.PromoteAppVersionRequest{
		ApplicationKey: ctx.GetStringFlagValue(commands.ApplicationKeyFlag),
		Version:        ctx.Arguments[0],
		Environment:    ctx.GetStringFlagValue(commands.StageVarsFlag),
	}, nil
}

func GetPromoteAppVersionCommand(appContext app.Context) components.Command {
	cmd := &promoteAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.PromoteAppVersion,
		Description: "Promote application version",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vp"},
		Arguments: []components.Argument{
			{
				Name:        "version-name",
				Description: "The name of the version",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.PromoteAppVersion),
		Action: cmd.prepareAndRunCommand,
	}
}
