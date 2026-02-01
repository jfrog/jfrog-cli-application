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
	applicationKey string
	version        string
	requestPayload *model.PromoteAppVersionRequest
	sync           bool
}

func (pv *promoteAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*pv.serverDetails)
	if err != nil {
		return err
	}

	return pv.versionService.PromoteAppVersion(ctx, pv.applicationKey, pv.version, pv.requestPayload, pv.sync)
}

func (pv *promoteAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return pv.serverDetails, nil
}

func (pv *promoteAppVersionCommand) CommandName() string {
	return commands.VersionPromote
}

func (pv *promoteAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 3 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	// Extract from arguments
	pv.applicationKey = ctx.Arguments[0]
	pv.version = ctx.Arguments[1]

	// Extract sync flag value
	pv.sync = ctx.GetBoolTFlagValue(commands.SyncFlag)

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
	stage := ctx.Arguments[2]

	promotionType, includedRepos, excludedRepos, err := BuildPromotionParams(ctx)
	if err != nil {
		return nil, err
	}

	artifactProps, err := ParseArtifactProps(ctx)
	if err != nil {
		return nil, err
	}

	overwriteStrategy, err := ParseOverwriteStrategy(ctx)
	if err != nil {
		return nil, err
	}

	return &model.PromoteAppVersionRequest{
		Stage: stage,
		CommonPromoteAppVersion: model.CommonPromoteAppVersion{
			PromotionType:                promotionType,
			IncludedRepositoryKeys:       includedRepos,
			ExcludedRepositoryKeys:       excludedRepos,
			ArtifactAdditionalProperties: artifactProps,
			OverwriteStrategy:            overwriteStrategy,
		},
	}, nil
}

func GetPromoteAppVersionCommand(appContext app.Context) components.Command {
	cmd := &promoteAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.VersionPromote,
		Description: "Promote application version.",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vp"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The application key.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to promote.",
				Optional:    false,
			},
			{
				Name:        "target-stage",
				Description: "The target stage to which the application version should be promoted.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionPromote),
		Action: cmd.prepareAndRunCommand,
	}
}
