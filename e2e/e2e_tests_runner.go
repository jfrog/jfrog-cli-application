//go:build e2e

package e2e

import (
	"flag"
	"fmt"
	"os"
	"testing"

	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
)

const (
	testJfrogUrlEnvVar   = "JFROG_APPTRUST_CLI_TESTS_JFROG_URL"
	testJfrogTokenEnvVar = "JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN"
)

var (
	serverDetails              *coreConfig.ServerDetails
	artifactoryServicesManager artifactory.ArtifactoryServicesManager

	credentials string
	AppTrustCli *coreTests.JfrogCli

	testProjectKey string
	testRepoKey    string

	testPackageType    string
	testPackageName    string
	testPackageVersion string
	testPackagePath    string
)

func loadCredentials() {
	platformUrlFlag := flag.String("jfrog.url", getTestUrlDefaultValue(), "JFrog Platform URL")
	accessTokenFlag := flag.String("jfrog.adminToken", os.Getenv(testJfrogTokenEnvVar), "JFrog Platform admin token")
	platformUrl := clientUtils.AddTrailingSlashIfNeeded(*platformUrlFlag)
	artifactoryUrl := platformUrl + "artifactory/"

	serverDetails = &coreConfig.ServerDetails{
		Url:            clientUtils.AddTrailingSlashIfNeeded(*platformUrlFlag),
		ArtifactoryUrl: artifactoryUrl,
		AccessToken:    *accessTokenFlag,
	}
	credentials = fmt.Sprintf("--url=%s --access-token=%s", *platformUrlFlag, *accessTokenFlag)
}

func getTestUrlDefaultValue() string {
	if os.Getenv(testJfrogUrlEnvVar) != "" {
		return os.Getenv(testJfrogUrlEnvVar)
	}
	return "http://localhost:8082/"
}

func GetTestProjectKey(t *testing.T) string {
	if testProjectKey == "" {
		createTestProject(t)
	}
	return testProjectKey
}

func getTestPackage(t *testing.T) (pkgType, pkgName, pkgVersion, pkgPath string) {
	// Upload the test package to Artifactory if not already done
	if testPackageName == "" {
		uploadPackageToArtifactory(t)
	}
	return testPackageType, testPackageName, testPackageVersion, testPackagePath
}
