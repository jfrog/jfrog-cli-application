//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/jfrog/jfrog-cli-application/cli"
	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
)

func TestMain(m *testing.M) {
	credentials := utils.LoadCredentials()
	utils.AppTrustCli = coreTests.NewJfrogCli(plugins.RunCliWithPlugin(cli.GetJfrogCliApptrustApp()), "jf at", credentials)
	code := m.Run()
	utils.DeleteTestProject()
	os.Exit(code)
}
