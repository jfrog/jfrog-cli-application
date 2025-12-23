//go:build e2e

package e2e

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

const (
	TestJfrogUrlEnvVar   = "JFROG_APPTRUST_CLI_TESTS_JFROG_URL"
	TestJfrogTokenEnvVar = "JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN"
)

var (
	serverDetails *config.ServerDetails

	credentials string
	AppTrustCli *coreTests.JfrogCli

	testProjectKey string
)

func loadCredentials() {
	platformUrl := flag.String("jfrog.url", getTestUrlDefaultValue(), "JFrog Platform URL")
	adminAccessToken := flag.String("jfrog.adminToken", os.Getenv(TestJfrogTokenEnvVar), "JFrog Platform admin token")

	serverDetails = &config.ServerDetails{
		Url:         *platformUrl,
		AccessToken: *adminAccessToken,
	}
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

func GetApplication(t *testing.T, appKey string) *model.AppDescriptor {
	ctx, err := service.NewContext(*serverDetails)
	assert.NoError(t, err)

	endpoint := fmt.Sprintf("/v1/applications/%s", appKey)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Expected status 200, got %d. Response: %s", response.StatusCode, string(responseBody))

	var appDescriptor model.AppDescriptor
	err = json.Unmarshal(responseBody, &appDescriptor)
	assert.NoError(t, errorutils.CheckError(err))

	return &appDescriptor
}
