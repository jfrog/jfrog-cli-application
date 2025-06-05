package version

import (
	"encoding/json"

	"github.com/jfrog/jfrog-cli-application/apptrust/service/versions"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type createAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	requestPayload *model.CreateAppVersionRequest
}

type createVersionSpec struct {
	Packages []model.CreateVersionPackage `json:"packages"`
}

func (cv *createAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*cv.serverDetails)
	if err != nil {
		return err
	}

	return cv.versionService.CreateAppVersion(ctx, cv.requestPayload)
}

func (cv *createAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return cv.serverDetails, nil
}

func (cv *createAppVersionCommand) CommandName() string {
	return commands.CreateAppVersion
}

func (cv *createAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if err := validateCreateAppVersionContext(ctx); err != nil {
		return err
	}
	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	cv.serverDetails = serverDetails
	cv.requestPayload, err = cv.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}
	return commonCLiCommands.Exec(cv)
}

func (cv *createAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.CreateAppVersionRequest, error) {
	var packages []model.CreateVersionPackage
	if ctx.IsFlagSet(commands.SpecFlag) {
		err := loadPackagesFromSpec(ctx, &packages)
		if errorutils.CheckError(err) != nil {
			return nil, err
		}
	} else {
		packages = append(packages, model.CreateVersionPackage{
			Type:       ctx.GetStringFlagValue(commands.PackageTypeFlag),
			Name:       ctx.GetStringFlagValue(commands.PackageNameFlag),
			Version:    ctx.GetStringFlagValue(commands.PackageVersionFlag),
			Repository: ctx.GetStringFlagValue(commands.PackageRepositoryFlag),
		})
	}

	return &model.CreateAppVersionRequest{
		ApplicationKey: ctx.GetStringFlagValue(commands.ApplicationKeyFlag),
		Version:        ctx.Arguments[0],
		Packages:       packages,
	}, nil
}

func loadPackagesFromSpec(ctx *components.Context, packages *[]model.CreateVersionPackage) error {
	specFilePath := ctx.GetStringFlagValue(commands.SpecFlag)
	spec := new(createVersionSpec)
	specVars := coreutils.SpecVarsStringToMap(ctx.GetStringFlagValue("spec-vars"))
	content, err := fileutils.ReadFile(specFilePath)
	if errorutils.CheckError(err) != nil {
		return err
	}

	if len(specVars) > 0 {
		content = coreutils.ReplaceVars(content, specVars)
	}

	err = json.Unmarshal(content, spec)
	if errorutils.CheckError(err) != nil {
		return err
	}

	// add spec packages to the packages list
	*packages = append(*packages, spec.Packages...)
	return nil
}

func validateCreateAppVersionContext(ctx *components.Context) error {
	if show, err := pluginsCommon.ShowCmdHelpIfNeeded(ctx, ctx.Arguments); show || err != nil {
		return err
	}
	if len(ctx.Arguments) != 1 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	// Use spec flag if provided, if not check for package flags
	err := utils.AssertValueProvided(ctx, commands.SpecFlag)
	if err != nil {
		err = utils.AssertValueProvided(ctx, commands.PackageNameFlag)
		if err != nil {
			return handleMissingPackageDetailsError()
		}
		err = utils.AssertValueProvided(ctx, commands.PackageVersionFlag)
		if err != nil {
			return handleMissingPackageDetailsError()
		}
		err = utils.AssertValueProvided(ctx, commands.PackageRepositoryFlag)
		if err != nil {
			return handleMissingPackageDetailsError()
		}
	}

	return nil
}

func GetCreateAppVersionCommand(appContext app.Context) components.Command {
	cmd := &createAppVersionCommand{versionService: appContext.GetVersionService()}
	return components.Command{
		Name:        commands.CreateAppVersion,
		Description: "Create application version",
		Category:    common.CategoryVersion,
		Aliases:     []string{"vc"},
		Arguments: []components.Argument{
			{
				Name:        "version-name",
				Description: "The name of the version",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.CreateAppVersion),
		Action: cmd.prepareAndRunCommand,
	}
}

func handleMissingPackageDetailsError() error {
	return errorutils.CheckErrorf("Missing packages information. Please provide the following flags --%s or the set of: --%s, --%s, --%s, --%s",
		commands.SpecFlag, commands.PackageTypeFlag, commands.PackageNameFlag, commands.PackageVersionFlag, commands.PackageRepositoryFlag)
}
