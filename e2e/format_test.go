//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertValidJSON parses output as a JSON object and returns the decoded map.
func assertValidJSON(t *testing.T, output string) map[string]interface{} {
	t.Helper()
	require.NotEmpty(t, strings.TrimSpace(output), "expected non-empty output")
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &parsed), "expected valid JSON, got: %s", output)
	return parsed
}

// assertTableOutput verifies the output is a FIELD/VALUE table containing all expectedFields.
func assertTableOutput(t *testing.T, output string, expectedFields ...string) {
	t.Helper()
	require.Contains(t, output, "FIELD", "table output should have a FIELD column header")
	require.Contains(t, output, "VALUE", "table output should have a VALUE column header")
	for _, field := range expectedFields {
		assert.Contains(t, output, field, "table output should contain field %q", field)
	}
}

func TestAppCreate_OutputFormat(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)

	runAppCreate := func(appKey string, formatArgs ...string) string {
		args := []string{"app-create", appKey, "--project=" + projectKey, "--application-name=" + appKey}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey := utils.GenerateUniqueKey("app-create-fmt-default")
		defer utils.DeleteApplication(t, appKey)
		parsed := assertValidJSON(t, runAppCreate(appKey))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey := utils.GenerateUniqueKey("app-create-fmt-json")
		defer utils.DeleteApplication(t, appKey)
		parsed := assertValidJSON(t, runAppCreate(appKey, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey := utils.GenerateUniqueKey("app-create-fmt-table")
		defer utils.DeleteApplication(t, appKey)
		assertTableOutput(t, runAppCreate(appKey, "--format=table"), "application_key", "application_name", "project_key")
	})
}

func TestAppUpdate_OutputFormat(t *testing.T) {
	runAppUpdate := func(appKey string, formatArgs ...string) string {
		args := []string{"app-update", appKey, "--application-name=Updated " + appKey}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey := utils.GenerateUniqueKey("app-update-fmt-default")
		utils.CreateBasicApplication(t, appKey)
		defer utils.DeleteApplication(t, appKey)
		parsed := assertValidJSON(t, runAppUpdate(appKey))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey := utils.GenerateUniqueKey("app-update-fmt-json")
		utils.CreateBasicApplication(t, appKey)
		defer utils.DeleteApplication(t, appKey)
		parsed := assertValidJSON(t, runAppUpdate(appKey, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey := utils.GenerateUniqueKey("app-update-fmt-table")
		utils.CreateBasicApplication(t, appKey)
		defer utils.DeleteApplication(t, appKey)
		assertTableOutput(t, runAppUpdate(appKey, "--format=table"), "application_key", "application_name")
	})
}

func TestVersionUpdate_OutputFormat(t *testing.T) {
	testPackage := utils.GetTestPackage(t)

	prepareVersion := func(t *testing.T, suffix string) (appKey, version string, cleanup func()) {
		appKey = utils.GenerateUniqueKey("version-update-fmt-" + suffix)
		utils.CreateBasicApplication(t, appKey)
		version = "1.0.0"
		packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
			testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
		require.NoError(t, utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag))
		return appKey, version, func() {
			utils.DeleteApplicationVersion(t, appKey, version)
			utils.DeleteApplication(t, appKey)
		}
	}

	runUpdate := func(appKey, version string, formatArgs ...string) string {
		args := []string{"version-update", appKey, version, "--tag=fmt-tag"}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "default")
		defer cleanup()
		parsed := assertValidJSON(t, runUpdate(appKey, version))
		assert.Equal(t, appKey, parsed["application_key"])
		assert.Equal(t, version, parsed["version"])
	})

	t.Run("json", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "json")
		defer cleanup()
		parsed := assertValidJSON(t, runUpdate(appKey, version, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "table")
		defer cleanup()
		assertTableOutput(t, runUpdate(appKey, version, "--format=table"), "application_key", "version")
	})
}

func TestVersionUpdateSources_OutputFormat(t *testing.T) {
	testPackage := utils.GetTestPackage(t)
	artifactRepo, artifactFile := utils.GetTestArtifact(t)
	artifactPath := artifactRepo + "/" + artifactFile

	prepareDraftVersion := func(t *testing.T, suffix string) (appKey, version string, cleanup func()) {
		appKey = utils.GenerateUniqueKey("version-upd-src-fmt-" + suffix)
		utils.CreateBasicApplication(t, appKey)
		version = "1.0.0"
		packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
			testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
		require.NoError(t, utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag, "--draft"))
		return appKey, version, func() {
			utils.DeleteApplicationVersion(t, appKey, version)
			utils.DeleteApplication(t, appKey)
		}
	}

	runUpdateSources := func(appKey, version string, formatArgs ...string) string {
		args := []string{"version-update-sources", appKey, version, "--source-type-artifacts=path=" + artifactPath}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey, version, cleanup := prepareDraftVersion(t, "default")
		defer cleanup()
		parsed := assertValidJSON(t, runUpdateSources(appKey, version))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey, version, cleanup := prepareDraftVersion(t, "json")
		defer cleanup()
		parsed := assertValidJSON(t, runUpdateSources(appKey, version, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey, version, cleanup := prepareDraftVersion(t, "table")
		defer cleanup()
		assertTableOutput(t, runUpdateSources(appKey, version, "--format=table"), "application_key", "version")
	})
}

