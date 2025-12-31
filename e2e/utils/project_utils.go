//go:build e2e

package utils

import (
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func createTestProject(t *testing.T) {
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	projectKey := GenerateUniqueKey("apptrust-cli-tests")
	projectParams := accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: projectKey,
			ProjectKey:  projectKey,
		},
	}
	err = accessManager.CreateProject(projectParams)
	assert.NoError(t, err)
	testProjectKey = projectKey
}

func DeleteTestProject() {
	if testProjectKey == "" {
		return
	}
	deleteBuild()
	deleteNpmRepo()
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	if err != nil {
		log.Error("Failed to create Access service manager", err)
	}
	err = accessManager.DeleteProject(testProjectKey)
	if err != nil {
		log.Error("Failed to delete project", err)
	}
}
