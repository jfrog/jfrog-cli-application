//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateApp(t *testing.T) {
	projectKey := GetTestProjectKey(t)
	appKey := GenerateUniqueKey("app-create")
	appName := "Full Test Application"
	description := "Application with all fields populated"
	businessCriticality := "critical"
	maturityLevel := "production"
	userOwners := []string{"admin", "developer"}
	groupOwners := []string{"devops-team", "security-team"}

	err := AppTrustCli.Exec("ac", appKey,
		"--project="+projectKey,
		"--application-name="+appName,
		"--desc="+description,
		"--business-criticality="+businessCriticality,
		"--maturity-level="+maturityLevel,
		"--labels=env=prod;team=devops",
		"--user-owners="+strings.Join(userOwners, ";"),
		"--group-owners="+strings.Join(groupOwners, ";"))
	assert.NoError(t, err)

	// Fetch and verify the application was created correctly
	app, _, err := GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, appName, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
	assert.Equal(t, description, *app.Description)
	assert.Equal(t, businessCriticality, *app.BusinessCriticality)
	assert.Equal(t, maturityLevel, *app.MaturityLevel)
	assert.Equal(t, map[string]string{"env": "prod", "team": "devops"}, *app.Labels)
	assert.Equal(t, userOwners, *app.UserOwners)
	assert.Equal(t, groupOwners, *app.GroupOwners)

	DeleteApplication(t, appKey)
}

func TestUpdateApp(t *testing.T) {
	projectKey := GetTestProjectKey(t)
	appKey := GenerateUniqueKey("app-update")

	CreateBasicApplication(t, appKey)

	// Update the application with new values
	updatedAppName := "Updated Test Application"
	updatedDescription := "Updated description"
	updatedBusinessCriticality := "high"
	updatedMaturityLevel := "production"
	updatedUserOwners := []string{"app-admin", "frog"}
	updatedGroupOwners := []string{"dev-team", "security-team"}

	err := AppTrustCli.Exec("au", appKey,
		"--application-name="+updatedAppName,
		"--desc="+updatedDescription,
		"--business-criticality="+updatedBusinessCriticality,
		"--maturity-level="+updatedMaturityLevel,
		"--labels=env=qa;team=dev",
		"--user-owners="+strings.Join(updatedUserOwners, ";"),
		"--group-owners="+strings.Join(updatedGroupOwners, ";"))
	assert.NoError(t, err)

	// Fetch and verify the application was updated correctly
	app, _, err := GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, updatedAppName, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
	assert.Equal(t, updatedDescription, *app.Description)
	assert.Equal(t, updatedBusinessCriticality, *app.BusinessCriticality)
	assert.Equal(t, updatedMaturityLevel, *app.MaturityLevel)
	assert.Equal(t, map[string]string{"env": "qa", "team": "dev"}, *app.Labels)
	assert.Equal(t, updatedUserOwners, *app.UserOwners)
	assert.Equal(t, updatedGroupOwners, *app.GroupOwners)

	DeleteApplication(t, appKey)
}

func TestDeleteApp(t *testing.T) {
	appKey := GenerateUniqueKey("app-delete")
	CreateBasicApplication(t, appKey)

	// Verify the application exists
	app, _, err := GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)

	// Delete the application
	err = AppTrustCli.Exec("ad", appKey)
	assert.NoError(t, err)

	// Verify the application no longer exists
	_, statusCode, err := GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)
}
