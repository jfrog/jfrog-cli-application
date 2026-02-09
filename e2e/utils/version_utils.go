//go:build e2e

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type VersionContentResponse struct {
	ApplicationKey string       `json:"application_key"`
	Version        string       `json:"version"`
	Status         string       `json:"status"`
	CurrentStage   string       `json:"current_stage,omitempty"`
	Tag            string       `json:"tag,omitempty"`
	Releasables    []releasable `json:"releasables"`
}

type releasable struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	PackageType string     `json:"package_type"`
	Artifacts   []artifact `json:"artifacts,omitempty"`
}

type artifact struct {
	Path string `json:"path"`
}

const (
	StatusDraft      = "DRAFT"
	StatusInProgress = "IN_PROGRESS"
	StatusStarted    = "STARTED"
	StatusCompleted  = "COMPLETED"
)

func GetApplicationVersion(appKey, version string) (*VersionContentResponse, int, error) {
	statusCode := 0
	ctx, err := service.NewContext(*serverDetails)
	if err != nil {
		return nil, statusCode, err
	}

	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/content", appKey, version)
	response, responseBody, err := ctx.GetHttpClient().Get(endpoint, map[string]string{"include": "releasables_expanded"})
	if response != nil {
		statusCode = response.StatusCode
	}
	if err != nil || statusCode != http.StatusOK {
		return nil, statusCode, err
	}

	var versionRes *VersionContentResponse
	err = json.Unmarshal(responseBody, &versionRes)
	if err != nil {
		return nil, statusCode, errorutils.CheckError(err)
	}

	return versionRes, statusCode, nil
}

// WaitForVersionCreation polls GetApplicationVersion until the version status is COMPLETED or the timeout is reached.
// This is useful before deleting an asynchronously created version.
func WaitForVersionCreation(t *testing.T, appKey, version string, timeout, interval time.Duration) {
	require.Eventually(t, func() bool {
		vc, sc, err := GetApplicationVersion(appKey, version)
		return err == nil && sc == http.StatusOK && vc != nil && vc.Status == StatusCompleted
	}, timeout, interval, "version %s/%s did not reach COMPLETED status within %s", appKey, version, timeout)
}

func DeleteApplicationVersion(t *testing.T, appKey, version string) {
	err := AppTrustCli.Exec("version-delete", appKey, version)
	assert.NoError(t, err)
}
