//go:build itest

package application

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	appCmd "github.com/jfrog/jfrog-cli-application/apptrust/commands/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestProjectKey returns the test project key from environment variable or default
func getTestProjectKey() string {
	if projectKey := os.Getenv("JF_TEST_PROJECT_KEY"); projectKey != "" {
		return projectKey
	}
	return "default" // Default project key - should exist in test platform
}

func TestCreateApp_WithFlags_Minimal(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-minimal")

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: getTestProjectKey(),
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	require.NoError(t, err, "Failed to create application with minimal flags")

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})
}

func TestCreateApp_WithFlags_Full(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-full")

	description := "Integration test application with all fields"
	businessCriticality := "high"
	maturityLevel := "production"

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                            serverDetails.Url,
		"access-token":                   serverDetails.AccessToken,
		commands.ProjectFlag:             getTestProjectKey(),
		commands.ApplicationNameFlag:     "Test App Full",
		commands.DescriptionFlag:         description,
		commands.BusinessCriticalityFlag: businessCriticality,
		commands.MaturityLevelFlag:       maturityLevel,
		commands.LabelsFlag:              "env=test;region=us-east",
		commands.UserOwnersFlag:          "admin",
		commands.GroupOwnersFlag:         "developers",
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	require.NoError(t, err, "Failed to create application with all flags")

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})
}

func TestCreateApp_WithSpecFile_Minimal(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-spec-min")

	// Create a temporary minimal spec file
	specContent := `{
  "project_key": "` + getTestProjectKey() + `"
}`
	specFile := createTempSpecFile(t, specContent)
	defer os.Remove(specFile)

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":             serverDetails.Url,
		"access-token":    serverDetails.AccessToken,
		commands.SpecFlag: specFile,
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	require.NoError(t, err, "Failed to create application with minimal spec file")

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})
}

func TestCreateApp_WithSpecFile_Full(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-spec-full")

	// Create a temporary full spec file
	specContent := `{
  "project_key": "` + getTestProjectKey() + `",
  "application_name": "Test App Full Spec",
  "description": "A comprehensive test application",
  "maturity_level": "production",
  "criticality": "high",
  "labels": {
    "environment": "test",
    "region": "us-east-1",
    "team": "qa"
  },
  "user_owners": [
    "admin"
  ],
  "group_owners": [
    "developers"
  ]
}`
	specFile := createTempSpecFile(t, specContent)
	defer os.Remove(specFile)

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":             serverDetails.Url,
		"access-token":    serverDetails.AccessToken,
		commands.SpecFlag: specFile,
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	require.NoError(t, err, "Failed to create application with full spec file")

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})
}

func TestCreateApp_WithSpecFile_WithVars(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-spec-vars")

	// Create a temporary spec file with variables
	specContent := `{
  "project_key": "${PROJECT_KEY}",
  "application_name": "${APP_NAME}",
  "description": "A test application for ${ENVIRONMENT}",
  "maturity_level": "${MATURITY_LEVEL}",
  "criticality": "${CRITICALITY}",
  "labels": {
    "environment": "${ENVIRONMENT}",
    "region": "${REGION}"
  }
}`
	specFile := createTempSpecFile(t, specContent)
	defer os.Remove(specFile)

	specVars := "PROJECT_KEY=" + getTestProjectKey() + ";APP_NAME=Test App Vars;ENVIRONMENT=test;MATURITY_LEVEL=production;CRITICALITY=high;REGION=us-east-1"

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                 serverDetails.Url,
		"access-token":        serverDetails.AccessToken,
		commands.SpecFlag:     specFile,
		commands.SpecVarsFlag: specVars,
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	require.NoError(t, err, "Failed to create application with spec file and variables")

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})
}

func TestCreateApp_DuplicateAppKey(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-dup")

	// Create the app first
	ctx1 := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: getTestProjectKey(),
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx1)
	require.NoError(t, err, "Failed to create initial application")

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})

	// Try to create the same app again - should fail
	ctx2 := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: getTestProjectKey(),
	})

	err = cmd.Action(ctx2)
	assert.Error(t, err, "Expected error when creating duplicate application")
	assert.Contains(t, err.Error(), "failed to create an application", "Error should indicate creation failure")
}

