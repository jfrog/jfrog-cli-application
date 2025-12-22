//go:build itest

package test

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-application/cli"
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
)

const (
	TestJfrogUrlEnvVar   = "JFROG_APPTRUST_CLI_TESTS_JFROG_URL"
	TestJfrogTokenEnvVar = "JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN"
)

var (
	credentials string
	PlatformCli *coreTests.JfrogCli
)

func init() {
	loadCredentials()
	PlatformCli = coreTests.NewJfrogCli(plugins.RunCliWithPlugin(cli.GetJfrogCliApptrustApp()), "jf at", credentials)
}

func loadCredentials() {
	platformUrl := flag.String("jfrog.url", getTestUrlDefaultValue(), "JFrog Platform URL")
	adminAccessToken := flag.String("jfrog.adminToken", os.Getenv(TestJfrogTokenEnvVar), "JFrog Platform admin token")
	credentials = fmt.Sprintf("--url=%s --access-token=%s", *platformUrl, *adminAccessToken)
}

func getTestUrlDefaultValue() string {
	if os.Getenv(TestJfrogUrlEnvVar) != "" {
		return os.Getenv(TestJfrogUrlEnvVar)
	}
	return "http://localhost:8082/"
}

// GenerateUniqueAppKey generates a unique app key for testing
func GenerateUniqueAppKey(t *testing.T, prefix string) string {
	// Use timestamp and PID for uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d-%d", prefix, timestamp, os.Getpid())
}
