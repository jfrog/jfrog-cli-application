//go:build e2e

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/stretchr/testify/assert"
)

func CreateBasicApplication(t *testing.T, appKey string) {
	projectKey := GetTestProjectKey(t)
	err := AppTrustCli.Exec("app-create", appKey, "--project="+projectKey, "--application-name="+appKey)
	assert.NoError(t, err)
}

func DeleteApplication(t *testing.T, appKey string) {
	err := AppTrustCli.Exec("app-delete", appKey)
	assert.NoError(t, err)
}

func GetApplication(appKey string) (*model.AppDescriptor, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s", appKey)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint, nil)
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var appDescriptor model.AppDescriptor
	err = json.Unmarshal(responseBody, &appDescriptor)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return &appDescriptor, statusCode, nil
}
