//go:build itest

package test

import (
	"os"
	"testing"

	"github.com/jfrog/jfrog-cli-application/cli"
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
)

func TestMain(m *testing.M) {
	loadCredentials()
	AppTrustCli = coreTests.NewJfrogCli(plugins.RunCliWithPlugin(cli.GetJfrogCliApptrustApp()), "jf at", credentials)
	code := m.Run()
	deleteTestProject()
	os.Exit(code)
}
