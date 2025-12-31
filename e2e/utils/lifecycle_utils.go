//go:build e2e

package utils

import (
	"testing"

	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/lifecycle"
	lifecycleServices "github.com/jfrog/jfrog-client-go/lifecycle/services"
	"github.com/stretchr/testify/require"
)

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
