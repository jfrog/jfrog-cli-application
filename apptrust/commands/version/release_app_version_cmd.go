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

type ReleaseAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.ReleaseAppVersionRequest
	sync           bool
}

func (rv *ReleaseAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*rv.serverDetails)
	if err != nil {
		return err
	}

	return rv.versionService.ReleaseAppVersion(ctx, rv.applicationKey, rv.version, rv.requestPayload, rv.sync)
}

func (rv *ReleaseAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return rv.serverDetails, nil
}

func (rv *ReleaseAppVersionCommand) CommandName() string {
	return commands.ReleaseAppVersion
}

func (rv *ReleaseAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	// Extract from arguments
	rv.applicationKey = ctx.Arguments[0]
	rv.version = ctx.Arguments[1]

	// Extract sync flag value
	rv.sync = ctx.GetBoolFlagValue(commands.SyncFlag)

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

func (rv *ReleaseAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.ReleaseAppVersionRequest, error) {
	var includedRepos []string
	var excludedRepos []string
	var artifactProps map[string]string

	if includeReposStr := ctx.GetStringFlagValue(commands.IncludeReposFlag); includeReposStr != "" {
		includedRepos = utils.ParseSliceFlag(includeReposStr)
	}

	if excludeReposStr := ctx.GetStringFlagValue(commands.ExcludeReposFlag); excludeReposStr != "" {
		excludedRepos = utils.ParseSliceFlag(excludeReposStr)
	}

	if propsStr := ctx.GetStringFlagValue(commands.PropsFlag); propsStr != "" {
		var err error
		artifactProps, err = utils.ParseMapFlag(propsStr)
		if err != nil {
			return nil, err
		}
	}

	// Validate promotion type flag
	promotionType := ctx.GetStringFlagValue(commands.PromotionTypeFlag)

	// For validation, we need to add the dry_run option
	allowedValues := append([]string{}, model.PromotionTypeValues...)
	allowedValues = append(allowedValues, model.PromotionTypeDryRun)

	validatedPromotionType, err := utils.ValidateEnumFlag(commands.PromotionTypeFlag, promotionType, model.PromotionTypeCopy, allowedValues)
	if err != nil {
		return nil, err
	}

	return &model.ReleaseAppVersionRequest{
		PromotionType:                validatedPromotionType,
		IncludedRepositoryKeys:       includedRepos,
		ExcludedRepositoryKeys:       excludedRepos,
		ArtifactAdditionalProperties: artifactProps,
	}, nil
}

func GetReleaseAppVersionCommand(appContext app.Context) components.Command {
	cmd := &ReleaseAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.ReleaseAppVersion,
		Description: "Release application version",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vr"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The application key",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to release",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.ReleaseAppVersion),
		Action: cmd.prepareAndRunCommand,
	}
}
