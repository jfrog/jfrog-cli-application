//go:build e2e

package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	artClientUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/lifecycle"
	lifecycleServices "github.com/jfrog/jfrog-client-go/lifecycle/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func DeleteTestProject() {
	if testProjectKey == "" {
		return
	}
	deleteBuild()
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
	err := AppTrustCli.Exec("app-create", appKey, "--project="+projectKey, "--application-name="+appKey)
	assert.NoError(t, err)
}

func DeleteApplication(t *testing.T, appKey string) {
	err := AppTrustCli.Exec("app-delete", appKey)
	assert.NoError(t, err)
}

func DeleteVersion(t *testing.T, appKey, version string) {
	err := AppTrustCli.Exec("version-delete", appKey, version)
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
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint, nil)
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

func uploadPackageToArtifactory(t *testing.T) {
	repoKey := createNpmRepo(t)

	// Get the absolute path to the testdata file
	_, testFilePath, _, _ := runtime.Caller(0)
	npmPackageFilePath := filepath.Join(filepath.Dir(testFilePath), "testdata", "pizza-frontend.tgz")

	targetPath := repoKey + "/pizza-frontend.tgz"
	buildName := GenerateUniqueKey("apptrust-cli-tests-build")
	buildNumber := "1"
	buildProps, _ := build.CreateBuildProperties(buildName, buildNumber, "")

	servicesManager := getArtifactoryServicesManager(t)
	uploadParams := services.NewUploadParams()
	uploadParams.Pattern = npmPackageFilePath
	uploadParams.Target = targetPath
	uploadParams.Flat = true
	uploadParams.BuildProps = buildProps
	summary, err := servicesManager.UploadFilesWithSummary(artifactory.UploadServiceOptions{FailFast: false}, uploadParams)
	require.NoError(t, err)
	require.Equal(t, 1, summary.TotalSucceeded, "Expected exactly one uploaded file")
	require.Equal(t, 0, summary.TotalFailed, "Expected zero failed uploads")
	defer func() {
		err = summary.Close()
		require.NoError(t, err)
	}()

	artifactDetails := new(artClientUtils.ArtifactDetails)
	err = summary.ArtifactsDetailsReader.NextRecord(artifactDetails)
	require.NoError(t, err)

	packageName := "@gpizza/pizza-frontend"
	packageVersion := "1.0.0"

	// Wait for the package to be indexed in Artifactory
	waitForPackageIndexing(t, packageName, packageVersion, repoKey)

	testPackageRes = &TestPackageResources{
		PackageType:    "npm",
		PackageName:    packageName,
		PackageVersion: packageVersion,
		PackagePath:    targetPath,
		RepoKey:        repoKey,
		BuildName:      buildName,
		BuildNumber:    buildNumber,
	}

	publishBuild(t, buildName, buildNumber, artifactDetails.Checksums.Sha256)
}

