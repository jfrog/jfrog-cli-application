//go:build e2e

package e2e

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

const (
	TestJfrogUrlEnvVar   = "JFROG_APPTRUST_CLI_TESTS_JFROG_URL"
	TestJfrogTokenEnvVar = "JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN"
)

var (
	platformUrl      *string
	adminAccessToken *string

	credentials string
	AppTrustCli *coreTests.JfrogCli

	testProjectKey string
)

func loadCredentials() {
	platformUrl = flag.String("jfrog.url", getTestUrlDefaultValue(), "JFrog Platform URL")
	adminAccessToken = flag.String("jfrog.adminToken", os.Getenv(TestJfrogTokenEnvVar), "JFrog Platform admin token")
	credentials = fmt.Sprintf("--url=%s --access-token=%s", *platformUrl, *adminAccessToken)
}

func getTestUrlDefaultValue() string {
	if os.Getenv(TestJfrogUrlEnvVar) != "" {
		return os.Getenv(TestJfrogUrlEnvVar)
	}
	return "http://localhost:8082/"
}

func GetTestProjectKey(t *testing.T) string {
	if testProjectKey == "" {
		createTestProject(t)
	}
	return testProjectKey
}

func createTestProject(t *testing.T) {
	serverDetails := &config.ServerDetails{
		Url:         *platformUrl,
		AccessToken: *adminAccessToken,
	}
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	projectKey := GenerateUniqueKey("apptrust-cli-tests")
	projectParams := accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: projectKey,
			ProjectKey:  projectKey,
		},
	}
	err = accessManager.CreateProject(projectParams)
	assert.NoError(t, err)
	testProjectKey = projectKey
}

func deleteTestProject() {
	if testProjectKey == "" {
		return
	}
	serverDetails := &config.ServerDetails{
		Url:         *platformUrl,
		AccessToken: *adminAccessToken,
	}
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	if err != nil {
		log.Error("Failed to create Access service manager", err)
	}
	err = accessManager.DeleteProject(testProjectKey)
	if err != nil {
		log.Error("Failed to delete project", err)
	}
}

func DeleteApplication(t *testing.T, appKey string) {
	err := AppTrustCli.Exec("ad", appKey)
	assert.NoError(t, err)
}

func GenerateUniqueKey(prefix string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	return fmt.Sprintf("%s-%s", prefix, timestamp)
}