func TestCreateApp_InvalidProject(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-invalid-proj")

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: "non-existent-project-key-12345-does-not-exist",
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	assert.Error(t, err, "Expected error when creating app with invalid project")
	assert.Contains(t, err.Error(), "failed to create an application", "Error should indicate creation failure")
}

func TestCreateApp_MissingProject(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-no-proj")

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":          serverDetails.Url,
		"access-token": serverDetails.AccessToken,
		// Intentionally missing project flag
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	assert.Error(t, err, "Expected error when project flag is missing")
	assert.Contains(t, err.Error(), "--project is mandatory", "Error should indicate project is mandatory")
}

func TestCreateApp_MissingAppKey(t *testing.T) {
	serverDetails := setupTest(t)

	ctx := createTestContext([]string{}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: getTestProjectKey(),
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	assert.Error(t, err, "Expected error when application-key argument is missing")
	assert.Contains(t, err.Error(), "Wrong number of arguments", "Error should indicate wrong number of arguments")
}

func TestCreateApp_SpecAndFlagsTogether(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-conflict")

	specContent := `{
  "project_key": "` + getTestProjectKey() + `"
}`
	specFile := createTempSpecFile(t, specContent)
	defer os.Remove(specFile)

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.SpecFlag:    specFile,
		commands.ProjectFlag: getTestProjectKey(), // This should cause an error
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	assert.Error(t, err, "Expected error when using both spec and flags")
	assert.Contains(t, err.Error(), "not allowed when --spec is provided", "Error should indicate spec and flags conflict")
}

func TestCreateApp_InvalidSpecFile(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-invalid-spec")

	nonExistentFile := filepath.Join(t.TempDir(), "non-existent-spec.json")

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":             serverDetails.Url,
		"access-token":    serverDetails.AccessToken,
		commands.SpecFlag: nonExistentFile,
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	assert.Error(t, err, "Expected error when spec file does not exist")
	assert.Contains(t, err.Error(), "no such file or directory", "Error should indicate file not found")
}

func TestCreateApp_InvalidSpecFormat(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-bad-json")

	// Create a malformed JSON spec file
	specContent := `{
  "project_key": "` + getTestProjectKey() + `"
  // Missing closing brace and invalid JSON
`
	specFile := createTempSpecFile(t, specContent)
	defer os.Remove(specFile)

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":             serverDetails.Url,
		"access-token":    serverDetails.AccessToken,
		commands.SpecFlag: specFile,
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	assert.Error(t, err, "Expected error when spec file has invalid JSON")
	assert.Contains(t, err.Error(), "invalid character", "Error should indicate JSON parsing error")
}

func TestCreateApp_DefaultApplicationName(t *testing.T) {
	serverDetails := setupTest(t)
	appKey := generateUniqueAppKey(t, "test-app-default-name")

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: getTestProjectKey(),
		// Intentionally not setting application-name flag
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	require.NoError(t, err, "Failed to create application without application-name")

	// The application should be created with application-name defaulting to application-key
	// We can't easily verify this without a get API, but the creation should succeed

	// Cleanup
	t.Cleanup(func() {
		cleanupApp(t, appKey, serverDetails)
	})
}

func TestCreateApp_SpecialCharacters(t *testing.T) {
	serverDetails := setupTest(t)
	// Test with app key containing hyphens and underscores (common special characters)
	appKey := generateUniqueAppKey(t, "test-app-special-chars") + "-with_underscores"

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":                serverDetails.Url,
		"access-token":       serverDetails.AccessToken,
		commands.ProjectFlag: getTestProjectKey(),
	})

	appCtx := createAppContext()
	cmd := appCmd.GetCreateAppCommand(appCtx)

	err := cmd.Action(ctx)
	// This might succeed or fail depending on platform validation rules
	// We just verify the command executes without panicking
	if err != nil {
		// If it fails, it should be a validation error, not a panic
		assert.Contains(t, err.Error(), "failed to create an application", "Error should be a creation failure, not a panic")
	} else {
		// Cleanup only if creation succeeded
		t.Cleanup(func() {
			cleanupApp(t, appKey, serverDetails)
		})
	}
}

// createTempSpecFile creates a temporary spec file with the given content
func createTempSpecFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "spec.json")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to create temporary spec file")
	return tmpFile
}
