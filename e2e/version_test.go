package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVersion_Package(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, packagePath := getTestPackage(t)
	version := "1.0.0"

	// Execute
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, "COMPLETED", versionContent.Status)
	assert.Len(t, versionContent.Releasables, 1)
	assert.Equal(t, packageType, versionContent.Releasables[0].PackageType)
	assert.Equal(t, packageName, versionContent.Releasables[0].Name)
	assert.Equal(t, packageVersion, versionContent.Releasables[0].Version)
	assert.Len(t, versionContent.Releasables[0].Artifacts, 1)
	assert.Contains(t, packagePath, versionContent.Releasables[0].Artifacts[0].Path)
}

func TestCreateVersion_Artifact(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create-artifact")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, packagePath := getTestPackage(t)
	version := "1.0.1"

	// Execute
	artifactFlag := fmt.Sprintf("--source-type-artifacts=path=%s", packagePath)
	err := AppTrustCli.Exec("vc", appKey, version, artifactFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, "COMPLETED", versionContent.Status)
	assert.Len(t, versionContent.Releasables, 1)
	assert.Equal(t, packageType, versionContent.Releasables[0].PackageType)
	assert.Equal(t, packageName, versionContent.Releasables[0].Name)
	assert.Equal(t, packageVersion, versionContent.Releasables[0].Version)
	assert.Len(t, versionContent.Releasables[0].Artifacts, 1)
	assert.Contains(t, packagePath, versionContent.Releasables[0].Artifacts[0].Path)
}

func TestCreateVersion_ApplicationVersion(t *testing.T) {
	// Prepare - create source application with a version
	sourceAppKey := generateUniqueKey("app-source-version")
	createBasicApplication(t, sourceAppKey)
	defer deleteApplication(t, sourceAppKey)

	packageType, packageName, packageVersion, packagePath := getTestPackage(t)
	sourceVersion := "1.0.2"
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", sourceAppKey, sourceVersion, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, sourceAppKey, sourceVersion)

	// Prepare - create target application
	targetAppKey := generateUniqueKey("app-target-version")
	createBasicApplication(t, targetAppKey)
	defer deleteApplication(t, targetAppKey)

	targetVersion := "1.0.3"

	// Execute
	appVersionFlag := fmt.Sprintf("--source-type-application-versions=application-key=%s, version=%s", sourceAppKey, sourceVersion)
	err = AppTrustCli.Exec("vc", targetAppKey, targetVersion, appVersionFlag)
	require.NoError(t, err)
	defer deleteVersion(t, targetAppKey, targetVersion)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(targetAppKey, targetVersion)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, targetAppKey, versionContent.ApplicationKey)
	assert.Equal(t, targetVersion, versionContent.Version)
	assert.Equal(t, "COMPLETED", versionContent.Status)
	assert.Len(t, versionContent.Releasables, 1)
	assert.Equal(t, packageType, versionContent.Releasables[0].PackageType)
	assert.Equal(t, packageName, versionContent.Releasables[0].Name)
	assert.Equal(t, packageVersion, versionContent.Releasables[0].Version)
	assert.Len(t, versionContent.Releasables[0].Artifacts, 1)
	assert.Contains(t, packagePath, versionContent.Releasables[0].Artifacts[0].Path)
}

func TestUpdateVersion(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-update")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, _ := getTestPackage(t)
	version := "1.0.4"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Execute
	tag := "release-candidate"
	err = AppTrustCli.Exec("vu", appKey, version, "--tag="+tag)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, tag, versionContent.Tag)
}

func TestDeleteVersion(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-delete")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, _ := getTestPackage(t)
	version := "1.0.5"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)

	// Verify the version exists
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, version, versionContent.Version)

	// Execute
	deleteVersion(t, appKey, version)

	// Assert
	_, statusCode, err = getApplicationVersion(appKey, version)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)
}

func TestPromoteVersion(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-promote")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, _ := getTestPackage(t)
	version := "1.0.6"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Execute
	targetStage := "DEV"
	err = AppTrustCli.Exec("vp", appKey, version, targetStage)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, targetStage, versionContent.CurrentStage)
}

func TestReleaseVersion(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-release")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, _ := getTestPackage(t)
	version := "1.0.7"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Execute
	err = AppTrustCli.Exec("vr", appKey, version)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, "PROD", versionContent.CurrentStage)
}

func TestRollbackVersion(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-rollback")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	packageType, packageName, packageVersion, _ := getTestPackage(t)
	version := "1.0.8"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Promote to DEV
	targetStage := "DEV"
	err = AppTrustCli.Exec("vp", appKey, version, targetStage)
	require.NoError(t, err)

	// Verify it's in DEV
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, targetStage, versionContent.CurrentStage)

	// Execute
	err = AppTrustCli.Exec("vrb", appKey, version, targetStage)
	require.NoError(t, err)

	// Assert
	versionContent, statusCode, err = getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Empty(t, versionContent.CurrentStage)
}
