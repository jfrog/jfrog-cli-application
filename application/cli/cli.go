package cli

import (
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

const category = "Application Lifecycle"

func GetJfrogApplicationCli() components.App {
	app := components.CreateEmbeddedApp(
		category,
		nil,
		components.Namespace{
			Name:        "app",
			Description: "Tools for Application Lifecycle management",
			Category:    category,
			Commands:    []components.Command{},
		},
	)
	return app
}
