package cli

import (
	"github.com/jfrog/jfrog-cli-application/application/app"
	"github.com/jfrog/jfrog-cli-application/application/commands/system"
	"github.com/jfrog/jfrog-cli-application/application/commands/version"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

//const category = "Application Lifecycle"
//
//
//func GetJfrogApplicationCli() components.App {
//	appContext := app.NewAppContext()
//	appEntity := components.CreateEmbeddedApp(
//		category,
//		nil,
//		components.Namespace{
//			Name:        "app",
//			Description: "Tools for Application Lifecycle management",
//			Category:    category,
//			Commands: []components.Command{
//				system.GetPingCommand(appContext),
//				version.GetCreateAppVersionCommand(appContext),
//			},
//		},
//	)
//	return appEntity
//}

func GetJfrogApplicationCli() components.App {
	appContext := app.NewAppContext()
	appEntity := components.CreateApp(
		"app",
		"1.0.5",
		"JFrog Application CLI",
		[]components.Command{
			system.GetPingCommand(appContext),
			version.GetCreateAppVersionCommand(appContext),
			version.GetPromoteAppVersionCommand(appContext),
		},
	)
	return appEntity
}
