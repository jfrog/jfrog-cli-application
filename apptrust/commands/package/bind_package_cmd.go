package packagecmds

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/packages"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

type bindPackageCommand struct {
	packageService packages.PackageService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	requestPayload *model.BindPackageRequest
}

func (bp *bindPackageCommand) Run() error {
	ctx, err := service.NewContext(*bp.serverDetails)
	if err != nil {
		return err
	}
	return bp.packageService.BindPackage(ctx, bp.applicationKey, bp.requestPayload)
}

func (bp *bindPackageCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return bp.serverDetails, nil
}

func (bp *bindPackageCommand) CommandName() string {
	return commands.PackageBind
}

func (bp *bindPackageCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 4 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	var err error
	bp.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	bp.extractFromArgs(ctx)

	return commonCLiCommands.Exec(bp)
}

func (bp *bindPackageCommand) extractFromArgs(ctx *components.Context) {
	bp.applicationKey = ctx.Arguments[0]
	packageType := ctx.Arguments[1]
	packageName := ctx.Arguments[2]
	version := ctx.Arguments[3]

	bp.requestPayload = &model.BindPackageRequest{
		Type:    packageType,
		Name:    packageName,
		Version: version,
	}
}

func GetBindPackageCommand(appContext app.Context) components.Command {
	cmd := &bindPackageCommand{packageService: appContext.GetPackageService()}
	return components.Command{
		Name:        commands.PackageBind,
		Description: "Bind packages to an application.",
		Category:    common.CategoryPackage,
		Aliases:     []string{"pb"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to bind the package to.",
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
				Name:        "package-version",
				Description: "Package version.",
			},
		},
		Flags:  commands.GetCommandFlags(commands.PackageBind),
		Action: cmd.prepareAndRunCommand,
	}
}