func TestVersionPromote_OutputFormat(t *testing.T) {
	testPackage := utils.GetTestPackage(t)

	prepareVersion := func(t *testing.T, suffix string) (appKey, version string, cleanup func()) {
		appKey = utils.GenerateUniqueKey("version-promote-fmt-" + suffix)
		utils.CreateBasicApplication(t, appKey)
		version = "1.0.0"
		packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
			testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
		require.NoError(t, utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag))
		return appKey, version, func() {
			utils.DeleteApplicationVersion(t, appKey, version)
			utils.DeleteApplication(t, appKey)
		}
	}

	runPromote := func(appKey, version string, formatArgs ...string) string {
		args := []string{"version-promote", appKey, version, "DEV"}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "default")
		defer cleanup()
		parsed := assertValidJSON(t, runPromote(appKey, version))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "json")
		defer cleanup()
		parsed := assertValidJSON(t, runPromote(appKey, version, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "table")
		defer cleanup()
		assertTableOutput(t, runPromote(appKey, version, "--format=table"), "application_key", "version", "target_stage")
	})
}

func TestVersionRelease_OutputFormat(t *testing.T) {
	testPackage := utils.GetTestPackage(t)

	prepareVersion := func(t *testing.T, suffix string) (appKey, version string, cleanup func()) {
		appKey = utils.GenerateUniqueKey("version-release-fmt-" + suffix)
		utils.CreateBasicApplication(t, appKey)
		version = "1.0.0"
		packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
			testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
		require.NoError(t, utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag))
		return appKey, version, func() {
			utils.DeleteApplicationVersion(t, appKey, version)
			utils.DeleteApplication(t, appKey)
		}
	}

	runRelease := func(appKey, version string, formatArgs ...string) string {
		args := []string{"version-release", appKey, version}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "default")
		defer cleanup()
		parsed := assertValidJSON(t, runRelease(appKey, version))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "json")
		defer cleanup()
		parsed := assertValidJSON(t, runRelease(appKey, version, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey, version, cleanup := prepareVersion(t, "table")
		defer cleanup()
		assertTableOutput(t, runRelease(appKey, version, "--format=table"), "application_key", "version")
	})
}

func TestVersionRollback_OutputFormat(t *testing.T) {
	testPackage := utils.GetTestPackage(t)

	preparePromoted := func(t *testing.T, suffix string) (appKey, version string, cleanup func()) {
		appKey = utils.GenerateUniqueKey("version-rollback-fmt-" + suffix)
		utils.CreateBasicApplication(t, appKey)
		version = "1.0.0"
		packageFlag := fmt.Sprintf("--source-type-packages=type=%s, name=%s, version=%s, repo-key=%s",
			testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion, testPackage.RepoKey)
		require.NoError(t, utils.AppTrustCli.Exec("version-create", appKey, version, packageFlag))
		require.NoError(t, utils.AppTrustCli.Exec("version-promote", appKey, version, "DEV"))
		return appKey, version, func() {
			utils.DeleteApplicationVersion(t, appKey, version)
			utils.DeleteApplication(t, appKey)
		}
	}

	runRollback := func(appKey, version string, formatArgs ...string) string {
		args := []string{"version-rollback", appKey, version, "DEV"}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey, version, cleanup := preparePromoted(t, "default")
		defer cleanup()
		parsed := assertValidJSON(t, runRollback(appKey, version))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey, version, cleanup := preparePromoted(t, "json")
		defer cleanup()
		parsed := assertValidJSON(t, runRollback(appKey, version, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey, version, cleanup := preparePromoted(t, "table")
		defer cleanup()
		assertTableOutput(t, runRollback(appKey, version, "--format=table"), "application_key", "version", "rollback_from_stage")
	})
}

func TestPackageBind_OutputFormat(t *testing.T) {
	testPackage := utils.GetTestPackage(t)

	prepareApp := func(t *testing.T, suffix string) (appKey string, cleanup func()) {
		appKey = utils.GenerateUniqueKey("package-bind-fmt-" + suffix)
		utils.CreateBasicApplication(t, appKey)
		return appKey, func() { utils.DeleteApplication(t, appKey) }
	}

	runBind := func(appKey string, formatArgs ...string) string {
		args := []string{"package-bind", appKey, testPackage.PackageType, testPackage.PackageName, testPackage.PackageVersion}
		args = append(args, formatArgs...)
		return utils.AppTrustCli.RunCliCmdWithOutput(t, args...)
	}

	t.Run("default", func(t *testing.T) {
		appKey, cleanup := prepareApp(t, "default")
		defer cleanup()
		parsed := assertValidJSON(t, runBind(appKey))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("json", func(t *testing.T) {
		appKey, cleanup := prepareApp(t, "json")
		defer cleanup()
		parsed := assertValidJSON(t, runBind(appKey, "--format=json"))
		assert.Equal(t, appKey, parsed["application_key"])
	})

	t.Run("table", func(t *testing.T) {
		appKey, cleanup := prepareApp(t, "table")
		defer cleanup()
		assertTableOutput(t, runBind(appKey, "--format=table"), "application_key", "package_type", "package_name", "package_version")
	})
}
