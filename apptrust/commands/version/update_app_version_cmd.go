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
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type updateAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.UpdateAppVersionRequest
}

func (uv *updateAppVersionCommand) Run() error {
	log.Info("Updating application version:", uv.applicationKey, "version:", uv.version)

	ctx, err := service.NewContext(*uv.serverDetails)
	if err != nil {
		log.Error("Failed to create service context:", err)
		return err
	}

	err = uv.versionService.UpdateAppVersion(ctx, uv.applicationKey, uv.version, uv.requestPayload)
	if err != nil {
		log.Error("Failed to update application version:", err)
		return err
	}

	log.Info("Successfully updated application version:", uv.applicationKey, "version:", uv.version)
	return nil
}

func (uv *updateAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return uv.serverDetails, nil
}

func (uv *updateAppVersionCommand) CommandName() string {
	return commands.VersionUpdate
}

func (uv *updateAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	if err := uv.parseFlagsAndSetFields(ctx); err != nil {
		return err
	}

	var err error
	uv.requestPayload, err = uv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}

	return commonCLiCommands.Exec(uv)
}

// parseFlagsAndSetFields parses CLI flags and sets struct fields accordingly.
func (uv *updateAppVersionCommand) parseFlagsAndSetFields(ctx *components.Context) error {
	uv.applicationKey = ctx.Arguments[0]
	uv.version = ctx.Arguments[1]

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	uv.serverDetails = serverDetails
	return nil
}

func (uv *updateAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.UpdateAppVersionRequest, error) {
	request := &model.UpdateAppVersionRequest{}

	if ctx.IsFlagSet(commands.TagFlag) {
		request.Tag = ctx.GetStringFlagValue(commands.TagFlag)
	}

	// Handle properties - use spec format: key=value1[,value2,...]
	if ctx.IsFlagSet(commands.PropertiesFlag) {
		properties, err := utils.ParseListPropertiesFlag(ctx.GetStringFlagValue(commands.PropertiesFlag))
		if err != nil {
			return nil, err
		}
		request.Properties = properties
	}

	// Handle delete properties
	if ctx.IsFlagSet(commands.DeletePropertiesFlag) {
		deleteProps := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.DeletePropertiesFlag))
		request.DeleteProperties = deleteProps
	}

	return request, nil
}

func GetUpdateAppVersionCommand(appContext app.Context) components.Command {
	cmd := &updateAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.VersionUpdate,
		Description: "Updates the user-defined annotations (tag and custom key-value properties) for a specified application version.",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vu"},
		Arguments: []components.Argument{
			{
				Name:        "app-key",
				Description: "The application key of the application for which the version is being updated.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version number (in SemVer format) for the application version to update.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionUpdate),
		Action: cmd.prepareAndRunCommand,
	}
}
