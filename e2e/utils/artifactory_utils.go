//go:build e2e

package utils

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	artClientUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/require"
)

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

var genericRepoKey string

func createGenericRepo(t *testing.T) string {
	servicesManager := getArtifactoryServicesManager(t)
	genericRepoKey = GetTestProjectKey(t) + "-generic-local"
	localRepoConfig := services.NewGenericLocalRepositoryParams()
	localRepoConfig.ProjectKey = GetTestProjectKey(t)
	localRepoConfig.Key = genericRepoKey
	err := servicesManager.CreateLocalRepository().Generic(localRepoConfig)
	require.NoError(t, err)
	return genericRepoKey
}

func deleteGenericRepo() {
	if genericRepoKey == "" || artifactoryServicesManager == nil {
		return
	}

	err := artifactoryServicesManager.DeleteRepository(genericRepoKey)
	if err != nil {
		log.Error("Failed to delete generic repo", err)
	}
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

func uploadPackageToArtifactory(t *testing.T, repoKey, buildName, buildNumber string) (sha256 string) {
	// Get the absolute path to the testdata file
	_, testFilePath, _, _ := runtime.Caller(0)
	npmPackageFilePath := filepath.Join(filepath.Dir(testFilePath), "testdata", "pizza-frontend.tgz")

	targetPath := repoKey + "/pizza-frontend.tgz"
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

	testPackageRes = &TestPackageResources{
		PackageType:    "npm",
		PackageName:    packageName,
		PackageVersion: packageVersion,
		PackagePath:    targetPath,
		RepoKey:        repoKey,
		BuildName:      buildName,
		BuildNumber:    buildNumber,
	}

	// Reindex the repo for the package to be available
	reindexRepo(t, repoKey)

	return artifactDetails.Checksums.Sha256
}

func uploadSimpleFileToArtifactory(t *testing.T, repoKey, targetFileName string) string {
	tmpFile, err := os.CreateTemp("", "e2e-artifact-*.txt")
	require.NoError(t, err)
	_, err = tmpFile.WriteString("test-artifact-content")
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	targetPath := repoKey + "/" + targetFileName
	servicesManager := getArtifactoryServicesManager(t)
	uploadParams := services.NewUploadParams()
	uploadParams.Pattern = tmpFile.Name()
	uploadParams.Target = targetPath
	uploadParams.Flat = true
	summary, err := servicesManager.UploadFilesWithSummary(artifactory.UploadServiceOptions{FailFast: false}, uploadParams)
	require.NoError(t, err)
	require.Equal(t, 1, summary.TotalSucceeded, "Expected exactly one uploaded file")
	require.Equal(t, 0, summary.TotalFailed, "Expected zero failed uploads")
	err = summary.Close()
	require.NoError(t, err)

	return targetPath
}

func reindexRepo(t *testing.T, repoKey string) {
	log.Info(fmt.Sprintf("Reindexing repository %s", repoKey))

	query := fmt.Sprintf(`{"paths": ["%s"]}`, repoKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	metadataUrl := serverDetails.GetArtifactoryUrl() + "api/metadata_server/reindex?async=false"
	req, err := http.NewRequest(http.MethodPost, metadataUrl, strings.NewReader(query))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+serverDetails.AccessToken)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
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
