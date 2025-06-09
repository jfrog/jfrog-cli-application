package packagecmds

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/packages"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

type unbindPackageCommand struct {
	packageService packages.PackageService
	serverDetails  *coreConfig.ServerDetails
	requestPayload *model.BindPackageRequest
}

func (up *unbindPackageCommand) Run() error {
	ctx, err := service.NewContext(*up.serverDetails)
	if err != nil {
		return err
	}
	return up.packageService.UnbindPackage(ctx, up.requestPayload)
}

func (up *unbindPackageCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return up.serverDetails, nil
}

func (up *unbindPackageCommand) CommandName() string {
	return commands.PackageUnbind
}

func (up *unbindPackageCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) < 3 || len(ctx.Arguments) > 4 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	var err error
	up.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	up.requestPayload, err = BuildPackageRequestPayload(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(up)
}

func GetUnbindPackageCommand(appContext app.Context) components.Command {
	cmd := &unbindPackageCommand{packageService: appContext.GetPackageService()}
	return components.Command{
		Name:        commands.PackageUnbind,
		Description: "Unbind packages from an application.",
		Aliases:     []string{"pu"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to unbind the package from.",
			},
			{
				Name:        "package-type",
				Description: "Package type (e.g., npm, docker, maven, generic).",
			},
			{
				Name:        "package-name",
				Description: "Package name.",
			},
			{
				Name:        "package-versions",
				Description: "Comma-separated versions of the package to unbind (e.g., '1.0.0,1.1.0,1.2.0'). If omitted, all versions will be unbound.",
			},
		},
		Flags:  commands.GetCommandFlags(commands.PackageUnbind),
		Action: cmd.prepareAndRunCommand,
	}
}
