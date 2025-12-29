package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindPackage(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("package-bind")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)
	packageType, packageName, packageVersion, _ := getTestPackage(t)

	// Execute
	err := AppTrustCli.Exec("pb", appKey, packageType, packageName, packageVersion)
	require.NoError(t, err)

	// Assert
	response, statusCode, err := getPackageBindings(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, response)
	assert.Len(t, response.Packages, 1)
	assert.Equal(t, packageType, response.Packages[0].Type)
	assert.Equal(t, packageName, response.Packages[0].Name)
	assert.Equal(t, 1, response.Packages[0].NumVersions)
	assert.Equal(t, packageVersion, response.Packages[0].LatestVersion)
}

func TestUnbindPackage(t *testing.T) {
	// Prepare
	appKey := generateUniqueKey("package-unbind")
	createBasicApplication(t, appKey)
	defer deleteApplication(t, appKey)
	packageType, packageName, packageVersion, _ := getTestPackage(t)

	// First bind the package
	err := AppTrustCli.Exec("pb", appKey, packageType, packageName, packageVersion)
	require.NoError(t, err)

	// Verify it's bound
	bindings, statusCode, err := getPackageBindings(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, bindings)
	assert.Len(t, bindings.Packages, 1)

	// Unbind the package
	err = AppTrustCli.Exec("pu", appKey, packageType, packageName, packageVersion)
	require.NoError(t, err)

	// Verify the package is no longer bound
	bindings, statusCode, err = getPackageBindings(appKey)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	require.NotNil(t, bindings)
	assert.Len(t, bindings.Packages, 0)
}
