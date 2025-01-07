package version

import (
	"github.com/jfrog/jfrog-cli-application/application/app"
	"github.com/jfrog/jfrog-cli-application/application/commands"
	"github.com/jfrog/jfrog-cli-application/application/commands/utils"
	"github.com/jfrog/jfrog-cli-application/application/common"
	"github.com/jfrog/jfrog-cli-application/application/model"
	"github.com/jfrog/jfrog-cli-application/application/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type promoteAppVersionCommand struct {
	versionService service.VersionService
	serverDetails  *coreConfig.ServerDetails
	requestPayload *model.PromoteAppVersionRequest
}

func (pv *promoteAppVersionCommand) Run() error {
	ctx := &service.Context{ServerDetails: pv.serverDetails}
	return pv.versionService.PromoteAppVersion(ctx, pv.requestPayload)
}

func (pv *promoteAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return pv.serverDetails, nil
}

func (pv *promoteAppVersionCommand) CommandName() string {
	return commands.PromoteAppVersion
}

func (pv *promoteAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if err := validatePromoteAppVersionContext(ctx); err != nil {
		return err
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

func validatePromoteAppVersionContext(ctx *components.Context) error {
	if show, err := pluginsCommon.ShowCmdHelpIfNeeded(ctx, ctx.Arguments); show || err != nil {
		return err
	}
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}
	return nil
}

func (pv *promoteAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.PromoteAppVersionRequest, error) {
	return &model.PromoteAppVersionRequest{
		ApplicationKey: ctx.GetStringFlagValue(commands.ApplicationKeyFlag),
		Version:        ctx.Arguments[0],
		Environment:    ctx.GetStringFlagValue(commands.EnvironmentVarsFlag),
	}, nil
}

func GetPromoteAppVersionCommand(appContext app.Context) components.Command {
	cmd := &promoteAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.PromoteAppVersion,
		Description: "Promote application version",
		Category:    common.CategoryVersion,
		Aliases:     []string{"pav"},
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
