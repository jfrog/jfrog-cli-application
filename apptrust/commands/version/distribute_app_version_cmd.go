package version

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

type distributeAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.DistributeAppVersionRequest
}

func (dv *distributeAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*dv.serverDetails)
	if err != nil {
		return err
	}

	return dv.versionService.DistributeAppVersion(ctx, dv.applicationKey, dv.version, dv.requestPayload)
}

func (dv *distributeAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return dv.serverDetails, nil
}

func (dv *distributeAppVersionCommand) CommandName() string {
	return commands.VersionDistribute
}

func (dv *distributeAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	dv.applicationKey = ctx.Arguments[0]
	dv.version = ctx.Arguments[1]

	if err := ValidateDistributionFlags(ctx); err != nil {
		return err
	}

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	dv.serverDetails = serverDetails

	dv.requestPayload, err = dv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}

	return commonCLiCommands.Exec(dv)
}

func (dv *distributeAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.DistributeAppVersionRequest, error) {
	distributionRules, err := ParseDistributionRules(ctx)
	if err != nil {
		return nil, err
	}

	modifications, err := ParseDistributionModifications(ctx)
	if err != nil {
		return nil, err
	}

	return &model.DistributeAppVersionRequest{
		DistributionRules: distributionRules,
		AutoCreateRepo:    ctx.GetBoolFlagValue(commands.CreateRepoFlag),
		Modifications:     modifications,
	}, nil
}

func GetDistributeAppVersionCommand(appContext app.Context) components.Command {
	cmd := &distributeAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionDistribute,
		Description: "Distribute application version to distribution targets.",
		AIDescription: `Distribute an application version's artifacts to one or more distribution targets according to distribution rules.

When to use:
- Make an application version's artifacts available on distribution targets close to consumers.
- Replicate a released version to remote sites for faster, local access.

Prerequisites:
- The application version must already exist.
- Configured server and distribution permission on the application's project.
- Reachable distribution target(s) matching the provided distribution rules.

Common patterns:
  $ jf apptrust version-distribute my-app 1.0.0
  $ jf apptrust version-distribute my-app 1.0.0 --site="us-*" --country-codes="US;CA"
  $ jf apptrust version-distribute my-app 1.0.0 --dist-rules=/path/to/dist-rules.json
  $ jf apptrust version-distribute my-app 1.0.0 --mapping-pattern="repo/(*)" --mapping-target="target/{1}"
  $ jf apptrust version-distribute my-app 1.0.0 --create-repo

Gotchas:
- Distribution is asynchronous: a successful result means it was triggered, not that it has completed on the distribution targets.
- --dist-rules can't be combined with --site, --city or --country-codes.
- --mapping-pattern and --mapping-target must be provided together.

Related: jf apptrust version-delete-remote, jf apptrust version-release`,
		Category: common.CategoryVersion,
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The application key.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to distribute.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionDistribute),
		Action: cmd.prepareAndRunCommand,
	}
}
