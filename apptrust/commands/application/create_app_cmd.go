package application

import (
	"encoding/json"
	"os"

	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
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
	responseBody       []byte
}

func (cac *createAppCommand) Run() error {
	ctx, err := service.NewContext(*cac.serverDetails)
	if err != nil {
		return err
	}

	cac.responseBody, err = cac.applicationService.CreateApplication(ctx, cac.requestBody)
	return err
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
	project := ctx.GetStringFlagValue(commands.ProjectFlag)
	if project == "" {
		return nil, errorutils.CheckErrorf("--%s is mandatory", commands.ProjectFlag)
	}

	descriptor := &model.AppDescriptor{
		ProjectKey: project,
	}

	err := populateApplicationFromFlags(ctx, descriptor)
	if err != nil {
		return nil, err
	}

	return descriptor, nil
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

	outputFormat, err := ctx.GetOutputFormat()
	if err != nil {
		return err
	}

	if err = commonCLiCommands.Exec(cac); err != nil {
		return err
	}

	return common.PrintResponse(cac.responseBody, outputFormat, os.Stdout, common.OrderedAppKeys)
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
			commands.MonitorPolicyFlag,
			commands.AutoPromoteStagesFlag,
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
		AIDescription: `Create a new application in AppTrust, identified by an application key and belonging to a project.

When to use:
- Register a new logical application that will own future versions and package bindings.
- Bootstrap AppTrust governance for a new service or product.

Prerequisites:
- A configured server with AppTrust enabled.
- Create permission on the target project.
- Either --project (mandatory when no --spec is used) or a --spec file with project_key set.

Common patterns:
  $ jf apptrust app-create my-app --project=default
  $ jf apptrust app-create my-app --project=default --application-name="My App" --desc="Service X"
  $ jf apptrust app-create my-app --project=default --business-criticality=high --maturity-level=production
  $ jf apptrust app-create my-app --project=default --labels="team=core;area=platform" --user-owners="alice;bob"
  $ jf apptrust app-create my-app --project=default --monitor-policy="type=version_count, value=5"
  $ jf apptrust app-create my-app --project=default --auto-promote-stages="DEV;PROD"
  $ jf apptrust app-create my-app --spec=app-spec.json --spec-vars="ENV=prod"

Gotchas:
- --spec is mutually exclusive with --application-name, --project, --desc, --business-criticality, --maturity-level, --labels, --user-owners, --group-owners, --monitor-policy, --auto-promote-stages.
- If --application-name is omitted, the application-key is used as the display name.
- --labels uses semicolon separators (not commas) and key=value pairs.
- --monitor-policy takes 'type=<type>[, value=<n>]'. 'value' is required (positive integer) when type is "time_frame_in_months" or "version_count", and must be omitted when type is "none".
- --auto-promote-stages uses semicolon separators and preserves the supplied stage order.

Related: jf apptrust app-update, jf apptrust app-delete, jf apptrust version-create`,
		Category:         common.CategoryApplication,
		Aliases:          []string{"ac"},
		SupportedFormats: []coreformat.OutputFormat{coreformat.Table, coreformat.Json},
		DefaultFormat:    coreformat.Json,
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
