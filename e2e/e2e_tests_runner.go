//go:build e2e

package e2e

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestJfrogUrlEnvVar   = "JFROG_APPTRUST_CLI_TESTS_JFROG_URL"
	TestJfrogTokenEnvVar = "JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN"
)

var (
	serverDetails              *coreConfig.ServerDetails
	artifactoryServicesManager artifactory.ArtifactoryServicesManager

	credentials string
	AppTrustCli *coreTests.JfrogCli

	testProjectKey  string
	testRepoKey     string
	testPackagePath string
)

func loadCredentials() {
	platformUrlFlag := flag.String("jfrog.url", getTestUrlDefaultValue(), "JFrog Platform URL")
	accessTokenFlag := flag.String("jfrog.adminToken", os.Getenv(TestJfrogTokenEnvVar), "JFrog Platform admin token")
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
	deleteNpmRepo()
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	if err != nil {
		log.Error("Failed to create Access service manager", err)
	}
	err = accessManager.DeleteProject(testProjectKey)
	if err != nil {
		log.Error("Failed to delete project", err)
	}
}

func CreateBasicApplication(t *testing.T, appKey string) {
	projectKey := GetTestProjectKey(t)
	err := AppTrustCli.Exec("ac", appKey, "--project="+projectKey, "--application-name="+appKey)
	assert.NoError(t, err)
}

func DeleteApplication(t *testing.T, appKey string) {
	err := AppTrustCli.Exec("ad", appKey)
	assert.NoError(t, err)
}

func GenerateUniqueKey(prefix string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	return fmt.Sprintf("%s-%s", prefix, timestamp)
}

func GetApplication(appKey string) (*model.AppDescriptor, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s", appKey)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint)
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var appDescriptor model.AppDescriptor
	err = json.Unmarshal(responseBody, &appDescriptor)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return &appDescriptor, statusCode, nil
}

func GetTestPackage(t *testing.T) string {
	if testPackagePath == "" {
		uploadPackageToArtifactory(t)
	}
	return testPackagePath
}

func uploadPackageToArtifactory(t *testing.T) {
	createNpmRepo(t)

	// Get the absolute path to the testdata file
	_, testFilePath, _, _ := runtime.Caller(0)
	npmPackageFilePath := filepath.Join(filepath.Dir(testFilePath), "testdata", "pizza-frontend.tgz")

	targetPath := testRepoKey + "/pizza-frontend.tgz"
	servicesManager := getArtifactoryServicesManager(t)
	uploadParams := services.NewUploadParams()
	uploadParams.Pattern = npmPackageFilePath
	uploadParams.Target = targetPath
	uploadParams.Flat = true
	uploaded, failed, err := servicesManager.UploadFiles(artifactory.UploadServiceOptions{FailFast: false}, uploadParams)
	require.NoError(t, err)
	require.Equal(t, 1, uploaded, "Expected exactly one uploaded file")
	require.Equal(t, 0, failed, "Expected zero failed uploads")
	testPackagePath = targetPath
}

func createNpmRepo(t *testing.T) {
	servicesManager := getArtifactoryServicesManager(t)
	repoKey := GetTestProjectKey(t) + "-npm-local"
	localRepoConfig := services.NewNpmLocalRepositoryParams()
	localRepoConfig.ProjectKey = GetTestProjectKey(t)
	localRepoConfig.Key = repoKey
	err := servicesManager.CreateLocalRepository().Npm(localRepoConfig)
	require.NoError(t, err)
	testRepoKey = repoKey
}

func deleteNpmRepo() {
	if testRepoKey == "" || artifactoryServicesManager == nil {
		return
	}

	err := artifactoryServicesManager.DeleteRepository(testRepoKey)
	if err != nil {
		log.Error("Failed to delete npm repo", err)
	}
}

func getArtifactoryServicesManager(t *testing.T) artifactory.ArtifactoryServicesManager {
	if artifactoryServicesManager == nil {
		var err error
		artifactoryServicesManager, err = utils.CreateServiceManager(serverDetails, -1, 0, false)
		require.NoError(t, err)
	}

	return artifactoryServicesManager
}
