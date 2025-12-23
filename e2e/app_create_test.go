//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateApp(t *testing.T) {
	projectKey := GetTestProjectKey(t)
	appKey := GenerateUniqueKey("app-create")
	err := AppTrustCli.Exec("ac", appKey, "--project="+projectKey)
	assert.NoError(t, err)
	DeleteApplication(t, appKey)
}
