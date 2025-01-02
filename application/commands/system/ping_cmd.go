package system

import (
	"github.com/jfrog/jfrog-cli-application/application/app"
	"github.com/jfrog/jfrog-cli-application/application/commands"
	"github.com/jfrog/jfrog-cli-application/application/commands/utils"
	"github.com/jfrog/jfrog-cli-application/application/common"
	"github.com/jfrog/jfrog-cli-application/application/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

type pingCommand struct {
	systemService service.SystemService
	serverDetails *coreConfig.ServerDetails
}

func (pc *pingCommand) Run() error {
	ctx := &service.Context{ServerDetails: pc.serverDetails}
	return pc.systemService.Ping(ctx)
}

func (pc *pingCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return pc.serverDetails, nil
}

func (pc *pingCommand) CommandName() string {
	return commands.Ping
}

func (pc *pingCommand) prepareAndRunCommand(ctx *components.Context) error {
	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	pc.serverDetails = serverDetails
	return commonCLiCommands.Exec(pc)
}

func GetPingCommand(appContext app.Context) components.Command {
	cmd := &pingCommand{systemService: appContext.GetSystemService()}
	return components.Command{
		Name:        commands.Ping,
		Description: "Ping the application server",
		Category:    common.CategorySystem,
		Aliases:     []string{"p"},
		Arguments:   []components.Argument{},
		Flags:       commands.GetCommandFlags(commands.Ping),
		Action:      cmd.prepareAndRunCommand,
	}
}
