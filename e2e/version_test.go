//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVersion_Package(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-create-package")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.0"

	// Execute
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, testPackage, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_Artifact(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-create-artifact")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.1"

	// Execute
	artifactFlag := fmt.Sprintf("--source-type-artifacts=path=%s", testPackage.PackagePath)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, artifactFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, testPackage, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_ApplicationVersion(t *testing.T) {
	// Prepare - create source application with a version
	sourceAppKey := utils.GenerateUniqueKey("app-version-create-app-version")
	utils.CreateBasicApplication(t, sourceAppKey)
	defer utils.DeleteApplication(t, sourceAppKey)

	testPackage := utils.GetTestPackage(t)
	sourceVersion := "1.0.2"
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", sourceAppKey, sourceVersion, packageFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, sourceAppKey, sourceVersion)

	// Prepare - create target application
	targetAppKey := utils.GenerateUniqueKey("app-target-version")
	utils.CreateBasicApplication(t, targetAppKey)
	defer utils.DeleteApplication(t, targetAppKey)

	targetVersion := "1.0.3"

	// Execute
	appVersionFlag := fmt.Sprintf("--source-type-application-versions=application-key=%s, version=%s", sourceAppKey, sourceVersion)
	err = utils.AppTrustCli.Exec("version-create", targetAppKey, targetVersion, appVersionFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, targetAppKey, targetVersion)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(targetAppKey, targetVersion)
	require.NoError(t, err)
	assertVersionContent(t, testPackage, versionContent, statusCode, targetAppKey, targetVersion)
}

func TestCreateVersion_ReleaseBundle(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-create-release-bundle")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	projectKey := utils.GetTestProjectKey(t)
	testPackage := utils.GetTestPackage(t)

	bundleName, bundleVersion, cleanup := utils.CreateReleaseBundle(t, projectKey, testPackage)
	defer cleanup()

	version := "1.0.9"

	// Execute
	releaseBundleFlag := fmt.Sprintf("--source-type-release-bundles=name=%s, version=%s, project-key=%s", bundleName, bundleVersion, projectKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, releaseBundleFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, testPackage, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_Build(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-create-build")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	// Make sure to upload a package associated with a build
	testPackage := utils.GetTestPackage(t)

	version := "1.0.10"

	// Execute
	buildInfoFlag := fmt.Sprintf("--source-type-builds=name=%s, id=%s", testPackage.BuildName, testPackage.BuildNumber)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, buildInfoFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, testPackage, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_Draft(t *testing.T) {
	t.Skip("Skipping draft version creation test")
	appKey := utils.GenerateUniqueKey("app-version-create-draft")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := utils.GenerateUniqueKey("draft")

	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag, "--draft")
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, utils.StatusDraft, versionContent.Status)
}

func TestCreateVersion_Async(t *testing.T) {
	appKey := utils.GenerateUniqueKey("app-version-create-async")
	utils.CreateBasicApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := utils.GenerateUniqueKey("async")
	defer utils.DeleteApplication(t, appKey)

	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)

	output := utils.AppTrustCli.RunCliCmdWithOutput(t, "version-create", appKey, version, packageFlag, "--sync=false")
	assert.NotEmpty(t, output)
	defer func() {
		utils.WaitForVersionCreation(t, appKey, version, 3*time.Second, 500*time.Millisecond)
		utils.DeleteApplicationVersion(t, appKey, version)
	}()

	var response struct {
		Status string `json:"status"`
	}
	err := json.Unmarshal([]byte(output), &response)
	require.NoError(t, err, "failed to parse CLI output as JSON: %s", output)
	assert.Contains(t, []string{utils.StatusInProgress, utils.StatusStarted}, response.Status)
}

func assertVersionContent(t *testing.T, expectedPackage *utils.TestPackageResources, versionContent *utils.VersionContentResponse, statusCode int, appKey, appVersion string) {
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, appVersion, versionContent.Version)
	assert.Equal(t, utils.StatusCompleted, versionContent.Status)
	assert.Len(t, versionContent.Releasables, 1)
	assert.Equal(t, expectedPackage.PackageType, versionContent.Releasables[0].PackageType)
	assert.Equal(t, expectedPackage.PackageName, versionContent.Releasables[0].Name)
	assert.Equal(t, expectedPackage.PackageVersion, versionContent.Releasables[0].Version)
	assert.Len(t, versionContent.Releasables[0].Artifacts, 1)
	assert.Contains(t, expectedPackage.PackagePath, versionContent.Releasables[0].Artifacts[0].Path)
}

func TestUpdateVersion(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-update")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.4"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Execute
	tag := "release-candidate"
	err = utils.AppTrustCli.Exec("version-update", appKey, version, "--tag="+tag)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, tag, versionContent.Tag)
}

func TestUpdateDraftVersionSources(t *testing.T) {
	appKey := utils.GenerateUniqueKey("app-version-update-sources")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)
	testPackage := utils.GetTestPackage(t)
	version := "1.0.6"
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag, "--draft")
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)
	artifactPath := utils.GetTestArtifact(t)
	artifactFlag := fmt.Sprintf("--source-type-artifacts=path=%s", artifactPath)

	err = utils.AppTrustCli.Exec("version-update-sources", appKey, version, artifactFlag)
	require.NoError(t, err)

	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Contains(t, utils.StatusDraft, versionContent.Status)
	var artifactPaths []string
	for _, r := range versionContent.Releasables {
		for _, a := range r.Artifacts {
			artifactPaths = append(artifactPaths, a.Path)
		}
	}
	assert.True(t, containsPath(artifactPaths, testPackage.PackagePath),
		"expected package path %q in version releasables (got %v)", testPackage.PackagePath, artifactPaths)
	assert.True(t, containsPath(artifactPaths, artifactPath),
		"expected artifact path %q in version releasables (got %v)", artifactPath, artifactPaths)
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if strings.Contains(target, path) {
			return true
		}
	}
	return false
}

func TestDeleteVersion(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-delete")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.5"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)

	// Verify the version exists
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, version, versionContent.Version)

	// Execute
	err = utils.AppTrustCli.Exec("version-delete", appKey, version)
	assert.NoError(t, err)

	// Assert
	_, statusCode, err = utils.GetApplicationVersion(appKey, version)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)
}

func TestPromoteVersion(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-promote")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.6"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Execute
	targetStage := "DEV"
	err = utils.AppTrustCli.Exec("version-promote", appKey, version, targetStage)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, targetStage, versionContent.CurrentStage)
}

func TestReleaseVersion(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-release")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.7"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Execute
	err = utils.AppTrustCli.Exec("version-release", appKey, version)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, "PROD", versionContent.CurrentStage)
}

func TestRollbackVersion(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("app-version-rollback")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)

	testPackage := utils.GetTestPackage(t)
	version := "1.0.8"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
	err := utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer utils.DeleteApplicationVersion(t, appKey, version)

	// Promote to DEV
	targetStage := "DEV"
	err = utils.AppTrustCli.Exec("version-promote", appKey, version, targetStage)
	require.NoError(t, err)

	// Verify it's in DEV
	versionContent, statusCode, err := utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, targetStage, versionContent.CurrentStage)

	// Execute
	err = utils.AppTrustCli.Exec("version-rollback", appKey, version, targetStage)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err = utils.GetApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Empty(t, versionContent.CurrentStage)
}
