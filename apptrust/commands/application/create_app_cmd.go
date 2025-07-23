package application

import (
	"encoding/json"

	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/applications"
)

type createAppCommand struct {
	serverDetails      *coreConfig.ServerDetails
	applicationService applications.ApplicationService
	requestBody        *model.AppDescriptor
}

func (cac *createAppCommand) Run() error {
	ctx, err := service.NewContext(*cac.serverDetails)
	if err != nil {
		return err
	}

	return cac.applicationService.CreateApplication(ctx, cac.requestBody)
}

func (cac *createAppCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return cac.serverDetails, nil
}

func (cac *createAppCommand) CommandName() string {
	return commands.AppCreate
}

func (cac *createAppCommand) buildRequestPayload(ctx *components.Context) (*model.AppDescriptor, error) {
	applicationKey := ctx.Arguments[0]

	var appDescriptor *model.AppDescriptor
	var err error

	if ctx.IsFlagSet(commands.SpecFlag) {
		appDescriptor, err = cac.loadFromSpec(ctx)
	} else {
		appDescriptor, err = cac.buildFromFlags(ctx)
	}

	if err != nil {
		return nil, err
	}

	appDescriptor.ApplicationKey = applicationKey
	if appDescriptor.ApplicationName == "" {
		appDescriptor.ApplicationName = applicationKey
	}

	return appDescriptor, nil
}

func (cac *createAppCommand) buildFromFlags(ctx *components.Context) (*model.AppDescriptor, error) {
	applicationName := ctx.GetStringFlagValue(commands.ApplicationNameFlag)

	project := ctx.GetStringFlagValue(commands.ProjectFlag)
	if project == "" {
		return nil, errorutils.CheckErrorf("--%s is mandatory", commands.ProjectFlag)
	}

	businessCriticalityStr := ctx.GetStringFlagValue(commands.BusinessCriticalityFlag)
	businessCriticality, err := utils.ValidateEnumFlag(
		commands.BusinessCriticalityFlag,
		businessCriticalityStr,
		model.BusinessCriticalityUnspecified,
		model.BusinessCriticalityValues)
	if err != nil {
		return nil, err
	}

	maturityLevelStr := ctx.GetStringFlagValue(commands.MaturityLevelFlag)
	maturityLevel, err := utils.ValidateEnumFlag(
		commands.MaturityLevelFlag,
		maturityLevelStr,
		model.MaturityLevelUnspecified,
		model.MaturityLevelValues)
	if err != nil {
		return nil, err
	}

	description := ctx.GetStringFlagValue(commands.DescriptionFlag)
	userOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.UserOwnersFlag))
	groupOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.GroupOwnersFlag))
	labelsMap, err := utils.ParseMapFlag(ctx.GetStringFlagValue(commands.LabelsFlag))
	if err != nil {
		return nil, err
	}

	return &model.AppDescriptor{
		ApplicationName:     applicationName,
		Description:         description,
		ProjectKey:          project,
		MaturityLevel:       maturityLevel,
		BusinessCriticality: businessCriticality,
		Labels:              labelsMap,
		UserOwners:          userOwners,
		GroupOwners:         groupOwners,
	}, nil
}

func (cac *createAppCommand) loadFromSpec(ctx *components.Context) (*model.AppDescriptor, error) {
	specFilePath := ctx.GetStringFlagValue(commands.SpecFlag)
	spec := new(model.AppDescriptor)
	specVars := coreutils.SpecVarsStringToMap(ctx.GetStringFlagValue(commands.SpecVarsFlag))
	content, err := fileutils.ReadFile(specFilePath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	if len(specVars) > 0 {
		content = coreutils.ReplaceVars(content, specVars)
	}

	err = json.Unmarshal(content, spec)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	if spec.ProjectKey == "" {
		return nil, errorutils.CheckErrorf("project_key is mandatory in spec file")
	}

	return spec, nil
}

func (cac *createAppCommand) prepareAndRunCommand(ctx *components.Context) error {
	if err := validateCreateAppContext(ctx); err != nil {
		return err
	}

	var err error
	cac.requestBody, err = cac.buildRequestPayload(ctx)
	if err != nil {
		return err
	}

	cac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(cac)
}

func validateCreateAppContext(ctx *components.Context) error {
	if err := validateNoSpecAndFlagsTogether(ctx); err != nil {
		return err
	}
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}
	return nil
}

func validateNoSpecAndFlagsTogether(ctx *components.Context) error {
	if ctx.IsFlagSet(commands.SpecFlag) {
		otherAppFlags := []string{
			commands.ApplicationNameFlag,
			commands.ProjectFlag,
			commands.DescriptionFlag,
			commands.BusinessCriticalityFlag,
			commands.MaturityLevelFlag,
			commands.LabelsFlag,
			commands.UserOwnersFlag,
			commands.GroupOwnersFlag,
		}
		for _, flag := range otherAppFlags {
			if ctx.IsFlagSet(flag) {
				return errorutils.CheckErrorf("the flag --%s is not allowed when --spec is provided.", flag)
			}
		}
	}
	return nil
}

func GetCreateAppCommand(appContext app.Context) components.Command {
	cmd := &createAppCommand{
		applicationService: appContext.GetApplicationService(),
	}
	return components.Command{
		Name:        commands.AppCreate,
		Description: "Create a new application.",
		Category:    common.CategoryApplication,
		Aliases:     []string{"ac"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to create.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.AppCreate),
		Action: cmd.prepareAndRunCommand,
	}
}
