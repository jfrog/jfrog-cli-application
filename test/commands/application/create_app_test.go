//go:build itest

package application

import (
	"testing"

	"github.com/jfrog/jfrog-cli-application/test"
	"github.com/stretchr/testify/assert"
)

func TestCreateApp(t *testing.T) {
	appKey := test.GenerateUniqueAppKey(t, "test-app-minimal")
	err := test.PlatformCli.Exec("ac", appKey)
	assert.NoError(t, err)
}
