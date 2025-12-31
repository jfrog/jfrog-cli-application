//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindPackage(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("package-bind")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)
	testPackage := utils.GetTestPackage(t)

	// Execute
	err := utils.AppTrustCli.Exec("package-bind", appKey, testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion)
	require.NoError(t, err)

	// Assert
	response, statusCode, err := utils.GetPackageBindings(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, response)
	assert.Len(t, response.Packages, 1)
	assert.Equal(t, testPackage.PackageType, response.Packages[0].Type)
	assert.Equal(t, testPackage.PackageName, response.Packages[0].Name)
	assert.Equal(t, 1, response.Packages[0].NumVersions)
	assert.Equal(t, testPackage.PackageVersion, response.Packages[0].LatestVersion)
}

func TestUnbindPackage(t *testing.T) {
	// Prepare
	appKey := utils.GenerateUniqueKey("package-unbind")
	utils.CreateBasicApplication(t, appKey)
	defer utils.DeleteApplication(t, appKey)
	testPackage := utils.GetTestPackage(t)

	// First bind the package
	err := utils.AppTrustCli.Exec("package-bind", appKey, testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion)
	require.NoError(t, err)

	// Verify it's bound
	bindings, statusCode, err := utils.GetPackageBindings(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, bindings)
	assert.Len(t, bindings.Packages, 1)

	// Unbind the package
	err = utils.AppTrustCli.Exec("package-unbind", appKey, testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion)
	require.NoError(t, err)

	// Verify the package is no longer bound
	bindings, statusCode, err = utils.GetPackageBindings(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, bindings)
	assert.Len(t, bindings.Packages, 0)
}
