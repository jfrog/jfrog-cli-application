//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/lifecycle"
	"github.com/jfrog/jfrog-client-go/lifecycle/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVersion_Package(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create-package")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	testPackage := getTestPackage(t)
	version := "1.0.0"

	// Execute
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_Artifact(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create-artifact")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	testPackage := getTestPackage(t)
	version := "1.0.1"

	// Execute
	artifactFlag := fmt.Sprintf("--source-type-artifacts=path=%s", testPackage.packagePath)
	err := AppTrustCli.Exec("version-create", appKey, version, artifactFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_ApplicationVersion(t *testing.T) {
	// Prepare - create source application with a version
	sourceAppKey := generateUniqueKey("app-version-create-app-version")
	createBasicApplication(t, sourceAppKey)
	defer deleteApplication(t, sourceAppKey)

	testPackage := getTestPackage(t)
	sourceVersion := "1.0.2"
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", sourceAppKey, sourceVersion, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, sourceAppKey, sourceVersion)

	// Prepare - create target application
	targetAppKey := generateUniqueKey("app-target-version")
	createBasicApplication(t, targetAppKey)
	defer deleteApplication(t, targetAppKey)

	targetVersion := "1.0.3"

	// Execute
	appVersionFlag := fmt.Sprintf("--source-type-application-versions=application-key=%s, version=%s", sourceAppKey, sourceVersion)
	err = AppTrustCli.Exec("version-create", targetAppKey, targetVersion, appVersionFlag)
	require.NoError(t, err)
	defer deleteVersion(t, targetAppKey, targetVersion)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(targetAppKey, targetVersion)
	require.NoError(t, err)
	assertVersionContent(t, versionContent, statusCode, targetAppKey, targetVersion)
}

func TestCreateVersion_ReleaseBundle(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create-release-bundle")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	lcDetails, err := serverDetails.CreateLifecycleAuthConfig()
	require.NoError(t, err)
	serviceConfig, err := config.NewConfigBuilder().SetServiceDetails(lcDetails).Build()
	require.NoError(t, err)
	lifecycleManager, err := lifecycle.New(serviceConfig)
	require.NoError(t, err)

	projectKey := GetTestProjectKey(t)
	testPackage := getTestPackage(t)
	bundleName := generateUniqueKey("apptrust-cli-tests-rb")
	bundleVersion := "1.0.0"

	rbDetails := services.ReleaseBundleDetails{ReleaseBundleName: bundleName, ReleaseBundleVersion: bundleVersion}
	params := services.CommonOptionalQueryParams{
		ProjectKey: projectKey,
	}

	source := services.CreateFromPackagesSource{Packages: []services.PackageSource{
		{
			PackageName:    testPackage.packageName,
			PackageVersion: testPackage.packageVersion,
			PackageType:    testPackage.packageType,
			RepositoryKey:  testRepoKey,
		},
	}}
	err = lifecycleManager.CreateReleaseBundleFromPackages(rbDetails, params, "default-lifecycle-key", source)
	require.NoError(t, err)

	defer func() {
		err = lifecycleManager.DeleteReleaseBundleVersion(rbDetails, params)
		require.NoError(t, err)
	}()

	version := "1.0.9"

	// Execute
	releaseBundleFlag := fmt.Sprintf("--source-type-release-bundles=name=%s, version=%s, project-key=%s", bundleName, bundleVersion, projectKey)
	err = AppTrustCli.Exec("version-create", appKey, version, releaseBundleFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, versionContent, statusCode, appKey, version)
}

func TestCreateVersion_Build(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-create-build")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	// Make sure to upload a package associated with a build
	testPackage := getTestPackage(t)

	version := "1.0.10"

	// Execute
	buildInfoFlag := fmt.Sprintf("--source-type-builds=name=%s, id=%s", testPackage.buildName, testPackage.buildNumber)
	err := AppTrustCli.Exec("version-create", appKey, version, buildInfoFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Assert
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assertVersionContent(t, versionContent, statusCode, appKey, version)
}

func assertVersionContent(t *testing.T, versionContent *versionContentResponse, statusCode int, appKey, version string) {
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, appKey, versionContent.ApplicationKey)
	assert.Equal(t, version, versionContent.Version)
	assert.Equal(t, "COMPLETED", versionContent.Status)
	assert.Len(t, versionContent.Releasables, 1)
	assert.Equal(t, testPackageRes.packageType, versionContent.Releasables[0].PackageType)
	assert.Equal(t, testPackageRes.packageName, versionContent.Releasables[0].Name)
	assert.Equal(t, testPackageRes.packageVersion, versionContent.Releasables[0].Version)
	assert.Len(t, versionContent.Releasables[0].Artifacts, 1)
	assert.Contains(t, testPackageRes.packagePath, versionContent.Releasables[0].Artifacts[0].Path)
}

func TestUpdateVersion(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("app-version-update")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)

	testPackage := getTestPackage(t)
	version := "1.0.4"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Execute
	tag := "release-candidate"
	err = AppTrustCli.Exec("version-update", appKey, version, "--tag="+tag)
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

	testPackage := getTestPackage(t)
	version := "1.0.5"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)

	// Verify the version exists
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, versionContent)
	assert.Equal(t, version, versionContent.Version)

	// Execute
	err = AppTrustCli.Exec("version-delete", appKey, version)
	assert.NoError(t, err)

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

	testPackage := getTestPackage(t)
	version := "1.0.6"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Execute
	targetStage := "DEV"
	err = AppTrustCli.Exec("version-promote", appKey, version, targetStage)
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

	testPackage := getTestPackage(t)
	version := "1.0.7"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Execute
	err = AppTrustCli.Exec("version-release", appKey, version)
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

	testPackage := getTestPackage(t)
	version := "1.0.8"

	// Create a version first
	packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
		testPackage.packageType, testPackage.packageName, testPackage.packageVersion, testRepoKey)
	err := AppTrustCli.Exec("version-create", appKey, version, packageFlag)
	require.NoError(t, err)
	defer deleteVersion(t, appKey, version)

	// Promote to DEV
	targetStage := "DEV"
	err = AppTrustCli.Exec("version-promote", appKey, version, targetStage)
	require.NoError(t, err)

	// Verify it's in DEV
	versionContent, statusCode, err := getApplicationVersion(appKey, version)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, targetStage, versionContent.CurrentStage)

	// Execute
	err = AppTrustCli.Exec("version-rollback", appKey, version, targetStage)
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
