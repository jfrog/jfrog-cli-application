package version

import (
	"encoding/json"
	"fmt"

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
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type importAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	archivePath    string
	options        *model.ImportAppVersionOptions
}

func (iac *importAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*iac.serverDetails)
	if err != nil {
		return err
	}

	responseBody, err := iac.versionService.ImportAppVersion(ctx, iac.applicationKey, iac.archivePath, iac.options)
	if err != nil {
		return err
	}

	result, err := parseImportAppVersionResponse(responseBody)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Import of application version '%s/%s' was successful.", result.Name, result.Version))
	return nil
}

type importAppVersionResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func parseImportAppVersionResponse(responseBody []byte) (*importAppVersionResponse, error) {
	var result importAppVersionResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse import response: %w", err)
	}
	return &result, nil
}

func (iac *importAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return iac.serverDetails, nil
}

func (iac *importAppVersionCommand) CommandName() string {
	return commands.VersionImport
}

func (iac *importAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	iac.applicationKey = ctx.Arguments[0]
	iac.archivePath = ctx.Arguments[1]

	exists, err := fileutils.IsFileExists(iac.archivePath, false)
	if err != nil {
		return err
	}
	if !exists {
		return errorutils.CheckErrorf("file not found: %s", iac.archivePath)
	}

	iac.options, err = iac.buildRequestPayload(ctx)
	if err != nil {
		return err
	}

	iac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(iac)
}

func (iac *importAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.ImportAppVersionOptions, error) {
	unpromoted := ctx.GetBoolFlagValue(commands.UnpromotedFlag)
	mappingSet := ctx.IsFlagSet(commands.MappingPatternFlag) || ctx.IsFlagSet(commands.MappingTargetFlag)
	if unpromoted && mappingSet {
		return nil, errorutils.CheckErrorf("the --%s option can't be used with --%s or --%s",
			commands.UnpromotedFlag, commands.MappingPatternFlag, commands.MappingTargetFlag)
	}
	if unpromoted {
		return &model.ImportAppVersionOptions{Mode: model.ImportModeUnpromoted}, nil
	}

	modifications, err := ParseDistributionModifications(ctx)
	if err != nil {
		return nil, err
	}

	options := &model.ImportAppVersionOptions{Mode: model.ImportModePathMapping}
	if modifications != nil {
		options.PathMappings = modifications.PathMappings
	}
	return options, nil
}

func GetImportAppVersionCommand(appContext app.Context) components.Command {
	cmd := &importAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionImport,
		Description: "Import an application version archive into AppTrust.",
		AIDescription: `Import an application version ZIP archive into an existing application for air-gap transfer.

When to use:
- Restore or copy an application version produced by jf apptrust version-export.

Prerequisites:
- The target application must already exist.
- A ZIP archive produced by version-export.
- Configured server and permission to import the version.

Common patterns:
  $ jf apptrust version-import my-app ./my-app-1.0.0.zip
  $ jf apptrust version-import my-app ./archive.zip --mapping-pattern="repo/(*)" --mapping-target="target/{1}"
  $ jf apptrust version-import my-app ./archive.zip --unpromoted
  $ jf at vimp my-app ./archive.zip --server-id=my-server

Gotchas:
- By default the artifacts are copied into Artifactory repositories on import (optionally remapped with --mapping-pattern and --mapping-target). With --unpromoted they are not placed in any repository until the version is promoted to a stage. Neither mode promotes the version.
- --mapping-pattern and --mapping-target must be provided together and cannot be combined with --unpromoted.
- A successful result means the import was accepted (HTTP 202). Artifact copy or path mapping may still run asynchronously.

Related: jf apptrust version-export, jf apptrust app-import`,
		Category: common.CategoryVersion,
		Aliases:  []string{"vimp"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application to import the version into.",
				Optional:    false,
			},
			{
				Name:        "path to archive",
				Description: "Local path to the application version ZIP archive.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionImport),
		Action: cmd.prepareAndRunCommand,
	}
}
