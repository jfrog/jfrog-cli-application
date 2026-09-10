package version

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/common"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/versions"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	commonCLiCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	artServices "github.com/jfrog/jfrog-client-go/artifactory/services"
	artServicesUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	exportStatusCompleted  = "COMPLETED"
	exportStatusInProgress = "IN_PROGRESS"
	exportStatusFailed     = "FAILED"

	exportPollInterval = 10 * time.Second
	exportPollTimeout  = 60 * time.Minute

	downloadMaxSplitCount = 15
)

type exportAppVersionCommand struct {
	versionService versions.VersionService
	serverDetails  *coreConfig.ServerDetails
	applicationKey string
	version        string
	targetPath     string
	minSplitSize   int64
	splitCount     int
}

func (eac *exportAppVersionCommand) Run() error {
	ctx, err := service.NewContext(*eac.serverDetails)
	if err != nil {
		return err
	}

	log.Info("Exporting application version archive...")
	if err = eac.versionService.TriggerExport(ctx, eac.applicationKey, eac.version); err != nil {
		return err
	}

	status, err := eac.waitForExport(ctx)
	if err != nil {
		return err
	}

	log.Debug("Downloading the exported application version archive...")
	downloadParams := artServices.DownloadParams{
		CommonParams: &artServicesUtils.CommonParams{
			Pattern: strings.TrimPrefix(status.RelativeDownloadURL, "/"),
			Target:  eac.targetPath,
		},
		MinSplitSize: eac.minSplitSize,
		SplitCount:   eac.splitCount,
	}
	downloaded, failed, err := downloadExportedArchive(eac.serverDetails, downloadParams)
	if err != nil {
		return err
	}
	if failed > 0 || downloaded < 1 {
		return errorutils.CheckErrorf("failed to download exported application version archive")
	}

	log.Info("Successfully downloaded application version archive")
	return nil
}

func (eac *exportAppVersionCommand) waitForExport(ctx service.Context) (*model.AppVersionExportStatus, error) {
	var status *model.AppVersionExportStatus
	pollingExecutor := &httputils.PollingExecutor{
		Timeout:         exportPollTimeout,
		PollingInterval: exportPollInterval,
		MsgPrefix:       fmt.Sprintf("Getting exported application version %s/%s status...", eac.applicationKey, eac.version),
		PollingAction: func() (shouldStop bool, responseBody []byte, err error) {
			status, err = eac.versionService.GetExportStatus(ctx, eac.applicationKey, eac.version)
			if err != nil {
				return true, nil, err
			}
			stop, err := shouldStopExportPolling(status)
			return stop, nil, err
		},
	}
	_, err := pollingExecutor.Execute()
	if err != nil {
		return nil, err
	}
	return status, nil
}

func shouldStopExportPolling(status *model.AppVersionExportStatus) (shouldStop bool, err error) {
	switch status.Status {
	case exportStatusInProgress:
		return false, nil
	case exportStatusFailed:
		return true, errorutils.CheckErrorf("application version export failed: %s", status.Message)
	case exportStatusCompleted:
		return true, nil
	default:
		return true, errorutils.CheckErrorf("received unexpected export status: '%s'", status.Status)
	}
}

func downloadExportedArchive(serverDetails *coreConfig.ServerDetails, params artServices.DownloadParams) (int, int, error) {
	serviceManager, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return 0, 0, err
	}
	return serviceManager.DownloadFiles(params)
}

func (eac *exportAppVersionCommand) ServerDetails() (*coreConfig.ServerDetails, error) {
	return eac.serverDetails, nil
}

func (eac *exportAppVersionCommand) CommandName() string {
	return commands.VersionExport
}

func (eac *exportAppVersionCommand) prepareAndRunCommand(ctx *components.Context) error {
	if len(ctx.Arguments) < 2 || len(ctx.Arguments) > 3 {
		return pluginsCommon.WrongNumberOfArgumentsHandler(ctx)
	}

	eac.applicationKey = ctx.Arguments[0]
	eac.version = ctx.Arguments[1]
	if len(ctx.Arguments) == 3 {
		eac.targetPath = ctx.Arguments[2]
	}

	var err error
	eac.minSplitSize, eac.splitCount, err = parseExportDownloadFlags(ctx)
	if err != nil {
		return err
	}

	eac.serverDetails, err = utils.ServerDetailsByFlags(ctx)
	if err != nil {
		return err
	}

	return commonCLiCommands.Exec(eac)
}

func parseExportDownloadFlags(ctx *components.Context) (minSplitSize int64, splitCount int, err error) {
	minSplitSize = commands.DefaultDownloadMinSplitKb
	if ctx.GetStringFlagValue(commands.MinSplitFlag) != "" {
		minSplitSize, err = strconv.ParseInt(ctx.GetStringFlagValue(commands.MinSplitFlag), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("the '--%s' option should have a numeric value", commands.MinSplitFlag)
		}
	}

	splitCount = commands.DefaultDownloadSplitCount
	if ctx.GetStringFlagValue(commands.SplitCountFlag) != "" {
		splitCount, err = strconv.Atoi(ctx.GetStringFlagValue(commands.SplitCountFlag))
		if err != nil {
			return 0, 0, fmt.Errorf("the '--%s' option should have a numeric value", commands.SplitCountFlag)
		}
		if splitCount > downloadMaxSplitCount {
			return 0, 0, fmt.Errorf("the '--%s' option value is limited to a maximum of %d", commands.SplitCountFlag, downloadMaxSplitCount)
		}
		if splitCount < 0 {
			return 0, 0, fmt.Errorf("the '--%s' option cannot have a negative value", commands.SplitCountFlag)
		}
	}
	return minSplitSize, splitCount, nil
}

func GetExportAppVersionCommand(appContext app.Context) components.Command {
	cmd := &exportAppVersionCommand{
		versionService: appContext.GetVersionService(),
	}
	return components.Command{
		Name:        commands.VersionExport,
		Description: "Triggers the export process and downloads the application version archive",
		AIDescription: `Export an application version as a ZIP archive for air-gap transfer.

When to use:
- Copy an application version between disconnected AppTrust instances.

Prerequisites:
- The application version must exist.
- Configured server and permission to export the version.
- Export is not available on Edge nodes.

Common patterns:
  $ jf apptrust version-export my-app 1.0.0
  $ jf apptrust version-export my-app 1.0.0 ./exports/
  $ jf apptrust version-export my-app 1.0.0 ./my-app-1.0.0.zip
  $ jf at vexp my-app 1.0.0 ./exports/ --server-id=my-server

Gotchas:
- The export is asynchronous: the CLI polls until the archive is ready, then downloads it from Artifactory.
- Trailing slash on target = directory (original archive name is written inside); no slash = rename to that file.
- The server-side export is not deleted after download.

Related: jf apptrust version-import, jf apptrust app-export`,
		Category: common.CategoryVersion,
		Aliases:  []string{"vexp"},
		Arguments: []components.Argument{
			{
				Name:        "application-key",
				Description: "The key of the application whose version to export.",
				Optional:    false,
			},
			{
				Name:        "version",
				Description: "The version to export.",
				Optional:    false,
			},
			{
				Name:        "target pattern",
				Description: "Local filesystem target path. If it ends with a slash, it is assumed to be a directory and the export archive is written into it. If there is no terminal slash, the target path is assumed to be a file to which the export file should be renamed.",
				Optional:    true,
			},
		},
		Flags:  commands.GetCommandFlags(commands.VersionExport),
		Action: cmd.prepareAndRunCommand,
	}
}
