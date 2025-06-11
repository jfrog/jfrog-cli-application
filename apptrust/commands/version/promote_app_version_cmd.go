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
	return commands.PromoteAppVersion
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

	var includedRepos []string
	var excludedRepos []string

	if includeReposStr := ctx.GetStringFlagValue(commands.IncludeReposFlag); includeReposStr != "" {
		includedRepos = utils.ParseSliceFlag(includeReposStr)
	}

	if excludeReposStr := ctx.GetStringFlagValue(commands.ExcludeReposFlag); excludeReposStr != "" {
		excludedRepos = utils.ParseSliceFlag(excludeReposStr)
	}

	// Validate promotion type flag
	promotionType := ctx.GetStringFlagValue(commands.PromotionTypeFlag)
	validatedPromotionType, err := utils.ValidateEnumFlag(commands.PromotionTypeFlag, promotionType, model.PromotionTypeCopy, model.PromotionTypeValues)
	if err != nil {
		return nil, err
	}

	// If dry-run is true, override with dry_run
	dryRun := ctx.GetBoolFlagValue(commands.DryRunFlag)
	if dryRun {
		validatedPromotionType = model.PromotionTypeDryRun
	}

	return &model.PromoteAppVersionRequest{
		Stage:                  stage,
		PromotionType:          validatedPromotionType,
		IncludedRepositoryKeys: includedRepos,
		ExcludedRepositoryKeys: excludedRepos,
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
				Name:        "application-key",
				Description: "The application key",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to promote",
				Optional:    false,
			},
			{
				Name:        "target-stage",
				Description: "The target stage to which the application version should be promoted",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.PromoteAppVersion),
		Action: cmd.prepareAndRunCommand,
	}
}
