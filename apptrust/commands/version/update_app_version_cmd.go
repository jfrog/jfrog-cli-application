package version

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"strings"

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

type updateAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.UpdateAppVersionRequest
}

func (uv *updateAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*uv.serverDetails)
	if err != nil {
		return err
	}
	return uv.versionService.UpdateAppVersion(ctx, uv.applicationKey, uv.version, uv.requestPayload)
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

	uv.applicationKey = ctx.Arguments[0]
	uv.version = ctx.Arguments[1]

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	uv.serverDetails = serverDetails

	uv.requestPayload, err = uv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}

	return commonCLiCommands.Exec(uv)
}

func (uv *updateAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.UpdateAppVersionRequest, error) {
	request := &model.UpdateAppVersionRequest{}

	// Handle tag - no validation, just pass through
	if ctx.IsFlagSet(commands.TagFlag) {
		request.Tag = ctx.GetStringFlagValue(commands.TagFlag)
	}

	// Handle properties - support multiple values per key
	if ctx.IsFlagSet(commands.PropertiesFlag) {
		properties, err := uv.parseProperties(ctx.GetStringFlagValue(commands.PropertiesFlag))
		if err != nil {
			return nil, err
		}
		request.Properties = properties
	}

	// Handle delete properties
	if ctx.IsFlagSet(commands.DeletePropertyFlag) {
		deleteProps := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.DeletePropertyFlag))
		request.DeleteProperties = deleteProps
	}

	return request, nil
}

func (uv *updateAppVersionCommand) parseProperties(propertiesStr string) (map[string][]string, error) {
	// Format: "key1=value1[,value2,...];key2=value3[,value4,...]"
	if propertiesStr == "" {
		return nil, nil
	}

	result := make(map[string][]string)
	pairs := strings.Split(propertiesStr, ";")

	for _, pair := range pairs {
		keyValue := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(keyValue) != 2 {
			return nil, errorutils.CheckErrorf("invalid property format: '%s' (expected key=value1[,value2,...])", pair)
		}

		key := strings.TrimSpace(keyValue[0])
		valuesStr := strings.TrimSpace(keyValue[1])

		if key == "" {
			return nil, errorutils.CheckErrorf("property key cannot be empty")
		}

		var values []string
		if valuesStr != "" {
			values = strings.Split(valuesStr, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
		}
		// Always set the key, even with empty values (to clear values)

		result[key] = values
	}

	return result, nil
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
