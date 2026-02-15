//go:build e2e

package utils

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
)

const (
	testJfrogUrlEnvVar   = "JFROG_APPTRUST_CLI_TESTS_JFROG_URL"
	testJfrogTokenEnvVar = "JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN"
)

type TestPackageResources struct {
	PackageType    string
	PackageName    string
	PackageVersion string
	PackagePath    string
	RepoKey        string
	BuildName      string
	BuildNumber    string
}

var (
	serverDetails              *coreConfig.ServerDetails
	artifactoryServicesManager artifactory.ArtifactoryServicesManager

	AppTrustCli *coreTests.JfrogCli

	testProjectKey  string
	testPackageRes  *TestPackageResources
	testArtifactPath string
)

func LoadCredentials() string {
	platformUrlFlag := flag.String("jfrog.url", getTestPlatformUrlFromEnvVar(), "JFrog Platform URL")
	accessTokenFlag := flag.String("jfrog.adminToken", os.Getenv(testJfrogTokenEnvVar), "JFrog Platform admin token")
	platformUrl := clientUtils.AddTrailingSlashIfNeeded(*platformUrlFlag)

	serverDetails = &coreConfig.ServerDetails{
		Url:            platformUrl,
		ArtifactoryUrl: platformUrl + "artifactory/",
		LifecycleUrl:   platformUrl + "lifecycle/",
		AccessToken:    *accessTokenFlag,
	}
	return fmt.Sprintf("--url=%s --access-token=%s", *platformUrlFlag, *accessTokenFlag)
}

func getTestPlatformUrlFromEnvVar() string {
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

func GetTestPackage(t *testing.T) *TestPackageResources {
	if testPackageRes == nil {
		buildName := GenerateUniqueKey("apptrust-cli-tests-build")
		buildNumber := "1"
		repoKey := createNpmRepo(t)
		sha256 := uploadPackageToArtifactory(t, repoKey, buildName, buildNumber)
		publishBuild(t, buildName, buildNumber, sha256)
	}
	return testPackageRes
}

func GetTestArtifact(t *testing.T) string {
	if testArtifactPath == "" {
		repoKey := createGenericRepo(t)
		testArtifactPath = uploadSimpleFileToArtifactory(t, repoKey, "test-artifact.txt")
	}
	return testArtifactPath
}

func GenerateUniqueKey(prefix string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	return fmt.Sprintf("%s-%s", prefix, timestamp)
}
