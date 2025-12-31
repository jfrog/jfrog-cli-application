//go:build e2e

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type PackagesResponse struct {
	Packages []packageBinding `json:"packages"`
}

type packageBinding struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	NumVersions   int    `json:"num_versions"`
	LatestVersion string `json:"latest_version"`
}

func GetPackageBindings(appKey string) (*PackagesResponse, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s/packages", appKey)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint, nil)
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var packagesRes *PackagesResponse
	err = json.Unmarshal(responseBody, &packagesRes)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return packagesRes, statusCode, nil
}
