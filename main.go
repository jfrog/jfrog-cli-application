package main

import (
	"github.com/jfrog/jfrog-cli-application/cli"
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
)

func main() {
	plugins.PluginMain(cli.GetJfrogCliApptrustApp())
}
