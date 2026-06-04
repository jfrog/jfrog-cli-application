package packagecmds

import (
	"os"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/packages"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

// orderedBindPackageKeys defines the display order for package-bind table output.
var orderedBindPackageKeys = []string{
	"application_key",
	"package_type",
	"package_name",
	"package_version",
	"status",
}

type bindPackageCommand struct {
	packageService packages.PackageService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	requestPayload *model.BindPackageRequest
	responseBody   []byte
}

func (bp *bindPackageCommand) Run() error {
	ctx, err := service.NewContext(*bp.serverDetails)
	if err != nil {
		return err
	}
	bp.responseBody, err = bp.packageService.BindPackage(ctx, bp.applicationKey, bp.requestPayload)
	return err
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

	outputFormat, err := ctx.GetOutputFormat()
	if err != nil {
		return err
	}

	if err = commonCLiCommands.Exec(bp); err != nil {
		return err
	}

	return common.PrintResponse(bp.responseBody, outputFormat, os.Stdout, orderedBindPackageKeys)
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
				Description: "Package type (e.g., npm, docker, maven).",
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
		Flags:            commands.GetCommandFlags(commands.PackageBind),
		SupportedFormats: []coreformat.OutputFormat{coreformat.Table, coreformat.Json},
		DefaultFormat:    coreformat.Json,
		Action:           cmd.prepareAndRunCommand,
	}
}
