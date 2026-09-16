package cli

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/application"
	packagecmds "github.com/jfrog/jfrog-cli-application/apptrust/commands/package"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/system"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/version"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

func GetJfrogCliApptrustApp() components.App {
	appContext := app.NewAppContext()
	appEntity := components.CreateEmbeddedApp(
		"apptrust",
		nil,
		components.Namespace{
			Name:        "apptrust",
			Aliases:     []string{"at"},
			Description: "AppTrust commands.",
			AIDescription: `JFrog AppTrust commands (alias: at) for managing applications, application versions, and package bindings.

Use this namespace to:
- Create, update, and delete applications (logical groupings of releasable units identified by an application key).
- Create application versions from sources (builds, release bundles, application versions, packages, artifacts), promote them through stages, release them, roll back, and delete them.
- Bind or unbind packages to/from an application.
- Ping the AppTrust service to verify connectivity.

Prerequisites:
- A configured JFrog Platform server (jf c add or jf login) with AppTrust enabled, or per-command flags --url, --user, --access-token, --server-id.
- Required permissions on the target project/application for write operations.

Common patterns:
  $ jf apptrust ping
  $ jf apptrust app-create my-app --project=default
  $ jf apptrust version-create my-app 1.0.0 --source-type-builds="name=my-build, id=1"
  $ jf apptrust version-promote my-app 1.0.0 PROD
  $ jf apptrust version-distribute my-app 1.0.0

Related: jf rt, jf release-bundle commands.`,
			Category: "Command Namespaces",
			Commands: []components.Command{
				system.GetPingCommand(appContext),
				version.GetCreateAppVersionCommand(appContext),
				version.GetPromoteAppVersionCommand(appContext),
				version.GetRollbackAppVersionCommand(appContext),
				version.GetReleaseAppVersionCommand(appContext),
				version.GetDeleteAppVersionCommand(appContext),
				version.GetDistributeAppVersionCommand(appContext),
				version.GetRemoteDeleteAppVersionCommand(appContext),
				version.GetUpdateAppVersionCommand(appContext),
				version.GetUpdateAppVersionSourcesCommand(appContext),
				version.GetExportAppVersionCommand(appContext),
				version.GetImportAppVersionCommand(appContext),
				packagecmds.GetBindPackageCommand(appContext),
				packagecmds.GetUnbindPackageCommand(appContext),
				application.GetCreateAppCommand(appContext),
				application.GetUpdateAppCommand(appContext),
				application.GetDeleteAppCommand(appContext),
				application.GetExportAppCommand(appContext),
				application.GetImportAppCommand(appContext),
			},
		},
	)
	return appEntity
}
