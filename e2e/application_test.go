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
	app := GetApplication(t, appKey)
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
