//go:build e2e

package e2e

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

func createTestProject(t *testing.T) {
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	projectKey := generateUniqueKey("apptrust-cli-tests")
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

func createBasicApplication(t *testing.T, appKey string) {
	projectKey := GetTestProjectKey(t)
	err := AppTrustCli.Exec("ac", appKey, "--project="+projectKey, "--application-name="+appKey)
	assert.NoError(t, err)
}

func deleteApplication(t *testing.T, appKey string) {
	err := AppTrustCli.Exec("ad", appKey)
	assert.NoError(t, err)
}

func generateUniqueKey(prefix string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	return fmt.Sprintf("%s-%s", prefix, timestamp)
}

func getApplication(appKey string) (*model.AppDescriptor, int, error) {
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

func getTestPackage(t *testing.T) (pkgType, pkgName, pkgVersion string) {
	// Upload the test package to Artifactory if not already done
	if testPackageName == "" {
		uploadPackageToArtifactory(t)
	}
	return testPackageType, testPackageName, testPackageVersion
}

func uploadPackageToArtifactory(t *testing.T) {
	createNpmRepo(t)

	// Get the absolute path to the testdata file
	_, testFilePath, _, _ := runtime.Caller(0)
	npmPackageFilePath := filepath.Join(filepath.Dir(testFilePath), "testdata", "pizza-frontend.tgz")

	servicesManager := getArtifactoryServicesManager(t)
	uploadParams := services.NewUploadParams()
	uploadParams.Pattern = npmPackageFilePath
	uploadParams.Target = testRepoKey + "/pizza-frontend.tgz"
	uploadParams.Flat = true
	uploaded, failed, err := servicesManager.UploadFiles(artifactory.UploadServiceOptions{FailFast: false}, uploadParams)
	require.NoError(t, err)
	require.Equal(t, 1, uploaded, "Expected exactly one uploaded file")
	require.Equal(t, 0, failed, "Expected zero failed uploads")

	testPackageType = "npm"
	testPackageName = "@gpizza/pizza-frontend"
	testPackageVersion = "1.0.0"

	// Wait for the package to be indexed in Artifactory
	waitForPackageIndexing(t, testPackageName, testPackageVersion)
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

type packagesResponse struct {
	Packages []packageBinding `json:"packages"`
}

type packageBinding struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	NumVersions   int    `json:"num_versions"`
	LatestVersion string `json:"latest_version"`
}

func getPackageBindings(appKey string) (*packagesResponse, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s/packages", appKey)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint)
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var packagesRes *packagesResponse
	err = json.Unmarshal(responseBody, &packagesRes)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return packagesRes, statusCode, nil
}

func waitForPackageIndexing(t *testing.T, packageName, packageVersion string) {
	found := false
	timeout := time.After(5 * time.Minute)
	log.Info(fmt.Sprintf("Waiting up to 5 minutes for package indexing on %s", serverDetails.Url))

	query := fmt.Sprintf(`{"query": "query { versions (first: 100, filter: {name: \"%s\", repositoriesIn: [{name: \"%s\"}]}) { edges { node { package { name }}}}}"}`, packageVersion, testRepoKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for !found {
		select {
		case <-timeout:
			log.Warn("Timeout reached waiting for package indexing")
			require.FailNow(t, "Package indexing timeout: package %s was not indexed within 5 minutes", packageName)
		default:
			metadataUrl := serverDetails.Url + "metadata/api/v1/query"
			req, err := http.NewRequest(http.MethodPost, metadataUrl, strings.NewReader(query))
			if err != nil {
				log.Error("Error creating request:", err)
				break
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+serverDetails.AccessToken)

			resp, err := client.Do(req)
			if err != nil {
				log.Error("Error querying packages:", err)
				break
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error("Error reading response body:", err)
				break
			}
			err = resp.Body.Close()
			if err != nil {
				log.Debug("Error reading response body:", err)
				break
			}

			stringBody := string(body)
			if strings.Contains(stringBody, packageName) {
				log.Info(fmt.Sprintf("Package %s found and indexed", packageName))
				found = true
			} else {
				log.Debug(fmt.Sprintf("Package %s not found yet, retrying in 2 seconds", packageName))
				time.Sleep(2 * time.Second)
			}
		}
	}
}