func waitForPackageIndexing(t *testing.T, packageName, packageVersion, repoKey string) {
	found := false
	timeout := time.After(5 * time.Minute)
	log.Info(fmt.Sprintf("Waiting up to 5 minutes for package indexing on %s", serverDetails.Url))

	query := fmt.Sprintf(`{"query": "query { versions (first: 100, filter: {name: \"%s\", repositoriesIn: [{name: \"%s\"}]}) { edges { node { package { name }}}}}"}`, packageVersion, repoKey)

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

func createNpmRepo(t *testing.T) string {
	servicesManager := getArtifactoryServicesManager(t)
	repoKey := GetTestProjectKey(t) + "-npm-local"
	localRepoConfig := services.NewNpmLocalRepositoryParams()
	localRepoConfig.ProjectKey = GetTestProjectKey(t)
	localRepoConfig.Key = repoKey
	localRepoConfig.Environments = []string{"DEV", "PROD"}
	err := servicesManager.CreateLocalRepository().Npm(localRepoConfig)
	require.NoError(t, err)
	return repoKey
}

func deleteNpmRepo() {
	if testPackageRes == nil || artifactoryServicesManager == nil {
		return
	}

	err := artifactoryServicesManager.DeleteRepository(testPackageRes.RepoKey)
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

func CreateReleaseBundle(t *testing.T, projectKey string, testPackage *TestPackageResources) (bundleName, bundleVersion string, cleanup func()) {
	lcDetails, err := serverDetails.CreateLifecycleAuthConfig()
	require.NoError(t, err)
	serviceConfig, err := config.NewConfigBuilder().SetServiceDetails(lcDetails).Build()
	require.NoError(t, err)
	lifecycleManager, err := lifecycle.New(serviceConfig)
	require.NoError(t, err)

	bundleName = GenerateUniqueKey("apptrust-cli-tests-rb")
	bundleVersion = "1.0.0"

	rbDetails := lifecycleServices.ReleaseBundleDetails{ReleaseBundleName: bundleName, ReleaseBundleVersion: bundleVersion}
	params := lifecycleServices.CommonOptionalQueryParams{
		ProjectKey: projectKey,
	}

	source := lifecycleServices.CreateFromPackagesSource{Packages: []lifecycleServices.PackageSource{
		{
			PackageName:    testPackage.PackageName,
			PackageVersion: testPackage.PackageVersion,
			PackageType:    testPackage.PackageType,
			RepositoryKey:  testPackage.RepoKey,
		},
	}}
	err = lifecycleManager.CreateReleaseBundleFromPackages(rbDetails, params, "default-lifecycle-key", source)
	require.NoError(t, err)
	cleanup = func() {
		err = lifecycleManager.DeleteReleaseBundleVersion(rbDetails, params)
		require.NoError(t, err)
	}
	return
}

func publishBuild(t *testing.T, buildName, buildNumber, sha256 string) {
	buildInfo := &buildinfo.BuildInfo{
		Name:    buildName,
		Number:  buildNumber,
		Started: "2024-01-01T12:00:00.000Z",
		Modules: []buildinfo.Module{
			{
				Id: "build-module",
				Artifacts: []buildinfo.Artifact{
					{
						Name: testPackageRes.PackageName,
						Checksum: buildinfo.Checksum{
							Sha256: sha256,
						},
					},
				},
			},
		},
	}
	servicesManager := getArtifactoryServicesManager(t)
	summary, err := servicesManager.PublishBuildInfo(buildInfo, "")
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.True(t, summary.IsSucceeded())
}

func deleteBuild() {
	if testPackageRes == nil {
		return
	}

	err := artifactoryServicesManager.DeleteBuildInfo(&buildinfo.BuildInfo{Name: testPackageRes.BuildName, Number: testPackageRes.BuildNumber}, "", 1)
	if err != nil {
		log.Error("Failed to delete build-info", err)
	}
}

type PackagesResponse struct {
	Packages []packageBinding `json:"packages"`
}

type packageBinding struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	NumVersions   int    `json:"num_versions"`
	LatestVersion string `json:"latest_version"`
}

func GetPackageBindings(appKey string) (*PackagesResponse, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s/packages", appKey)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint, nil)
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var packagesRes *PackagesResponse
	err = json.Unmarshal(responseBody, &packagesRes)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return packagesRes, statusCode, nil
}

type VersionContentResponse struct {
	ApplicationKey string       `json:"application_key"`
	Version        string       `json:"version"`
	Status         string       `json:"status"`
	CurrentStage   string       `json:"current_stage,omitempty"`
	Tag            string       `json:"tag,omitempty"`
	Releasables    []releasable `json:"releasables"`
}

type releasable struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	PackageType string     `json:"package_type"`
	Artifacts   []artifact `json:"artifacts,omitempty"`
}

type artifact struct {
	Path string `json:"path"`
}

func GetApplicationVersion(appKey, version string) (*VersionContentResponse, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/content", appKey, version)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint, map[string]string{"include": "releasables_expanded"})
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var versionRes *VersionContentResponse
	err = json.Unmarshal(responseBody, &versionRes)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return versionRes, statusCode, nil
}
