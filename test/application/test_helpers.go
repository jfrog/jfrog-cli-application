//go:build itest

package application

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-application/apptrust/app"
	appCmd "github.com/jfrog/jfrog-cli-application/apptrust/commands/application"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/require"
)

const (
	envPlatformURL         = "JF_PLATFORM_URL"
	envPlatformAccessToken = "JF_PLATFORM_ACCESS_TOKEN"
)

// getServerDetailsFromEnv reads server details from environment variables
func getServerDetailsFromEnv(t *testing.T) *coreConfig.ServerDetails {
	url := os.Getenv(envPlatformURL)
	require.NotEmpty(t, url, "%s environment variable must be set", envPlatformURL)

	accessToken := os.Getenv(envPlatformAccessToken)
	require.NotEmpty(t, accessToken, "%s environment variable must be set", envPlatformAccessToken)

	return &coreConfig.ServerDetails{
		Url:         url,
		AccessToken: accessToken,
	}
}

// createTestContext creates a components.Context with the given arguments and flags
func createTestContext(args []string, flags map[string]string) *components.Context {
	ctx := &components.Context{
		Arguments: args,
	}

	for key, value := range flags {
		ctx.AddStringFlag(key, value)
	}

	return ctx
}

// createAppContext creates an app.Context instance
func createAppContext() app.Context {
	return app.NewAppContext()
}

// setupTest validates environment variables are set and returns server details
func setupTest(t *testing.T) *coreConfig.ServerDetails {
	return getServerDetailsFromEnv(t)
}

// cleanupApp deletes an application using the delete command
func cleanupApp(t *testing.T, appKey string, serverDetails *coreConfig.ServerDetails) {
	t.Helper()

	ctx := createTestContext([]string{appKey}, map[string]string{
		"url":          serverDetails.Url,
		"access-token": serverDetails.AccessToken,
	})

	appCtx := createAppContext()
	cmd := appCmd.GetDeleteAppCommand(appCtx)

	err := cmd.Action(ctx)
	if err != nil {
		// Log but don't fail test if cleanup fails
		t.Logf("Warning: Failed to cleanup application %s: %v", appKey, err)
	}
}

// generateUniqueAppKey generates a unique app key for testing
func generateUniqueAppKey(t *testing.T, prefix string) string {
	// Use timestamp and PID for uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d-%d", prefix, timestamp, os.Getpid())
}

// getTestSpecPath returns the path to a test spec file
func getTestSpecPath(filename string) string {
	// Try relative path first (from test directory)
	relativePath := filepath.Join("testfiles", filename)
	if _, err := os.Stat(relativePath); err == nil {
		return relativePath
	}

	// Try absolute path from command testfiles
	cmdTestfilesPath := filepath.Join("apptrust", "commands", "application", "testfiles", filename)
	if _, err := os.Stat(cmdTestfilesPath); err == nil {
		return cmdTestfilesPath
	}

	// Fallback to relative from test directory
	return relativePath
}
