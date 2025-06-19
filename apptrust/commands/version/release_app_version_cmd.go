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

type releaseAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.ReleaseAppVersionRequest
	sync           bool
}

func (rv *releaseAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*rv.serverDetails)
	if err != nil {
		return err
	}

	return rv.versionService.ReleaseAppVersion(ctx, rv.applicationKey, rv.version, rv.requestPayload, rv.sync)
}

func (rv *releaseAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return rv.serverDetails, nil
}

func (rv *releaseAppVersionCommand) CommandName() string {
	return commands.VersionRelease
}

func (rv *releaseAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	// Extract from arguments
	rv.applicationKey = ctx.Arguments[0]
	rv.version = ctx.Arguments[1]

	// Extract sync flag value
	rv.sync = ctx.GetBoolTFlagValue(commands.SyncFlag)

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	rv.serverDetails = serverDetails
	rv.requestPayload, err = rv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}
	return commonCLiCommands.Exec(rv)
}

func (rv *releaseAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.ReleaseAppVersionRequest, error) {
	promotionType, includedRepos, excludedRepos, err := BuildPromotionParams(ctx)
	if err != nil {
		return nil, err
	}

	artifactProps, err := ParseArtifactProps(ctx)
	if err != nil {
		return nil, err
	}

	// Create the release request
	return model.NewReleaseAppVersionRequest(
		promotionType,
		includedRepos,
		excludedRepos,
		artifactProps,
	), nil
}

func GetReleaseAppVersionCommand(appContext app.Context) components.Command {
	cmd := &releaseAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionRelease,
		Description: "Release application version.",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vr"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The application key.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to release.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionRelease),
		Action: cmd.prepareAndRunCommand,
	}
}
