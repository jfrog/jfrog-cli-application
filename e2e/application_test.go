//go:build e2e

package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateApp(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)
	appKey := utils.GenerateUniqueKey("app-create")
	appName := "Full Test Application"
	description := "Application with all fields populated"
	businessCriticality := "critical"
	maturityLevel := "production"
	userOwners := []string{"admin", "developer"}
	groupOwners := []string{"devops-team", "security-team"}
	monitorPolicyValue := 5

	err := utils.AppTrustCli.Exec("app-create", appKey,
		"--project="+projectKey,
		"--application-name="+appName,
		"--desc="+description,
		"--business-criticality="+businessCriticality,
		"--maturity-level="+maturityLevel,
		"--labels=env=prod;team=devops",
		"--user-owners="+strings.Join(userOwners, ";"),
		"--group-owners="+strings.Join(groupOwners, ";"),
		"--monitor-policy=type=version_count, value=5")
	assert.NoError(t, err)

	// Fetch and verify the application was created correctly
	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, appName, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
	assert.Equal(t, description, *app.Description)
	assert.Equal(t, businessCriticality, *app.BusinessCriticality)
	assert.Equal(t, maturityLevel, *app.MaturityLevel)
	assert.ElementsMatch(t, []model.LabelEntry{{Key: "env", Value: "prod"}, {Key: "team", Value: "devops"}}, *app.Labels)
	assert.Equal(t, userOwners, *app.UserOwners)
	assert.Equal(t, groupOwners, *app.GroupOwners)
	assert.Equal(t, &model.MonitorPolicy{Type: model.MonitorPolicyTypeVersionCount, Value: &monitorPolicyValue}, app.MonitorPolicy)

	utils.DeleteApplication(t, appKey)
}

func TestUpdateApp(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)
	appKey := utils.GenerateUniqueKey("app-update")

	utils.CreateBasicApplication(t, appKey)

	updatedAppName := "Updated Test Application"
	updatedDescription := "Updated description"
	updatedBusinessCriticality := "high"
	updatedMaturityLevel := "production"
	updatedUserOwners := []string{"app-admin", "frog"}
	updatedGroupOwners := []string{"dev-team", "security-team"}
	monitorPolicyValue := 6

	err := utils.AppTrustCli.Exec("app-update", appKey,
		"--application-name="+updatedAppName,
		"--desc="+updatedDescription,
		"--business-criticality="+updatedBusinessCriticality,
		"--maturity-level="+updatedMaturityLevel,
		"--labels=env=qa;team=dev",
		"--user-owners="+strings.Join(updatedUserOwners, ";"),
		"--group-owners="+strings.Join(updatedGroupOwners, ";"),
		"--monitor-policy=type=time_frame_in_months, value=6")
	assert.NoError(t, err)

	// Fetch and verify the application was updated correctly
	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, updatedAppName, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
	assert.Equal(t, updatedDescription, *app.Description)
	assert.Equal(t, updatedBusinessCriticality, *app.BusinessCriticality)
	assert.Equal(t, updatedMaturityLevel, *app.MaturityLevel)
	assert.ElementsMatch(t, []model.LabelEntry{{Key: "env", Value: "qa"}, {Key: "team", Value: "dev"}}, *app.Labels)
	assert.Equal(t, updatedUserOwners, *app.UserOwners)
	assert.Equal(t, updatedGroupOwners, *app.GroupOwners)
	assert.Equal(t, &model.MonitorPolicy{Type: model.MonitorPolicyTypeTimeframe, Value: &monitorPolicyValue}, app.MonitorPolicy)

	utils.DeleteApplication(t, appKey)
}

func TestDeleteApp(t *testing.T) {
	appKey := utils.GenerateUniqueKey("app-delete")
	utils.CreateBasicApplication(t, appKey)

	// Verify the application exists
	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)

	// Delete the application
	err = utils.AppTrustCli.Exec("app-delete", appKey)
	assert.NoError(t, err)

	// Verify the application no longer exists
	_, statusCode, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)
}

func TestExportImportApp(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)
	appKey := utils.GenerateUniqueKey("app-export")
	utils.CreateBasicApplication(t, appKey)

	exportPath := filepath.Join(t.TempDir(), "export.json")
	err := utils.AppTrustCli.Exec("app-export", appKey, exportPath)
	assert.NoError(t, err)

	err = utils.AppTrustCli.Exec("app-delete", appKey)
	assert.NoError(t, err)

	_, statusCode, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)

	err = utils.AppTrustCli.Exec("app-import", exportPath)
	assert.NoError(t, err)
	defer utils.DeleteApplication(t, appKey)

	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, appKey, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
}
