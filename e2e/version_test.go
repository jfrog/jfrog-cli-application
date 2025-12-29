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

	packageType, packageName, packageVersion, _ := getTestPackage(t)
	version := "1.0.0"

	// Execute
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s", packageType, packageName, packageVersion, testRepoKey)
	err := AppTrustCli.Exec("vc", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versions, statusCode, err := getApplicationVersions(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versions)
	assert.Len(t, versions.Versions, 1)
	assert.Equal(t, appKey, versions.Versions[0].ApplicationKey)
	assert.Equal(t, version, versions.Versions[0].Version)
	assert.Equal(t, "COMPLETED", versions.Versions[0].Status)
}

func TestCreateVersion_Artifact(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create-artifact")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	_, _, _, packagePath := getTestPackage(t)
	version := "1.0.1"

	// Execute
	artifactFlag := fmt.Sprintf("--source-type-artifacts=path=%s", packagePath)
	err := AppTrustCli.Exec("vc", appKey, version, artifactFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versions, statusCode, err := getApplicationVersions(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versions)
	assert.Len(t, versions.Versions, 1)
	assert.Equal(t, appKey, versions.Versions[0].ApplicationKey)
	assert.Equal(t, version, versions.Versions[0].Version)
	assert.Equal(t, "COMPLETED", versions.Versions[0].Status)
}

func TestCreateVersion_ApplicationVersion(t *testing.T) {
	// Prepare - create source application with a version
	sourceAppKey := generateUniqueKey("app-source-version")
	createBasicApplication(t, sourceAppKey)
	defer deleteApplication(t, sourceAppKey)

	packageType, packageName, packageVersion, _ := getTestPackage(t)
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
	versions, statusCode, err := getApplicationVersions(targetAppKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versions)
	assert.Len(t, versions.Versions, 1)
	assert.Equal(t, targetAppKey, versions.Versions[0].ApplicationKey)
	assert.Equal(t, targetVersion, versions.Versions[0].Version)
	assert.Equal(t, "COMPLETED", versions.Versions[0].Status)
}
