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
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type remoteDeleteAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	requestPayload *model.RemoteDeleteAppVersionRequest
	quiet          bool
}

func (rd *remoteDeleteAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*rd.serverDetails)
	if err != nil {
		return err
	}

	return rd.versionService.RemoteDeleteAppVersion(ctx, rd.applicationKey, rd.version, rd.requestPayload)
}

func (rd *remoteDeleteAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return rd.serverDetails, nil
}

func (rd *remoteDeleteAppVersionCommand) CommandName() string {
	return commands.VersionRemoteDelete
}

func (rd *remoteDeleteAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) != 2 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	rd.applicationKey = ctx.Arguments[0]
	rd.version = ctx.Arguments[1]
	rd.quiet = ctx.GetBoolFlagValue(commands.QuietFlag)

	if err := ValidateDistributionFlags(ctx); err != nil {
		return err
	}

	serverDetails, err := utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}
	rd.serverDetails = serverDetails

	rd.requestPayload, err = rd.buildRequestPayload(ctx)
	if errorutils.CheckError(err) != nil {
		return err
	}

	confirmed, err := rd.confirmRemoteDelete()
	if err != nil || !confirmed {
		return err
	}

	return commonCLiCommands.Exec(rd)
}

func (rd *remoteDeleteAppVersionCommand) buildRequestPayload(ctx *components.Context) (*model.RemoteDeleteAppVersionRequest, error) {
	distributionRules, err := ParseDistributionRules(ctx)
	if err != nil {
		return nil, err
	}

	return &model.RemoteDeleteAppVersionRequest{
		DryRun:            ctx.GetBoolFlagValue(commands.DryRunFlag),
		DistributionRules: distributionRules,
	}, nil
}

func (rd *remoteDeleteAppVersionCommand) distributionRulesEmpty() bool {
	rules := rd.requestPayload.DistributionRules
	if len(rules) == 0 {
		return true
	}
	if len(rules) == 1 {
		rule := rules[0]
		return rule.SiteName == "" && rule.CityName == "" && len(rule.CountryCodes) == 0
	}
	return false
}

func (rd *remoteDeleteAppVersionCommand) confirmRemoteDelete() (bool, error) {
	if rd.quiet {
		return true, nil
	}

	message := fmt.Sprintf("Are you sure you want to delete the application version '%s/%s' remotely ", rd.applicationKey, rd.version)
	if rd.distributionRulesEmpty() {
		message += "from all distribution targets?"
	} else {
		bytes, err := json.Marshal(rd.requestPayload.DistributionRules)
		if err != nil {
			return false, errorutils.CheckError(err)
		}

		log.Output(clientutils.IndentJson(bytes))
		message += "from all targets with the above distribution rules?"
	}

	return coreutils.AskYesNo(message, false), nil
}

func GetRemoteDeleteAppVersionCommand(appContext app.Context) components.Command {
	cmd := &remoteDeleteAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionRemoteDelete,
		Description: "Delete an application version from distribution targets.",
		AIDescription: `Delete a previously distributed application version's artifacts from one or more distribution targets according to distribution rules.

When to use:
- Remove a version that was distributed to distribution targets and should no longer be available there.
- Roll back a distribution without deleting the application version itself.

Prerequisites:
- The application version must have been distributed to the targeted distribution target(s).
- Configured server and distribution permission on the application's project.

Common patterns:
  $ jf apptrust version-delete-remote my-app 1.0.0
  $ jf apptrust version-delete-remote my-app 1.0.0 --quiet
  $ jf apptrust version-delete-remote my-app 1.0.0 --site="us-*" --country-codes="US;CA"
  $ jf apptrust version-delete-remote my-app 1.0.0 --dist-rules=/path/to/dist-rules.json
  $ jf apptrust version-delete-remote my-app 1.0.0 --dry-run

Gotchas:
- This only removes the version from distribution targets; the application version itself remains. To delete the version record, use version-delete.
- A confirmation prompt is shown before deletion; pass --quiet to skip it (e.g. in scripts).
- --dist-rules can't be combined with --site, --city or --country-codes.
- --dry-run simulates the operation without deleting anything.

Related: jf apptrust version-distribute, jf apptrust version-delete`,
		Category: common.CategoryVersion,
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The application key.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to delete from the distribution targets.",
				Optional:    false,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionRemoteDelete),
		Action: cmd.prepareAndRunCommand,
	}
}
