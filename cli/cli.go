package cli

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/application"
	packagecmds "github.com/jfrog/jfrog-cli-application/apptrust/commands/package"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/system"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/version"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

//func GetJfrogCliApptrustApp() components.App {
//	appContext := app.NewAppContext()
//	appEntity := components.CreateEmbeddedApp(
//		"apptrust",
//		nil,
//		components.Namespace{
//			Name:        "apptrust",
//			Aliases:     []string{"at"},
//			Description: "AppTrust commands.",
//			Category:    "Command Namespaces",
//			Commands: []components.Command{
//				system.GetPingCommand(appContext),
//				version.GetCreateAppVersionCommand(appContext),
//				version.GetPromoteAppVersionCommand(appContext),
//				version.GetDeleteAppVersionCommand(appContext),
//				packagecmds.GetBindPackageCommand(appContext),
//				packagecmds.GetUnbindPackageCommand(appContext),
//				application.GetCreateAppCommand(appContext),
//				application.GetUpdateAppCommand(appContext),
//				application.GetDeleteAppCommand(appContext),
//			},
//		},
//	)
//	return appEntity
//}

func GetJfrogCliApptrustApp() components.App {
	appContext := app.NewAppContext()
	appEntity := components.CreateApp(
		"apptrust",
		"1.0.5",
		"JFrog AppTrust CLI",
		[]components.Command{
			system.GetPingCommand(appContext),
			version.GetCreateAppVersionCommand(appContext),
			version.GetPromoteAppVersionCommand(appContext),
			version.GetDeleteAppVersionCommand(appContext),
			packagecmds.GetBindPackageCommand(appContext),
			packagecmds.GetUnbindPackageCommand(appContext),
			application.GetCreateAppCommand(appContext),
			application.GetUpdateAppCommand(appContext),
			application.GetDeleteAppCommand(appContext),
		},
	)
	return appEntity
}
